package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

/** Config holds all application configuration loaded from environment variables. */
type Config struct {
	TikTokUsername string
	TTSLanguage    string
	TTSFolder      string
	ChatDelayMs    int64
	PlayTimeout    time.Duration
}

/** Load reads configuration from the given username arg and environment variables. */
func Load(username string) (*Config, error) {
	if username == "" {
		return nil, fmt.Errorf("username argument is required")
	}

	chatDelayMs, err := strconv.ParseInt(getEnv("CHAT_DELAY_MS", "10000"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid CHAT_DELAY_MS: %w", err)
	}

	playTimeoutMin, err := strconv.Atoi(getEnv("PLAY_TIMEOUT_MIN", "15"))
	if err != nil {
		return nil, fmt.Errorf("invalid PLAY_TIMEOUT_MIN: %w", err)
	}

	return &Config{
		TikTokUsername: username,
		TTSLanguage:    getEnv("TTS_LANGUAGE", "id"),
		TTSFolder:      getEnv("TTS_FOLDER", ".tmp"),
		ChatDelayMs:    chatDelayMs,
		PlayTimeout:    time.Duration(playTimeoutMin) * time.Minute,
	}, nil
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
