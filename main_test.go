package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// testAPIServerResponse holds the JSON response our test API server will return.
var testAPIServerResponse atomic.Value

// apiHandler is the HTTP handler for our test API server.
// It writes the current response (stored in testAPIServerResponse) to the response writer.
func apiHandler(w http.ResponseWriter, r *http.Request) {
	response, ok := testAPIServerResponse.Load().(string)
	if !ok {
		http.Error(w, "No response set", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(response))
}

// discordPayloads will store payloads that our test Discord server receives.
var discordPayloads []byte

// discordHandler is the HTTP handler for our test Discord server.
func discordHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusInternalServerError)
		return
	}
	// Store the payload globally for inspection.
	discordPayloads = body
	w.WriteHeader(http.StatusNoContent)
}

func TestRunStatusCheck_StatusChange(t *testing.T) {
	// Create a test API server.
	apiServer := httptest.NewServer(http.HandlerFunc(apiHandler))
	defer apiServer.Close()

	// Create a test Discord server.
	discordServer := httptest.NewServer(http.HandlerFunc(discordHandler))
	defer discordServer.Close()

	// Prepare a temporary file for storing state.
	tmpStateFile := "test_state.json"
	// Ensure cleanup.
	defer os.Remove(tmpStateFile)

	// Set initial API response (first call) for "Faerie" status.
	// For example, Faerie is Online.
	initialResponse := `[{"name": "Faerie", "status": "Online", "congestion": "Standard", "creation": "2023-01-01T00:00:00Z"}]`
	testAPIServerResponse.Store(initialResponse)

	// Use a short maxRetryDuration for testing purposes.
	maxRetryDuration := 200 * time.Millisecond
	serverName := "Faerie"
	debug := false

	// Clear any previous discord payload.
	discordPayloads = nil

	// Call RunStatusCheck once (this should record the initial state and post it).
	RunStatusCheck(apiServer.URL, discordServer.URL, serverName, tmpStateFile, maxRetryDuration, debug)

	// At this point, discordPayloads should be non-empty.
	if len(discordPayloads) == 0 {
		t.Fatal("Expected a Discord payload on first run, got none")
	}

	// Parse the posted payload to verify the initial status.
	var payload map[string]interface{}
	if err := json.Unmarshal(discordPayloads, &payload); err != nil {
		t.Fatalf("Failed to unmarshal Discord payload: %v", err)
	}
	embeds, ok := payload["embeds"].([]interface{})
	if !ok || len(embeds) == 0 {
		t.Fatal("No embeds found in Discord payload on first run")
	}
	firstEmbed, ok := embeds[0].(map[string]interface{})
	if !ok {
		t.Fatal("First embed is not a map")
	}
	desc, _ := firstEmbed["description"].(string)
	if !strings.Contains(desc, "Online") {
		t.Errorf("Expected initial status 'Online' in embed, got %s", desc)
	}

	// Now update the API response to simulate a status change.
	// For example, Faerie changes to Offline and becomes Congested.
	changedResponse := `[{"name": "Faerie", "status": "Offline", "congestion": "Congested", "creation": "2023-01-01T00:00:00Z"}]`
	testAPIServerResponse.Store(changedResponse)

	// Clear discord payload before second call.
	discordPayloads = nil

	// Call RunStatusCheck a second time.
	RunStatusCheck(apiServer.URL, discordServer.URL, serverName, tmpStateFile, maxRetryDuration, debug)

	// Now, discordPayloads should contain the new embed.
	if len(discordPayloads) == 0 {
		t.Fatal("Expected a Discord payload on second run (status change), got none")
	}

	// Parse the payload.
	var payload2 map[string]interface{}
	if err := json.Unmarshal(discordPayloads, &payload2); err != nil {
		t.Fatalf("Failed to unmarshal Discord payload on second run: %v", err)
	}
	embeds2, ok := payload2["embeds"].([]interface{})
	if !ok || len(embeds2) == 0 {
		t.Fatal("No embeds found in Discord payload on second run")
	}
	secondEmbed, ok := embeds2[0].(map[string]interface{})
	if !ok {
		t.Fatal("Second embed is not a map")
	}
	desc2, _ := secondEmbed["description"].(string)
	if !strings.Contains(desc2, "Offline") {
		t.Errorf("Expected changed status 'Offline' in embed, got %s", desc2)
	}
	// Optionally, verify the embed color is red (0xFF0000) for congested.
	colorFloat, ok := secondEmbed["color"].(float64)
	if !ok {
		t.Error("Embed color not found or not a number")
	} else {
		if int(colorFloat) != 0xFF0000 {
			t.Errorf("Expected embed color 0xFF0000 for congested status, got 0x%X", int(colorFloat))
		}
	}
}
