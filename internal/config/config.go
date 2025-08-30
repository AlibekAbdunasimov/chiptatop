package config

import (
    "log"
    "os"

    "github.com/joho/godotenv"
)

type Config struct {
    TelegramBotToken string
    Environment      string
}

func Load() Config {
    _ = godotenv.Load()

    cfg := Config{
        TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
        Environment:      valueOrDefault(os.Getenv("ENVIRONMENT"), "development"),
    }

    if cfg.TelegramBotToken == "" {
        log.Fatal("TELEGRAM_BOT_TOKEN is required")
    }

    return cfg
}

func valueOrDefault(value string, def string) string {
    if value == "" {
        return def
    }
    return value
}


