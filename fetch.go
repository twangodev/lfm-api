package lfm_api

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/net/publicsuffix"
)

const maxResponseBytes = 2 << 20

// A client keeps verification cookies across polls. Serialize handshakes so two
// concurrent callers cannot overwrite each other's challenge-start cookie.
type lastFMClient struct {
	base string
	http *http.Client
	mu   sync.Mutex
}

func newLastFMClient(base string) *lastFMClient {
	jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	origin, _ := url.Parse(base)
	return &lastFMClient{base: base, http: &http.Client{
		Jar: jar, Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 || req.URL.Scheme != origin.Scheme || req.URL.Host != origin.Host {
				return fmt.Errorf("last.fm returned an unexpected redirect")
			}
			return nil
		},
	}}
}

func (c *lastFMClient) request(ctx context.Context, method, target string, payload []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, target, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept-Language", "en")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch last.fm: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("last.fm returned status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read last.fm response: %w", err)
	}
	if len(body) > maxResponseBytes {
		return nil, fmt.Errorf("last.fm response exceeds size limit")
	}
	return body, nil
}

func (c *lastFMClient) recentTracks(username string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	target := c.base + "user/" + url.PathEscape(username) + "/partial/recenttracks?ajax=1"
	body, err := c.request(ctx, http.MethodGet, target, nil)
	if err != nil {
		return "", err
	}
	if isChallenge(body) {
		if err := c.verify(ctx, body); err != nil {
			return "", err
		}
		body, err = c.request(ctx, http.MethodGet, target, nil)
		if err != nil {
			return "", err
		}
		if isChallenge(body) {
			return "", fmt.Errorf("last.fm challenge persisted after verification")
		}
	}
	if err := validateRecentTracks(body); err != nil {
		return "", err
	}
	return string(body), nil
}

func isChallenge(body []byte) bool {
	return bytes.Contains(body, []byte("/_fs-ch-")) || bytes.Contains(body, []byte("Client Challenge"))
}

var scriptAssignment = regexp.MustCompile(`\bsrc\s*=\s*["']([^"']+)["']`)
var challengeScriptPath = regexp.MustCompile(`^/_fs-ch-[A-Za-z0-9_-]+/script\.js$`)
var initCall = regexp.MustCompile(`(?:^|[;\n])\s*(?:window\.)?init\s*\(`)

type powData struct {
	Base    string `json:"base"`
	Hash    string `json:"hash"`
	HMAC    string `json:"hmac"`
	Expires string `json:"expires"`
}
type challenge struct {
	Type string  `json:"ty"`
	Data powData `json:"data"`
}
type challengeAnswer struct {
	Type    string `json:"ty"`
	Base    string `json:"base"`
	Answer  string `json:"answer"`
	HMAC    string `json:"hmac"`
	Expires string `json:"expires"`
}

// Parse only the JSON arguments of the bootstrap call; never execute downloaded
// JavaScript. The currently supported puzzle is SHA256(base + two base62 chars).
func parseChallenge(script []byte) ([]challenge, string, string, error) {
	matches := initCall.FindAllIndex(script, -1)
	if len(matches) == 0 {
		return nil, "", "", fmt.Errorf("last.fm challenge bootstrap not recognized")
	}
	rest := bytes.TrimSpace(script[matches[len(matches)-1][1]:])
	var tasks []challenge
	var token, prefix string
	for i, dst := range []interface{}{&tasks, &token, &prefix} {
		if i > 0 {
			if len(rest) == 0 || rest[0] != ',' {
				return nil, "", "", fmt.Errorf("invalid last.fm challenge arguments")
			}
			rest = bytes.TrimSpace(rest[1:])
		}
		dec := json.NewDecoder(bytes.NewReader(rest))
		if err := dec.Decode(dst); err != nil {
			return nil, "", "", fmt.Errorf("decode last.fm challenge: %w", err)
		}
		rest = bytes.TrimSpace(rest[dec.InputOffset():])
	}
	if token == "" || len(tasks) == 0 || len(tasks) > 4 {
		return nil, "", "", fmt.Errorf("invalid last.fm challenge parameters")
	}
	return tasks, token, prefix, nil
}

