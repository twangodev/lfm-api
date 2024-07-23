package lfm_api_test

import (
	"github.com/twangodev/lfm-api" // Adjust the import path according to your module name
	"testing"
)

func TestGetActiveScrobble(t *testing.T) {
	scrobble, err := lfm_api.GetActiveScrobble("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Log(scrobble)
}
