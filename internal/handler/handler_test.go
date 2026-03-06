package handler

import (
	"context"
	"testing"
	"time"

	"NgodingWok/stream-assistant/internal/config"
	"NgodingWok/stream-assistant/internal/tts"

	"github.com/steampoweredtaco/gotiktoklive"
)

func TestIsStaleMessage(t *testing.T) {
	now := time.Now().UnixMilli()

	tests := []struct {
		name        string
		timestampMs int64
		thresholdMs int64
		wantStale   bool
	}{
		{
			name:        "fresh message",
			timestampMs: now,
			thresholdMs: 10000,
			wantStale:   false,
		},
		{
			name:        "stale message",
			timestampMs: now - 20000,
			thresholdMs: 10000,
			wantStale:   true,
		},
		{
			name:        "exactly at threshold",
			timestampMs: now - 9999,
			thresholdMs: 10000,
			wantStale:   false,
		},
		{
			name:        "message from the future",
			timestampMs: now + 5000,
			thresholdMs: 10000,
			wantStale:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isStaleMessage(tt.timestampMs, tt.thresholdMs)
			if got != tt.wantStale {
				t.Errorf("isStaleMessage(%d, %d) = %v, want %v",
					tt.timestampMs, tt.thresholdMs, got, tt.wantStale)
			}
		})
	}
}

func TestNew(t *testing.T) {
	cfg := &config.Config{
		TikTokUsername: "testuser",
		ChatDelayMs:    10000,
		PlayTimeout:    15 * time.Minute,
	}
	speaker := tts.NewSpeaker("id", t.TempDir())

	h := New(cfg, speaker)
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
	if h.cfg != cfg {
		t.Error("handler config does not match")
	}
	if h.speaker != speaker {
		t.Error("handler speaker does not match")
	}
}

func TestProcessEvents_ContextCancellation(t *testing.T) {
	cfg := &config.Config{ChatDelayMs: 10000, PlayTimeout: time.Minute}
	speaker := tts.NewSpeaker("id", t.TempDir())
	h := New(cfg, speaker)

	ctx, cancel := context.WithCancel(context.Background())
	events := make(chan gotiktoklive.Event, 1)

	done := make(chan struct{})
	go func() {
		h.ProcessEvents(ctx, events)
		close(done)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("ProcessEvents did not exit after context cancellation")
	}
}

func TestProcessEvents_ChannelClosed(t *testing.T) {
	cfg := &config.Config{ChatDelayMs: 10000, PlayTimeout: time.Minute}
	speaker := tts.NewSpeaker("id", t.TempDir())
	h := New(cfg, speaker)

	events := make(chan gotiktoklive.Event)

	done := make(chan struct{})
	go func() {
		h.ProcessEvents(context.Background(), events)
		close(done)
	}()

	close(events)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("ProcessEvents did not exit after channel closed")
	}
}
