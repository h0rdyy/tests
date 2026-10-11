package config

import "os"

type Config struct {
	Port     string
	DBPath   string
	LogPath  string
	SMTPHost string
	SMTPPort string
	SMTPUser string
	SMTPPass string
	Receiver string
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func Load() Config {
	return Config{
		Port:     env("PORT", "5000"),
		DBPath:   env("DB_PATH", "data/cybersec.db"),
		LogPath:  env("LOG_PATH", "logs/app.log"),
		SMTPHost: os.Getenv("SMTP_HOST"),
		SMTPPort: env("SMTP_PORT", "465"),
		SMTPUser: os.Getenv("SMTP_USER"),
		SMTPPass: os.Getenv("SMTP_PASSWORD"),
		Receiver: os.Getenv("RECEIVER_EMAIL"),
	}
}
