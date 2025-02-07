package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/anthonyma94/ffxiv-status-checker/discord"
)

func TestGetServersWithRetryUnavailable_PostsErrorEmbed(t *testing.T) {
	// Create a test API server that always returns 503 Service Unavailable.
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
	}))
	defer apiServer.Close()

	// Create a test Discord server to capture the webhook payload.
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

	// Use a very short ticker interval.
	tickerInterval := 50 * time.Millisecond
	// Set max retry duration to 5x the ticker interval.
	maxDuration := tickerInterval * 5

	_, err := GetServersWithRetryWithTimeout(apiServer.URL, maxDuration)
	if err == nil {
		t.Error("Expected an error when the API is unavailable, got nil")
	}

	// Simulate the error embed as in the main code.
	errorEmbed := map[string]interface{}{
		"title":       "Error Retrieving Server Status",
		"description": "Error retrieving status for **TestServer** after " + maxDuration.String() + ":\n`" + err.Error() + "`",
		"color":       0xFF0000,
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	// Post the error embed to our test Discord server.
	if err := discord.PostEmbedToDiscord(discordServer.URL, errorEmbed); err != nil {
		t.Fatalf("Failed to post embed to test Discord server: %v", err)
	}

	// Unmarshal the captured payload.
	var payload map[string]interface{}
	if err := json.Unmarshal(postedPayload, &payload); err != nil {
		t.Fatalf("Failed to unmarshal posted payload: %v", err)
	}

	// Check that the payload contains an "embeds" array with at least one embed.
	embeds, ok := payload["embeds"].([]interface{})
	if !ok || len(embeds) == 0 {
		t.Fatal("No embeds found in payload")
	}
	embed, ok := embeds[0].(map[string]interface{})
	if !ok {
		t.Fatal("Embed is not a map")
	}

	// Check that the embed's title is as expected.
	title, ok := embed["title"].(string)
	if !ok {
		t.Fatal("Embed title not found or not a string")
	}
	expectedTitle := "Error Retrieving Server Status"
	if title != expectedTitle {
		t.Errorf("Expected embed title %q, got %q", expectedTitle, title)
	}
}
