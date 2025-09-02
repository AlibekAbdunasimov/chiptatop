package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramBotToken string
	Environment      string

	// Railway API Configuration - now optional since we'll get them dynamically
	RailwayXSRFToken string
	RailwayCookies   string
}

func Load() Config {
	_ = godotenv.Load()

	cfg := Config{
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		Environment:      valueOrDefault(os.Getenv("ENVIRONMENT"), "development"),

		// Railway API credentials - now optional, will be obtained dynamically
		RailwayXSRFToken: os.Getenv("RAILWAY_XSRF_TOKEN"),
		RailwayCookies:   os.Getenv("RAILWAY_COOKIES"),
	}

	if cfg.TelegramBotToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is required")
	}

	// Log railway API configuration status
	if cfg.RailwayXSRFToken != "" && cfg.RailwayCookies != "" {
		log.Printf("Railway API authentication configured from environment")
	} else {
		log.Printf("Railway API authentication will be obtained dynamically")
	}

	return cfg
}

func valueOrDefault(value string, def string) string {
	if value == "" {
		return def
	}
	return value
}
