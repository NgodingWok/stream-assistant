//go:build integration

package integration

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"NgodingWok/stream-assistant/internal/youtube"
)

func TestYouTube_Search(t *testing.T) {
	videoID, err := youtube.Search("Rick Astley Never Gonna Give You Up")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(videoID) != 11 {
		t.Fatalf("expected 11-char video ID, got %q", videoID)
	}
	t.Logf("found video ID: %s", videoID)
}

func TestYouTube_SearchEmpty(t *testing.T) {
	_, err := youtube.Search("asdkjhqwekjhasdlkjqhwelkjhasdzxcmnb")
	if err == nil {
		t.Log("nonsense query returned a result (YouTube may still match)")
	}
}

func TestYouTube_Play(t *testing.T) {
	if _, err := exec.LookPath("yt-dlp"); err != nil {
		t.Skip("yt-dlp not installed, skipping playback test")
	}
	if _, err := exec.LookPath("ffplay"); err != nil {
		t.Skip("ffplay not installed, skipping playback test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := youtube.Play(ctx, "dQw4w9WgXcQ", 30*time.Second)
	if err != nil && ctx.Err() == nil {
		t.Fatalf("Play failed: %v", err)
	}
	t.Log("playback completed or timed out as expected")
}
