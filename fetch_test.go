package lfm_api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

const activeTracks = `<table class="chartlist"><thead><tr><th>Play</th><th>Album</th><th>Loved</th><th>Track name</th><th>Artist name</th><th>Timestamp</th></tr></thead><tbody><tr class="chartlist-row chartlist-row--now-scrobbling" data-recenttrack-id="track-id" data-timestamp="1700000000"><td><a href="https://example.com/play" title="Play track">Play</a></td><td><img src="https://example.com/cover.jpg" alt="Album &amp; One" /></td><td><div data-toggle-button-current-state="loved"></div></td><td><a>Song &amp; One</a></td><td><a>Artist</a></td><td>Scrobbling now</td></tr></tbody></table>`
const challengePage = `<title>Client Challenge</title><script>script.src = '/_fs-ch-test/script.js?reload=true';</script>`

func testTask() challenge {
	sum := sha256.Sum256([]byte("test-base" + "Z9"))
	return challenge{"pow", powData{"test-base", hex.EncodeToString(sum[:]), "signed-proof", "1700000000"}}
}
func bootstrap(tasks []challenge, prefix string) string {
	data, _ := json.Marshal(tasks)
	return fmt.Sprintf(`"use strict";init(%s, "test-token", %q, true);`, data, prefix)
}

