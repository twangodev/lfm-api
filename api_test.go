package lfm_api_test

import (
	"os"
	"testing"

	"github.com/twangodev/lfm-api" // Adjust the import path according to your module name
)

func TestGetActiveScrobble(t *testing.T) {
	username := os.Getenv("LFM_LIVE_TEST_USER")
	if username == "" {
		t.Skip("set LFM_LIVE_TEST_USER to run the Last.fm integration test")
	}
	scrobble, err := lfm_api.GetActiveScrobble(username)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("active=%v %v", scrobble.Active, scrobble)
}
