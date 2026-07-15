package lfm_api_test

import (
	"strings"
	"testing"

	"github.com/twangodev/lfm-api" // Adjust the import path according to your module name
)

func TestGetActiveScrobble(t *testing.T) {
	scrobble, err := lfm_api.GetActiveScrobble("rj")
	if err != nil {
		// Live integration test; last.fm intermittently serves non-200 responses
		if strings.Contains(err.Error(), "last.fm returned status") {
			t.Skipf("last.fm unavailable: %v", err)
		}
		t.Fatalf("unexpected error: %v", err)
	}
	t.Log(scrobble)
}
