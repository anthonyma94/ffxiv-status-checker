package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"github.com/anthonyma94/ffxiv-status-checker/api"
	"github.com/anthonyma94/ffxiv-status-checker/discord"
	"github.com/anthonyma94/ffxiv-status-checker/storage"
)

// RunStatusCheck performs one iteration of fetching the server status, handling retries,
// and posting an embed (or error embed) to Discord.
func RunStatusCheck(apiURL, discordWebhook, serverName, stateFileName string, maxRetryDuration time.Duration, debug bool) {
	servers, err := api.GetServersWithRetryWithTimeout(apiURL, maxRetryDuration)
	if err != nil {
		errorEmbed := map[string]interface{}{
			"title":       "Error Retrieving Server Status",
			"description": fmt.Sprintf("Error retrieving status for **%s** after %v:\n`%v`", serverName, maxRetryDuration, err),
			"color":       0xFF0000,
			"timestamp":   time.Now().Format(time.RFC3339),
		}
		log.Printf("Error after retries: %v", err)
		if err := discord.PostEmbedToDiscord(discordWebhook, errorEmbed); err != nil {
			log.Printf("Error posting error message to Discord: %v", err)
		} else {
			log.Println("Posted error message to Discord successfully.")
		}
		return
	}

	currentServer := api.GetServerByName(servers, serverName)
	if currentServer == nil {
		log.Printf("Server '%s' not found in the API response.", serverName)
		return
	}

	if !debug {
		lastServer, err := storage.LoadLastServerState(stateFileName)
		if err != nil {
			log.Printf("Error loading last state for server %s: %v", serverName, err)
		}
		if lastServer != nil &&
			currentServer.Status == lastServer.Status &&
			currentServer.Congestion == lastServer.Congestion &&
			currentServer.Creation == lastServer.Creation {
			log.Printf("No change in %s's status; nothing to post.", serverName)
			return
		}
	}

	var color int
	if strings.ToLower(currentServer.Congestion) == "congested" {
		color = 0xFF0000
	} else {
		color = 0x00FF00
	}

	embed := map[string]interface{}{
		"title":       fmt.Sprintf("Updated Status for **%s**", currentServer.Name),
		"description": fmt.Sprintf("Status: %s\nCongestion: %s\nCreation: %s", currentServer.Status, currentServer.Congestion, currentServer.Creation),
		"color":       color,
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	logMessage := fmt.Sprintf("Updated Status for %s:\nStatus: %s\nCongestion: %s\nCreation: %s",
		currentServer.Name, currentServer.Status, currentServer.Congestion, currentServer.Creation)
	log.Println(logMessage)

	if err := discord.PostEmbedToDiscord(discordWebhook, embed); err != nil {
		log.Printf("Error posting to Discord: %v", err)
	} else {
		log.Println("Posted server status to Discord successfully.")
	}

	if err := storage.SaveServerState(stateFileName, currentServer); err != nil {
		log.Printf("Error saving state for server %s: %v", serverName, err)
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error reading .env file; proceeding with system environment variables.")
	}

	debug := os.Getenv("DEBUG") == "true"

	serverName := os.Getenv("SERVER_NAME")
	if serverName == "" {
		serverName = "Faerie"
	}
	apiURL := "https://api.xivstatus.com/api/servers"
	discordWebhook := os.Getenv("DISCORD_WEBHOOK_URL")
	if discordWebhook == "" {
		log.Fatal("Environment variable DISCORD_WEBHOOK_URL is not set.")
	}

	tickerInterval := 1 * time.Minute
	if intervalStr := os.Getenv("TICKER_INTERVAL"); intervalStr != "" {
		if d, err := time.ParseDuration(intervalStr); err == nil {
			tickerInterval = d
		} else {
			log.Printf("Invalid TICKER_INTERVAL value %q, using default of 1m: %v", intervalStr, err)
		}
	}

	maxRetryDuration := tickerInterval * 5
	stateFileName := storage.FileNameForServer(serverName)

	if debug {
		log.Println("DEBUG mode enabled: always posting message and disabling ticker.")
		RunStatusCheck(apiURL, discordWebhook, serverName, stateFileName, maxRetryDuration, debug)
		return
	}

	ticker := time.NewTicker(tickerInterval)
	defer ticker.Stop()

	RunStatusCheck(apiURL, discordWebhook, serverName, stateFileName, maxRetryDuration, debug)
	log.Println("Starting ticker-based checks...")
	for range ticker.C {
		RunStatusCheck(apiURL, discordWebhook, serverName, stateFileName, maxRetryDuration, debug)
	}
}
