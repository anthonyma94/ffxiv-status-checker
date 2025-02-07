package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestRunStatusCheck_APIUnavailable_PostsErrorEmbed(t *testing.T) {
	// Create a test API server that always returns 503 Service Unavailable.
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
	}))
	defer apiServer.Close()

	// Create a test Discord server to capture the payload.
	var postedPayload []byte
	discordServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		postedPayload, err = io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Failed to read request body: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer discordServer.Close()

	// Use a very short ticker interval (e.g., 50ms) so that max retry duration becomes 5x that.
	tickerInterval := 50 * time.Millisecond
	maxRetryDuration := tickerInterval * 5

	// Use a temporary file for state.
	stateFileName := "test_state.json"
	defer os.Remove(stateFileName)
	serverName := "TestServer"
	debug := false

	// Call RunStatusCheck with the test API server and test Discord server.
	RunStatusCheck(apiServer.URL, discordServer.URL, serverName, stateFileName, maxRetryDuration, debug)

	// Check that something was posted to the Discord test server.
	if len(postedPayload) == 0 {
		t.Fatal("No payload posted to Discord")
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(postedPayload, &payload); err != nil {
		t.Fatalf("Failed to unmarshal posted payload: %v", err)
	}

	embeds, ok := payload["embeds"].([]interface{})
	if !ok || len(embeds) == 0 {
		t.Fatal("No embeds found in payload")
	}
	embed, ok := embeds[0].(map[string]interface{})
	if !ok {
		t.Fatal("Embed is not a map")
	}
	title, ok := embed["title"].(string)
	if !ok {
		t.Fatal("Embed title not found or not a string")
	}
	expectedTitle := "Error Retrieving Server Status"
	if title != expectedTitle {
		t.Errorf("Expected embed title %q, got %q", expectedTitle, title)
	}
}
