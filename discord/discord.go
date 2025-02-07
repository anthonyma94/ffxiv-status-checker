package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// PostEmbedToDiscord posts an embed payload to Discord using the provided webhook URL.
func PostEmbedToDiscord(webhookURL string, embed map[string]interface{}) error {
	payload := map[string]interface{}{
		"embeds": []map[string]interface{}{embed},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal embed payload: %w", err)
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to post embed to Discord: %w", err)
	}
	defer res.Body.Close()

	// Discord webhook typically returns 204 No Content or 200 OK on success.
	if res.StatusCode != http.StatusNoContent && res.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(res.Body)
		return fmt.Errorf("unexpected response from Discord (status %d): %s", res.StatusCode, string(respBody))
	}
	return nil
}