func TestChallengeCookieReuse(t *testing.T) {
	var scripts, posts, tracks int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/user/test/partial/recenttracks":
			tracks++
			if _, err := r.Cookie("verified"); err == nil {
				fmt.Fprint(w, activeTracks)
				return
			}
			http.SetCookie(w, &http.Cookie{Name: "challenge-start", Value: "start", Path: "/"})
			fmt.Fprint(w, challengePage)
		case "/_fs-ch-test/script.js":
			scripts++
			if _, err := r.Cookie("challenge-start"); err != nil {
				t.Error("start cookie missing on script request")
			}
			fmt.Fprint(w, bootstrap([]challenge{testTask()}, "/_fs-ch-test"))
		case "/_fs-ch-test/fst-post-back":
			posts++
			if r.Method != http.MethodPost {
				t.Error("expected POST")
			}
			if _, err := r.Cookie("challenge-start"); err != nil {
				t.Error("start cookie missing on verification")
			}
			var payload struct {
				Token string            `json:"token"`
				Data  []challengeAnswer `json:"data"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Error(err)
			}
			if payload.Token != "test-token" || len(payload.Data) != 1 {
				t.Error("invalid verification payload")
				w.WriteHeader(400)
				return
			}
			answer := payload.Data[0]
			if answer.Answer != "Z9" || answer.Base != "test-base" || answer.HMAC != "signed-proof" || answer.Expires != "1700000000" || answer.Type != "pow" {
				t.Errorf("incorrect proof: %+v", answer)
			}
			http.SetCookie(w, &http.Cookie{Name: "verified", Value: "yes", Path: "/", MaxAge: 3600})
			fmt.Fprint(w, `{"status":"success"}`)
		default:
			t.Errorf("unexpected request %s", r.URL)
			w.WriteHeader(404)
		}
	}))
	defer server.Close()
	c := newLastFMClient(server.URL + "/")
	// Concurrent callers also share one verification, with a race-safe cookie jar.
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			body, err := c.recentTracks("test")
			if err != nil {
				t.Error(err)
				return
			}
			s, err := parseActiveScrobble(body)
			if err != nil || !s.Active || s.Name != "Song & One" || s.Artist != "Artist" || s.Album != "Album & One" || !s.Loved || s.DataId != "track-id" || s.DataTimestamp.Unix() != 1700000000 || s.DataLink != "https://example.com/play" || s.CoverArtUrl != "https://example.com/cover.jpg" {
				t.Errorf("incorrect scrobble: %+v, %v", s, err)
			}
		}()
	}
	wg.Wait()
	if scripts != 1 || posts != 1 || tracks != 4 {
		t.Fatalf("scripts=%d posts=%d tracks=%d", scripts, posts, tracks)
	}
	// Delete verification to model expiry: the next poll must perform a fresh handshake.
	u, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	c.http.Jar.SetCookies(u.URL, []*http.Cookie{{Name: "verified", Path: "/", MaxAge: -1}})
	if _, err := c.recentTracks("test"); err != nil {
		t.Fatal(err)
	}
	if scripts != 2 || posts != 2 {
		t.Fatalf("cookie was not renewed: scripts=%d posts=%d", scripts, posts)
	}
}

func TestChallengeFailures(t *testing.T) {
	unsupported := testTask()
	unsupported.Type = "captcha"
	for _, tc := range []struct {
		name, page, script, result, want string
		posts                            int
	}{
		{"unsupported", challengePage, bootstrap([]challenge{unsupported}, "/_fs-ch-test"), "", "unsupported", 0},
		{"malformed script", challengePage, "not javascript", "", "bootstrap", 0},
		{"foreign endpoint", challengePage, bootstrap([]challenge{testTask()}, "https://example.com"), "", "endpoint", 0},
		{"foreign script", `<title>Client Challenge</title><script src="https://example.com/_fs-ch-test/script.js"></script>`, "", "", "script", 0},
		{"rejected", challengePage, bootstrap([]challenge{testTask()}, "/_fs-ch-test"), `{"status":"error"}`, "rejected", 1},
		{"malformed result", challengePage, bootstrap([]challenge{testTask()}, "/_fs-ch-test"), `not json`, "verification response", 1},
		{"persistent challenge", challengePage, bootstrap([]challenge{testTask()}, "/_fs-ch-test"), `{"status":"success"}`, "persisted", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			posts := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/_fs-ch-test/script.js":
					fmt.Fprint(w, tc.script)
				case "/_fs-ch-test/fst-post-back":
					posts++
					fmt.Fprint(w, tc.result)
				default:
					fmt.Fprint(w, tc.page)
				}
			}))
			defer server.Close()
			_, err := newLastFMClient(server.URL + "/").recentTracks("test")
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("got %v, want %q", err, tc.want)
			}
			if posts != tc.posts {
				t.Fatalf("posts=%d, want %d", posts, tc.posts)
			}
		})
	}
}

func TestChallengeRetryLimit(t *testing.T) {
	posts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/_fs-ch-test/script.js":
			fmt.Fprint(w, bootstrap([]challenge{testTask()}, "/_fs-ch-test"))
		case "/_fs-ch-test/fst-post-back":
			posts++
			json.NewEncoder(w).Encode(map[string]interface{}{"ch": []challenge{testTask()}, "tok": "new-token"})
		default:
			fmt.Fprint(w, challengePage)
		}
	}))
	defer server.Close()
	_, err := newLastFMClient(server.URL + "/").recentTracks("test")
	if err == nil || !strings.Contains(err.Error(), "retry limit") || posts != 2 {
		t.Fatalf("posts=%d err=%v", posts, err)
	}
}

func TestResponseClassification(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		wantError  bool
	}{
		{"idle", strings.ReplaceAll(activeTracks, "Scrobbling now", "Yesterday"), false},
		{"empty table", `<table class="chartlist"><tbody></tbody></table>`, false},
		{"maintenance", `<html><title>Temporarily unavailable</title></html>`, true},
		{"login", `<html><form>Sign in</form></html>`, true},
		{"empty body", "", true},
		{"truncated idle table", `<table class="chartlist"><tbody>`, true},
		{"oversized", strings.Repeat("x", maxResponseBytes+1), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, tc.body) }))
			defer server.Close()
			body, err := newLastFMClient(server.URL + "/").recentTracks("test")
			if (err != nil) != tc.wantError {
				t.Fatalf("err=%v", err)
			}
			if err == nil {
				s, err := parseActiveScrobble(body)
				if err != nil || s.Active {
					t.Fatalf("expected idle, got %+v %v", s, err)
				}
			}
		})
	}
}

func TestMalformedActiveTrackTerminates(t *testing.T) {
	for _, body := range []string{
		`Scrobbling now<thead><tr><th>Track name`,
		`Scrobbling now<tbody>`,
		`Scrobbling now<thead><tr><th>Track name</th></tr></thead><tbody><tr data-timestamp="1700000000"><td>`,
		`Scrobbling now<table></table>`,
	} {
		if _, err := parseActiveScrobble(body); err == nil {
			t.Errorf("expected error for %q", body)
		}
	}
}

func TestPOWBounds(t *testing.T) {
	task := testTask()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := solvePOW(ctx, task); err != context.Canceled {
		t.Fatalf("got %v", err)
	}
	sum := sha256.Sum256([]byte("no matching answer"))
	task.Data.Hash = hex.EncodeToString(sum[:])
	if _, err := solvePOW(context.Background(), task); err == nil {
		t.Fatal("expected unsupported puzzle error")
	}
	task.Data.Hash = "invalid"
	if _, err := solvePOW(context.Background(), task); err == nil {
		t.Fatal("expected invalid hash error")
	}
}

func TestRequestCancellationAndRedirect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "https://example.com/", 302)
			return
		}
		<-r.Context().Done()
	}))
	defer server.Close()
	c := newLastFMClient(server.URL + "/")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if _, err := c.request(ctx, http.MethodGet, server.URL, nil); err == nil {
		t.Fatal("expected timeout")
	}
	if _, err := c.request(context.Background(), http.MethodGet, server.URL+"/redirect", nil); err == nil || !strings.Contains(err.Error(), "redirect") {
		t.Fatalf("got %v", err)
	}
}
