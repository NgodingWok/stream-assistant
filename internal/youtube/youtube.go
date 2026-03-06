package youtube

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

var videoIDPattern = regexp.MustCompile(`watch\?v=([A-Za-z0-9_-]{11})`)

/** Search queries YouTube and returns the first matching video ID. */
func Search(query string) (string, error) {
	searchURL := "https://www.youtube.com/results?search_query=" + url.QueryEscape(query)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(searchURL)
	if err != nil {
		return "", fmt.Errorf("youtube request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("youtube returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading youtube response: %w", err)
	}

	matches := videoIDPattern.FindSubmatch(body)
	if len(matches) < 2 {
		return "", fmt.Errorf("no video found for query: %s", query)
	}
	return string(matches[1]), nil
}

/** Play streams audio for the given video ID using yt-dlp and ffplay.
 *  It extracts the direct stream URL then plays it without downloading. */
func Play(ctx context.Context, videoID string, timeout time.Duration) error {
	ytdlp, err := exec.LookPath("yt-dlp")
	if err != nil {
		return fmt.Errorf("yt-dlp not found: %w", err)
	}
	ffplay, err := exec.LookPath("ffplay")
	if err != nil {
		return fmt.Errorf("ffplay not found: %w", err)
	}

	playCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	videoURL := "https://www.youtube.com/watch?v=" + videoID

	streamURL, err := extractStreamURL(playCtx, ytdlp, videoURL)
	if err != nil {
		return err
	}

	return streamAudio(playCtx, ffplay, streamURL)
}

func extractStreamURL(ctx context.Context, ytdlpPath, videoURL string) (string, error) {
	cmd := exec.CommandContext(ctx, ytdlpPath, "-f", "bestaudio", "-g", videoURL)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("yt-dlp extract failed: %w", err)
	}
	streamURL := strings.TrimSpace(string(out))
	if streamURL == "" {
		return "", fmt.Errorf("yt-dlp returned empty stream URL")
	}
	return streamURL, nil
}

func streamAudio(ctx context.Context, ffplayPath, streamURL string) error {
	cmd := exec.CommandContext(ctx, ffplayPath, "-autoexit", "-nodisp", "-loglevel", "error", streamURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
