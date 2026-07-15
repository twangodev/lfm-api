package lfm_api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	httpClient "github.com/bozd4g/go-http-client"
)

// Last.fm's Varnish layer intermittently serves HTTP 600 "Temporarily Unavailable"
func TestGetActiveScrobbleNon200ReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(600)
		_, _ = w.Write([]byte("<title>Last.fm - Temporarily Unavailable</title>"))
	}))
	defer server.Close()

	original := lastFm
	lastFm = httpClient.New(server.URL + "/")
	defer func() { lastFm = original }()

	scrobble, err := GetActiveScrobble("test")
	if err == nil {
		t.Fatal("expected error for non-200 response, got nil")
	}
	if scrobble.Active {
		t.Fatal("expected inactive scrobble on non-200 response")
	}
}
