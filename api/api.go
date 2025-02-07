package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/anthonyma94/ffxiv-status-checker/model"
)

// FetchServerStatus retrieves the FFXIV server statuses from the API.
func FetchServerStatus(apiURL string) ([]model.Server, error) {
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch server status: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read API response: %w", err)
	}

	var servers []model.Server
	if err := json.Unmarshal(body, &servers); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return servers, nil
}

// GetServersWithRetryWithTimeout repeatedly calls FetchServerStatus using exponential backoff
// until it succeeds or maxDuration has elapsed. If it still fails after maxDuration, it returns an error.
func GetServersWithRetryWithTimeout(apiURL string, maxDuration time.Duration) ([]model.Server, error) {
	start := time.Now()

	// Set the default backoff to 10 seconds.
	defaultBackoff := 10 * time.Second
	// If maxDuration is very short, scale down the initial backoff.
	var backoff time.Duration
	if maxDuration < defaultBackoff {
		backoff = maxDuration / 2
		if backoff <= 0 {
			backoff = 1 * time.Millisecond
		}
	} else {
		backoff = defaultBackoff
	}

	for {
		servers, err := FetchServerStatus(apiURL)
		if err == nil {
			return servers, nil
		}
		if time.Since(start) >= maxDuration {
			return nil, fmt.Errorf("failed after %v: %w", maxDuration, err)
		}
		fmt.Printf("Error fetching API: %v. Retrying in %v...\n", err, backoff)
		time.Sleep(backoff)
		backoff *= 2
	}
}

// GetServersWithRetry uses a maximum duration of 10 minutes.
// (This function is kept for backward compatibility; in production, you now call GetServersWithRetryWithTimeout
// with maxDuration = tickerInterval * 5.)
func GetServersWithRetry(apiURL string) ([]model.Server, error) {
	return GetServersWithRetryWithTimeout(apiURL, 10*time.Minute)
}

// GetServerByName searches for a server with the specified name.
func GetServerByName(servers []model.Server, serverName string) *model.Server {
	for i := range servers {
		if servers[i].Name == serverName {
			return &servers[i]
		}
	}
	return nil
}

// GetEmbedColor returns red (0xFF0000) if the congestion is "congested" (case-insensitive),
// otherwise it returns green (0x00FF00).
func GetEmbedColor(congestion string) int {
	if strings.ToLower(congestion) == "congested" {
		return 0xFF0000
	}
	return 0x00FF00
}