func solvePOW(ctx context.Context, task challenge) (challengeAnswer, error) {
	if task.Type != "pow" {
		return challengeAnswer{}, fmt.Errorf("unsupported last.fm challenge type %q", task.Type)
	}
	d := task.Data
	hash, err := hex.DecodeString(d.Hash)
	if err != nil || len(hash) != sha256.Size || d.Base == "" || len(d.Base) > 256 || d.HMAC == "" || len(d.HMAC) > 256 || d.Expires == "" {
		return challengeAnswer{}, fmt.Errorf("invalid last.fm proof-of-work parameters")
	}
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for _, a := range alphabet {
		if err := ctx.Err(); err != nil {
			return challengeAnswer{}, err
		}
		for _, b := range alphabet {
			answer := string([]rune{a, b})
			sum := sha256.Sum256([]byte(d.Base + answer))
			if bytes.Equal(sum[:], hash) {
				return challengeAnswer{"pow", d.Base, answer, d.HMAC, d.Expires}, nil
			}
		}
	}
	return challengeAnswer{}, fmt.Errorf("unsupported last.fm proof-of-work puzzle")
}

func (c *lastFMClient) verify(ctx context.Context, page []byte) error {
	// Both external script tags and Last.fm's inline script.src assignment occur.
	var candidates []string
	z := html.NewTokenizer(bytes.NewReader(page))
	for z.Next() != html.ErrorToken {
		t := z.Token()
		if t.Type == html.StartTagToken && t.Data == "script" {
			candidates = append(candidates, searchHTMLAttribute(t.Attr, "src"))
		}
		if t.Type == html.TextToken {
			for _, m := range scriptAssignment.FindAllStringSubmatch(t.Data, -1) {
				candidates = append(candidates, m[1])
			}
		}
	}
	var scriptURL *url.URL
	for _, candidate := range candidates {
		u, err := url.Parse(candidate)
		if err == nil && u.Scheme == "" && u.Host == "" && u.User == nil && u.Fragment == "" && challengeScriptPath.MatchString(u.Path) {
			scriptURL = u
			break
		}
	}
	if scriptURL == nil {
		return fmt.Errorf("last.fm challenge script not recognized")
	}
	origin, _ := url.Parse(c.base)
	script, err := c.request(ctx, http.MethodGet, origin.ResolveReference(scriptURL).String(), nil)
	if err != nil {
		return err
	}
	tasks, token, prefix, err := parseChallenge(script)
	if err != nil {
		return err
	}
	if prefix != path.Dir(scriptURL.Path) {
		return fmt.Errorf("unexpected last.fm challenge endpoint")
	}
	endpoint := origin.ResolveReference(&url.URL{Path: prefix + "/fst-post-back"}).String()
	for round := 0; round < 2; round++ {
		answers := make([]challengeAnswer, 0, len(tasks))
		if len(tasks) == 0 || len(tasks) > 4 || token == "" {
			return fmt.Errorf("invalid last.fm challenge parameters")
		}
		for _, task := range tasks {
			answer, err := solvePOW(ctx, task)
			if err != nil {
				return err
			}
			answers = append(answers, answer)
		}
		payload, err := json.Marshal(struct {
			Token string            `json:"token"`
			Data  []challengeAnswer `json:"data"`
		}{token, answers})
		if err != nil {
			return err
		}
		body, err := c.request(ctx, http.MethodPost, endpoint, payload)
		if err != nil {
			return err
		}
		var result struct {
			Status     string      `json:"status"`
			Challenges []challenge `json:"ch"`
			Token      string      `json:"tok"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			return fmt.Errorf("decode last.fm verification response: %w", err)
		}
		if result.Status == "success" {
			return nil
		}
		if len(result.Challenges) == 0 {
			return fmt.Errorf("last.fm challenge verification rejected")
		}
		tasks, token = result.Challenges, result.Token
	}
	return fmt.Errorf("last.fm challenge retry limit exceeded")
}

func validateRecentTracks(body []byte) error {
	// Require explicit closing tags too: html.Parse repairs truncated documents.
	z := html.NewTokenizer(bytes.NewReader(body))
	var closedTable, closedBody bool
	for z.Next() != html.ErrorToken {
		t := z.Token()
		if t.Type == html.EndTagToken {
			closedTable = closedTable || t.Data == "table"
			closedBody = closedBody || t.Data == "tbody"
		}
	}
	if !closedTable || !closedBody {
		return fmt.Errorf("incomplete last.fm recent-tracks response")
	}
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("parse last.fm response: %w", err)
	}
	var table, tbody bool
	var walk func(*html.Node, bool)
	walk = func(n *html.Node, inTable bool) {
		if n.Type == html.ElementNode && n.Data == "table" && hasClass(n, "chartlist") {
			table = true
			inTable = true
		}
		if inTable && n.Type == html.ElementNode && n.Data == "tbody" {
			tbody = true
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child, inTable)
		}
	}
	walk(doc, false)
	if !table || !tbody {
		return fmt.Errorf("unexpected last.fm recent-tracks response")
	}
	return nil
}

func hasClass(n *html.Node, class string) bool {
	for _, value := range strings.Fields(searchHTMLAttribute(n.Attr, "class")) {
		if value == class {
			return true
		}
	}
	return false
}
