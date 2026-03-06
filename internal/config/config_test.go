package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad_ValidUsername(t *testing.T) {
	cfg, err := Load("testuser")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.TikTokUsername != "testuser" {
		t.Errorf("expected username 'testuser', got %q", cfg.TikTokUsername)
	}
}

func TestLoad_EmptyUsername(t *testing.T) {
	_, err := Load("")
	if err == nil {
		t.Fatal("expected error for empty username")
	}
}

func TestLoad_Defaults(t *testing.T) {
	os.Unsetenv("CHAT_DELAY_MS")
	os.Unsetenv("PLAY_TIMEOUT_MIN")
	os.Unsetenv("TTS_LANGUAGE")
	os.Unsetenv("TTS_FOLDER")

	cfg, err := Load("testuser")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ChatDelayMs != 10000 {
		t.Errorf("expected ChatDelayMs 10000, got %d", cfg.ChatDelayMs)
	}
	if cfg.PlayTimeout != 15*time.Minute {
		t.Errorf("expected PlayTimeout 15m, got %v", cfg.PlayTimeout)
	}
	if cfg.TTSLanguage != "id" {
		t.Errorf("expected TTSLanguage 'id', got %q", cfg.TTSLanguage)
	}
	if cfg.TTSFolder != ".tmp" {
		t.Errorf("expected TTSFolder '.tmp', got %q", cfg.TTSFolder)
	}
}

func TestLoad_EnvOverrides(t *testing.T) {
	t.Setenv("CHAT_DELAY_MS", "5000")
	t.Setenv("PLAY_TIMEOUT_MIN", "30")
	t.Setenv("TTS_LANGUAGE", "en")
	t.Setenv("TTS_FOLDER", "/tmp/tts")

	cfg, err := Load("testuser")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ChatDelayMs != 5000 {
		t.Errorf("expected ChatDelayMs 5000, got %d", cfg.ChatDelayMs)
	}
	if cfg.PlayTimeout != 30*time.Minute {
		t.Errorf("expected PlayTimeout 30m, got %v", cfg.PlayTimeout)
	}
	if cfg.TTSLanguage != "en" {
		t.Errorf("expected TTSLanguage 'en', got %q", cfg.TTSLanguage)
	}
	if cfg.TTSFolder != "/tmp/tts" {
		t.Errorf("expected TTSFolder '/tmp/tts', got %q", cfg.TTSFolder)
	}
}

func TestLoad_InvalidChatDelayMs(t *testing.T) {
	t.Setenv("CHAT_DELAY_MS", "notanumber")

	_, err := Load("testuser")
	if err == nil {
		t.Fatal("expected error for invalid CHAT_DELAY_MS")
	}
}

func TestLoad_InvalidPlayTimeoutMin(t *testing.T) {
	t.Setenv("CHAT_DELAY_MS", "10000")
	t.Setenv("PLAY_TIMEOUT_MIN", "notanumber")

	_, err := Load("testuser")
	if err == nil {
		t.Fatal("expected error for invalid PLAY_TIMEOUT_MIN")
	}
}
