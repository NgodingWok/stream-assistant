package handler

import (
	"context"
	"errors"
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
	// playCh is a single-element channel for play requests.
	// The dedicated player goroutine (started in ProcessEvents) is the only consumer.
	playCh chan string
}

/** New creates a Handler with the given config and TTS speaker. */
func New(cfg *config.Config, speaker *tts.Speaker) *Handler {
	return &Handler{
		cfg:     cfg,
		speaker: speaker,
		playCh:  make(chan string, 1),
	}
}

/** ProcessEvents reads events from the channel and dispatches them to the appropriate handler.
 *  It also starts the dedicated audio player goroutine for the lifetime of ctx. */
func (h *Handler) ProcessEvents(ctx context.Context, events <-chan gotiktoklive.Event) {
	go h.runPlayer(ctx)
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
	// Drain any pending (unprocessed) request so the newest always wins.
	select {
	case <-h.playCh:
	default:
	}
	h.playCh <- query
}

// runPlayer is the dedicated audio goroutine started once by ProcessEvents.
// It processes play requests from playCh sequentially. If a new request
// arrives while audio is playing, the current playback is cancelled and the
// new request starts immediately.
func (h *Handler) runPlayer(ctx context.Context) {
	var (
		cancelCurrent context.CancelFunc
		done          <-chan struct{}
	)

	for {
		if done == nil {
			// Idle: wait for the first play request.
			select {
			case <-ctx.Done():
				return
			case query := <-h.playCh:
				cancelCurrent, done = h.launchPlay(ctx, query)
			}
		} else {
			// Playing: wait for completion, a new request, or shutdown.
			select {
			case <-ctx.Done():
				cancelCurrent()
				return
			case <-done:
				cancelCurrent()
				done = nil
			case query := <-h.playCh:
				cancelCurrent()
				<-done
				cancelCurrent, done = h.launchPlay(ctx, query)
			}
		}
	}
}

// launchPlay starts a playback goroutine for query and returns its cancel func
// and a done channel that closes when playback finishes or errors.
func (h *Handler) launchPlay(ctx context.Context, query string) (context.CancelFunc, <-chan struct{}) {
	playCtx, cancel := context.WithCancel(ctx)
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		fmt.Printf("[play] searching: %s\n", query)
		videoID, err := youtube.Search(query)
		if err != nil {
			fmt.Printf("[play] search error: %v\n", err)
			return
		}
		fmt.Printf("[play] playing: %s\n", videoID)
		if err := youtube.Play(playCtx, videoID, h.cfg.PlayTimeout, h.cfg.TTSFolder); err != nil &&
			!errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			fmt.Printf("[play] error: %v\n", err)
		}
	}()
	return cancel, ch
}

func isStaleMessage(timestampMs, thresholdMs int64) bool {
	return time.Now().UnixMilli()-timestampMs > thresholdMs
}
