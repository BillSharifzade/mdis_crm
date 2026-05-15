package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port             string
	Database         DatabaseConfig
	TelegramBotToken string
}

type DatabaseConfig struct {
	DSN string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) {
		log.Printf("Error loading .env file: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Println("WARNING: DATABASE_URL not set in environment")
	}

	tgToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if tgToken == "" {
		log.Println("WARNING: TELEGRAM_BOT_TOKEN not set — outbound Telegram disabled")
	}

	return &Config{
		Port:             port,
		Database:         DatabaseConfig{DSN: dsn},
		TelegramBotToken: tgToken,
	}, nil
}
