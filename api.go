package lfm_api

import (
	"fmt"
	httpClient "github.com/bozd4g/go-http-client"
	"golang.org/x/net/context"
	"golang.org/x/net/html"
	"io"
	"strconv"
	"strings"
	"time"
)

const LastFmUrl = "https://www.last.fm/"

var lastFm = httpClient.New(LastFmUrl)
var ctx = context.Background()

// GetActiveScrobble returns the active scrobble for the given user.
func GetActiveScrobble(username string) (Scrobble, error) {
	request, err := lastFm.Get(ctx, fmt.Sprintf("user/%v/partial/recenttracks?ajax=1", username))
	if err != nil { // Error would request formed
		return EmptyScrobble, err
	}

	body := string(request.Body())
	code := request.Status()
	if code != 200 { // Request unsuccessful
		return EmptyScrobble, err
	}

	// No active scrobble detected
	if !strings.Contains(body, "Scrobbling now") {
		return EmptyScrobble, err
	}

	ioReader := strings.NewReader(strings.ReplaceAll(body, "\n", ""))
	tokenizer := html.NewTokenizer(ioReader)

	var keys []string
	var name string
	var artist string
	var album string
	var loved bool
	var dataId string
	var dataTime time.Time
	var dataLink string
	var dataLinkTitle string
	var coverArtUrl string

	for {
		tokenType := tokenizer.Next()

		// Error tokenizing HTML
		if tokenType == html.ErrorToken {
			err = tokenizer.Err()
			if err == io.EOF {
				break
			}
		}

		if tokenType == html.StartTagToken {
			token := tokenizer.Token()

			if "thead" == token.Data { // Get keys for the map
				tokenizer.Next() // Table row
				for {            // Each header cell
					tokenType = tokenizer.Next()
					if tokenType == html.StartTagToken && "th" == tokenizer.Token().Data { // Each header
						tokenizer.Next()
						key := tokenizer.Token().Data
						trimmed := strings.TrimSpace(key)
						keys = append(keys, trimmed)
					}
					if tokenType == html.EndTagToken && "thead" == tokenizer.Token().Data { // At the end of thead
						break
					}
				}
			}

			if "tbody" == token.Data { // Get values for map
				// Get Table row representing the latest scrobble
				for {
					tokenType = tokenizer.Next()
					token = tokenizer.Token()
					if tokenType == html.StartTagToken && "tr" == token.Data {
						break
					}
				}

				attributes := token.Attr
				dataId = searchHTMLAttribute(attributes, "data-recenttrack-id")
				dataTimeString := searchHTMLAttribute(attributes, "data-timestamp")
				dataTimeInt64, err := strconv.ParseInt(dataTimeString, 10, 64)
				if err != nil {
					return EmptyScrobble, err
				}
				dataTime = time.Unix(dataTimeInt64, 0)
				index := 0

				for {

					tokenizerCopy := tokenizer

					if index == len(keys)-1 {
						break
					}

					tokenType = tokenizerCopy.Next()
					token = tokenizerCopy.Token()
					if tokenType == html.StartTagToken && "td" == token.Data {
						currentKey := keys[index]
						if currentKey == "Play" {
							for {
								tokenType = tokenizerCopy.Next()
								token = tokenizerCopy.Token()
								if tokenType == html.StartTagToken && "a" == token.Data {
									attributes = token.Attr
									dataLink = searchHTMLAttribute(attributes, "href")
									dataLinkTitle = searchHTMLAttribute(attributes, "title")
									break
								}

								if tokenType == html.EndTagToken && "td" == token.Data {
									dataLink = ""
									dataLinkTitle = ""
									break
								}
							}
						} else if currentKey == "Album" {
							for {
								tokenType = tokenizerCopy.Next()
								token = tokenizerCopy.Token()
								if tokenType == html.SelfClosingTagToken && "img" == token.Data {
									attributes = token.Attr
									album = html.UnescapeString(searchHTMLAttribute(attributes, "alt"))
									coverArtUrl = searchHTMLAttribute(attributes, "src")
									break
								}

								if tokenType == html.EndTagToken && "td" == token.Data {
									album = ""
									coverArtUrl = "lfm_logo"
									break
								}
							}
						} else if currentKey == "Loved" {
							for {
								tokenType = tokenizerCopy.Next()
								token = tokenizerCopy.Token()
								if tokenType == html.StartTagToken && "div" == token.Data {
									attributes = token.Attr
									lovedStringState := searchHTMLAttribute(attributes, "data-toggle-button-current-state")
									if lovedStringState == "loved" {
										loved = true
									} else {
										loved = false
									}
									break
								}

								if tokenType == html.EndTagToken && "td" == token.Data {
									loved = false
									break
								}
							}
						} else if currentKey == "Track name" {
							for {
								tokenType = tokenizerCopy.Next()
								token = tokenizerCopy.Token()
								if tokenType == html.StartTagToken && "a" == token.Data {
									// Get text token, which is after the start tag token
									tokenizerCopy.Next()
									token = tokenizerCopy.Token()
									name = html.UnescapeString(token.Data)
									break
								}

								if tokenType == html.EndTagToken && "td" == token.Data {
									name = "Unknown Song"
									break
								}
							}
						} else if currentKey == "Artist name" {
							for {
								tokenType = tokenizerCopy.Next()
								token = tokenizerCopy.Token()
								if tokenType == html.StartTagToken && "a" == token.Data {
									// Get text token, which is after the start tag token
									tokenizerCopy.Next()
									token = tokenizerCopy.Token()
									artist = html.UnescapeString(token.Data)
									break
								}

								if tokenType == html.EndTagToken && "td" == token.Data {
									artist = "Unknown Song"
									break
								}
							}
						}
						index++
					}
				}
			}
		}
	}

	return Scrobble{
		Active:        true,
		Name:          name,
		Artist:        artist,
		Album:         album,
		Loved:         loved,
		DataId:        dataId,
		DataTimestamp: dataTime,
		DataLink:      dataLink,
		DataLinkTitle: dataLinkTitle,
		CoverArtUrl:   coverArtUrl,
	}, nil

}
