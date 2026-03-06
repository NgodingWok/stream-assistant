package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"NgodingWok/stream-assistant/internal/config"
	"NgodingWok/stream-assistant/internal/handler"
	"NgodingWok/stream-assistant/internal/tts"

	"github.com/steampoweredtaco/gotiktoklive"
)

const pollInterval = 30 * time.Second

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

	roomInfo, err := tiktok.GetLiveRoomUserInfo(cfg.TikTokUsername)
	if err != nil {
		return fmt.Errorf("user %s not found: %w", cfg.TikTokUsername, err)
	}

	if err := waitForLive(ctx, tiktok, roomInfo, cfg.TikTokUsername); err != nil {
		return err
	}

	live, err := tiktok.TrackUser(cfg.TikTokUsername)
	if err != nil {
		return fmt.Errorf("tracking user %s: %w", cfg.TikTokUsername, err)
	}

	displayName := cfg.TikTokUsername
	if live.Info != nil && live.Info.Owner != nil {
		displayName = live.Info.Owner.Username
	}
	fmt.Printf("Tracking user: %s\n", displayName)

	speaker := tts.NewSpeaker(cfg.TTSLanguage, cfg.TTSFolder)
	h := handler.New(cfg, speaker)
	h.ProcessEvents(ctx, live.Events)

	return nil
}

// waitForLive polls until the user goes live, the context is cancelled, or an error occurs.
func waitForLive(ctx context.Context, tiktok *gotiktoklive.TikTok, roomInfo gotiktoklive.LiveRoomUserInfo, username string) error {
	for {
		isLive, err := tiktok.IsLive(roomInfo)
		if err != nil {
			return fmt.Errorf("checking live status for %s: %w", username, err)
		}
		if isLive {
			return nil
		}

		fmt.Printf("user %s is not live yet, retrying in %s...\n", username, pollInterval)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(pollInterval):
		}
	}
}
