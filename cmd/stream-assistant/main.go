package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"NgodingWok/stream-assistant/internal/config"
	"NgodingWok/stream-assistant/internal/handler"
	"NgodingWok/stream-assistant/internal/tts"

	"github.com/steampoweredtaco/gotiktoklive"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: stream-assistant <tiktok-username>")
	}

	cfg, err := config.Load(os.Args[1])
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	tiktok, err := gotiktoklive.NewTikTok()
	if err != nil {
		return fmt.Errorf("creating tiktok client: %w", err)
	}

	live, err := tiktok.TrackUser(cfg.TikTokUsername)
	if err != nil {
		return fmt.Errorf("tracking user %s: %w", cfg.TikTokUsername, err)
	}
	fmt.Printf("Tracking user: %s\n", live.Info.Owner.Username)

	speaker := tts.NewSpeaker(cfg.TTSLanguage, cfg.TTSFolder)
	h := handler.New(cfg, speaker)
	h.ProcessEvents(ctx, live.Events)

	return nil
}
