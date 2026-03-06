package handler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"NgodingWok/stream-assistant/internal/config"
	"NgodingWok/stream-assistant/internal/tts"
	"NgodingWok/stream-assistant/internal/youtube"

	"github.com/steampoweredtaco/gotiktoklive"
)

/** Handler processes TikTok live events with TTS and YouTube playback. */
type Handler struct {
	cfg     *config.Config
	speaker *tts.Speaker
}

/** New creates a Handler with the given config and TTS speaker. */
func New(cfg *config.Config, speaker *tts.Speaker) *Handler {
	return &Handler{
		cfg:     cfg,
		speaker: speaker,
	}
}

/** ProcessEvents reads events from the channel and dispatches them to the appropriate handler. */
func (h *Handler) ProcessEvents(ctx context.Context, events <-chan gotiktoklive.Event) {
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-events:
			if !ok {
				return
			}
			h.dispatch(ctx, event)
		}
	}
}

func (h *Handler) dispatch(ctx context.Context, event gotiktoklive.Event) {
	switch e := event.(type) {
	case gotiktoklive.UserEvent:
		fmt.Printf("[user] %s %s\n", e.Event, e.User.Username)
	case gotiktoklive.ViewersEvent:
		fmt.Printf("[viewers] %d\n", e.Viewers)
	case gotiktoklive.ChatEvent:
		h.handleChat(ctx, e)
	default:
		fmt.Printf("[event] %T: %+v\n", e, e)
	}
}

func (h *Handler) handleChat(ctx context.Context, e gotiktoklive.ChatEvent) {
	if isStaleMessage(e.Timestamp, h.cfg.ChatDelayMs) {
		return
	}

	fmt.Printf("[chat] %s: %s\n", e.User.Username, e.Comment)

	if strings.HasPrefix(e.Comment, "!play") {
		h.handlePlayCommand(ctx, e.Comment)
		return
	}

	if err := h.speaker.Speak(e.Comment); err != nil {
		fmt.Printf("[tts] error: %v\n", err)
	}
}

func (h *Handler) handlePlayCommand(ctx context.Context, comment string) {
	query := strings.TrimSpace(strings.TrimPrefix(comment, "!play"))
	if query == "" {
		fmt.Println("[play] missing search query")
		return
	}

	go func() {
		fmt.Printf("[play] searching: %s\n", query)
		videoID, err := youtube.Search(query)
		if err != nil {
			fmt.Printf("[play] search error: %v\n", err)
			return
		}
		fmt.Printf("[play] found video: %s\n", videoID)
		if err := youtube.Play(ctx, videoID, h.cfg.PlayTimeout); err != nil {
			fmt.Printf("[play] playback error: %v\n", err)
		}
	}()
}

func isStaleMessage(timestampMs, thresholdMs int64) bool {
	return time.Now().UnixMilli()-timestampMs > thresholdMs
}
