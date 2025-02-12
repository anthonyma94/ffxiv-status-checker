# Final Fantasy XIV Status Checker for Discord

A simple server status checker that posts any changes to a Discord channel.

## Install & Run
`docker run -e DISCORD_WEBHOOK_URL=<webhook here> -v <data dir here>:/app/data ghcr.io/anthonyma94/ffxiv-status-checker:latest`

## Environment Variables
| Key | Description | Default | Required |
| -- |  -- | -- | -- |
| DISCORD_WEBHOOK_URL | The webhook used to post messages to Discord. | `null` | `true` |
| TICKER_INTERVAL | How often should the checker check. The format is Golang's duration format. | `1m` | `false` |
| DEBUG | Debug will ignore cache and post a single message to Discord before exiting. | `false` | `false` |
| SERVER_NAME | Name of the FFXIV server to check. | `"Faerie"` | `false` |