# ChiptaTop Telegram Bot

Instant ticket availability alerts for trains, buses, and flights.

## Quick start

1. Create an environment file:
   - Copy `env.example` to `.env`
   - Set `TELEGRAM_BOT_TOKEN` to your bot token

2. Install dependencies:

```bash
make deps
```

3. Run locally:

```bash
make run
```

## Build

```bash
make build
```

## Docker

```bash
make docker-build
# Run with env vars
TELEGRAM_BOT_TOKEN=xxxx make docker-run
```

## Config

- `TELEGRAM_BOT_TOKEN`: Telegram bot token from BotFather
- `ENVIRONMENT`: development|production (default: development)

## Structure

- `cmd/bot`: application entrypoint
- `internal/config`: configuration loader
- `internal/bot`: Telegram bot setup and handlers
