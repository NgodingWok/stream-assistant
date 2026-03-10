package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"NgodingWok/stream-assistant/internal/tts"
	"NgodingWok/stream-assistant/internal/youtube"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"

	"github.com/steampoweredtaco/gotiktoklive"
)

const (
	pollInterval  = 30 * time.Second
	maxLogEntries = 500
)

// streamApp holds all GUI state and session lifecycle.
type streamApp struct {
	win fyne.Window

	// Bindings are goroutine-safe; UI updates happen automatically via listeners.
	logEntries  binding.StringList
	statusText  binding.String
	viewerCount binding.String

	// Settings form widgets – read once on session start, then disabled.
	usernameEntry    *widget.Entry
	ttsLangEntry     *widget.Entry
	ttsFolderEntry   *widget.Entry
	chatDelayEntry   *widget.Entry
	playTimeoutEntry *widget.Entry

	startStopBtn *widget.Button
	logList      *widget.List

	// Session state.
	cancel  context.CancelFunc
	running bool
	mu      sync.Mutex

	// playCh carries !play queries to the dedicated audio player goroutine.
	// Created fresh per session; capacity 1 so the newest request always wins.
	playCh chan string
}

func main() {
	a := app.New()
	w := a.NewWindow("Stream Assistant")
	w.Resize(fyne.NewSize(960, 640))

	sa := &streamApp{
		win:         w,
		logEntries:  binding.NewStringList(),
		statusText:  binding.NewString(),
		viewerCount: binding.NewString(),
	}
	_ = sa.statusText.Set("Idle")
	_ = sa.viewerCount.Set("—")

	w.SetContent(sa.buildUI())
	w.ShowAndRun()
}

func (sa *streamApp) buildUI() fyne.CanvasObject {
	// Settings form 
	sa.usernameEntry = widget.NewEntry()
	sa.usernameEntry.SetPlaceHolder("tiktok_username")

	sa.ttsLangEntry = widget.NewEntry()
	sa.ttsLangEntry.SetText("id")

	sa.ttsFolderEntry = widget.NewEntry()
	sa.ttsFolderEntry.SetText(".tmp")

	sa.chatDelayEntry = widget.NewEntry()
	sa.chatDelayEntry.SetText("10000")

	sa.playTimeoutEntry = widget.NewEntry()
	sa.playTimeoutEntry.SetText("15")

	form := widget.NewForm(
		widget.NewFormItem("TikTok Username", sa.usernameEntry),
		widget.NewFormItem("TTS Language", sa.ttsLangEntry),
		widget.NewFormItem("TTS Folder", sa.ttsFolderEntry),
		widget.NewFormItem("Chat Delay (ms)", sa.chatDelayEntry),
		widget.NewFormItem("Play Timeout (min)", sa.playTimeoutEntry),
	)

	// Action buttons
	sa.startStopBtn = widget.NewButton("Start", sa.toggleSession)
	sa.startStopBtn.Importance = widget.HighImportance

	clearBtn := widget.NewButton("Clear Log", func() {
		_ = sa.logEntries.Set([]string{})
	})

	btnRow := container.NewHBox(sa.startStopBtn, clearBtn)

	// Status bar
	statusBar := container.NewHBox(
		widget.NewLabel("Status:"),
		widget.NewLabelWithData(sa.statusText),
		widget.NewSeparator(),
		widget.NewLabel("Viewers:"),
		widget.NewLabelWithData(sa.viewerCount),
	)

	// Log list
	sa.logList = widget.NewListWithData(
		sa.logEntries,
		func() fyne.CanvasObject {
			l := widget.NewLabel("")
			l.TextStyle = fyne.TextStyle{Monospace: true}
			return l
		},
		func(item binding.DataItem, obj fyne.CanvasObject) {
			text, _ := item.(binding.String).Get()
			obj.(*widget.Label).SetText(text)
		},
	)

	top := container.NewVBox(
		form,
		btnRow,
		widget.NewSeparator(),
		statusBar,
		widget.NewSeparator(),
	)

	return container.NewBorder(top, nil, nil, nil, sa.logList)
}

// toggleSession starts or stops the active monitoring session.
func (sa *streamApp) toggleSession() {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	if sa.running {
		sa.cancel()
		return
	}

	username := strings.TrimSpace(sa.usernameEntry.Text)
	if username == "" {
		_ = sa.statusText.Set("Error: username required")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	sa.cancel = cancel
	sa.running = true

	sa.startStopBtn.SetText("Stop")
	sa.startStopBtn.Importance = widget.DangerImportance
	sa.startStopBtn.Refresh()
	sa.setFormEnabled(false)

	go sa.runSession(ctx, username)
}

// runSession runs the full TikTok monitoring lifecycle in a goroutine.
func (sa *streamApp) runSession(ctx context.Context, username string) {
	defer func() {
		sa.mu.Lock()
		sa.running = false
		sa.mu.Unlock()

		_ = sa.viewerCount.Set("—")
		fyne.Do(func() {
			sa.startStopBtn.SetText("Start")
			sa.startStopBtn.Importance = widget.HighImportance
			sa.startStopBtn.Refresh()
			sa.setFormEnabled(true)
		})
	}()

	sa.appendLog("▶ Starting session for @" + username)
	_ = sa.statusText.Set("Connecting…")

	// read settings from form
	chatDelayMs := int64(parseIntOrDefault(sa.chatDelayEntry.Text, 10000))
	playTimeout := time.Duration(parseIntOrDefault(sa.playTimeoutEntry.Text, 15)) * time.Minute
	ttsLang := fallbackStr(sa.ttsLangEntry.Text, "id")
	ttsFolder := fallbackStr(sa.ttsFolderEntry.Text, ".tmp")

	speaker := tts.NewSpeaker(ttsLang, ttsFolder)

	// create the player channel and start the dedicated audio goroutine
	sa.playCh = make(chan string, 1)
	go sa.runPlayer(ctx, playTimeout, ttsFolder)

	// connect
	tikTok, err := gotiktoklive.NewTikTok()
	if err != nil {
		sa.appendLog("✗ TikTok client error: " + err.Error())
		_ = sa.statusText.Set("Error")
		return
	}

	roomInfo, err := tikTok.GetLiveRoomUserInfo(username)
	if err != nil {
		sa.appendLog(fmt.Sprintf("✗ User @%s not found: %v", username, err))
		_ = sa.statusText.Set("Error")
		return
	}

	// wait for live
	_ = sa.statusText.Set("Waiting for live…")
	sa.appendLog(fmt.Sprintf("⏳ Waiting for @%s to go live…", username))

	for {
		isLive, err := tikTok.IsLive(roomInfo)
		if err != nil {
			sa.appendLog("✗ Live check error: " + err.Error())
			_ = sa.statusText.Set("Error")
			return
		}
		if isLive {
			break
		}
		sa.appendLog(fmt.Sprintf("⏳ Not live, retrying in %s…", pollInterval))
		select {
		case <-ctx.Done():
			sa.appendLog("◼ Session cancelled")
			_ = sa.statusText.Set("Stopped")
			return
		case <-time.After(pollInterval):
		}
	}

	// track user
	live, err := tikTok.TrackUser(username)
	if err != nil {
		sa.appendLog(fmt.Sprintf("✗ TrackUser error: %v", err))
		_ = sa.statusText.Set("Error")
		return
	}

	displayName := username
	if live.Info != nil && live.Info.Owner != nil {
		displayName = live.Info.Owner.Username
	}
	sa.appendLog(fmt.Sprintf("✔ Tracking @%s", displayName))
	_ = sa.statusText.Set("Live • @" + displayName)

	// event loop
	for {
		select {
		case <-ctx.Done():
			sa.appendLog("◼ Session stopped")
			_ = sa.statusText.Set("Stopped")
			return
		case event, ok := <-live.Events:
			if !ok {
				sa.appendLog("⚠ Event stream closed")
				_ = sa.statusText.Set("Disconnected")
				return
			}
			sa.dispatchEvent(ctx, event, chatDelayMs, playTimeout, speaker)
		}
	}
}

func (sa *streamApp) dispatchEvent(
	ctx context.Context,
	event gotiktoklive.Event,
	chatDelayMs int64,
	playTimeout time.Duration,
	speaker *tts.Speaker,
) {
	switch e := event.(type) {
	case gotiktoklive.UserEvent:
		sa.appendLog(fmt.Sprintf("[user] %s %s", e.Event, e.User.Username))

	case gotiktoklive.ViewersEvent:
		_ = sa.viewerCount.Set(strconv.Itoa(e.Viewers))

	case gotiktoklive.ChatEvent:
		if isStaleMessage(e.Timestamp, chatDelayMs) {
			return
		}
		sa.appendLog(fmt.Sprintf("[chat] %s: %s", e.User.Username, e.Comment))
		if strings.HasPrefix(e.Comment, "!play") {
			query := strings.TrimSpace(strings.TrimPrefix(e.Comment, "!play"))
			if query != "" {
				// Drain any pending (unprocessed) request so the newest always wins.
				select {
				case <-sa.playCh:
				default:
				}
				sa.playCh <- query
			}
			return
		}
		go func() {
			if err := speaker.Speak(e.Comment); err != nil {
				sa.appendLog(fmt.Sprintf("[tts] %v", err))
			}
		}()

	default:
		sa.appendLog(fmt.Sprintf("[event] %T", event))
	}
}

// runPlayer is the dedicated audio goroutine started once per session.
// It processes play requests from playCh sequentially. If a new request arrives
// while audio is playing, the current playback is cancelled and the new request
// starts immediately.
func (sa *streamApp) runPlayer(ctx context.Context, timeout time.Duration, tmpDir string) {
	var (
		cancelCurrent context.CancelFunc
		done          <-chan struct{}
	)

	for {
		if done == nil {
			select {
			case <-ctx.Done():
				return
			case query := <-sa.playCh:
				cancelCurrent, done = sa.launchPlay(ctx, query, timeout, tmpDir)
			}
		} else {
			select {
			case <-ctx.Done():
				cancelCurrent()
				return
			case <-done:
				cancelCurrent()
				done = nil
			case query := <-sa.playCh:
				cancelCurrent()
				<-done
				cancelCurrent, done = sa.launchPlay(ctx, query, timeout, tmpDir)
			}
		}
	}
}

// launchPlay starts a playback goroutine for query and returns its cancel func
// and a done channel that closes when playback finishes or errors.
func (sa *streamApp) launchPlay(ctx context.Context, query string, timeout time.Duration, tmpDir string) (context.CancelFunc, <-chan struct{}) {
	playCtx, cancel := context.WithCancel(ctx)
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		sa.appendLog("[play] searching: " + query)
		videoID, err := youtube.Search(query)
		if err != nil {
			sa.appendLog(fmt.Sprintf("[play] search error: %v", err))
			return
		}
		sa.appendLog("[play] playing: " + videoID)
		if err := youtube.Play(playCtx, videoID, timeout, tmpDir); err != nil &&
			!errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			sa.appendLog(fmt.Sprintf("[play] error: %v", err))
		}
	}()
	return cancel, ch
}

// appendLog appends a timestamped line to the log list and scrolls to bottom.
func (sa *streamApp) appendLog(msg string) {
	entry := time.Now().Format("15:04:05") + "  " + msg
	items, _ := sa.logEntries.Get()
	if len(items) >= maxLogEntries {
		_ = sa.logEntries.Set(append(items[1:], entry))
	} else {
		_ = sa.logEntries.Append(entry)
	}
	if sa.logList != nil {
		if n := sa.logEntries.Length(); n > 0 {
			fyne.Do(func() {
				sa.logList.ScrollTo(n - 1)
			})
		}
	}
}

func (sa *streamApp) setFormEnabled(enabled bool) {
	for _, e := range []*widget.Entry{
		sa.usernameEntry, sa.ttsLangEntry, sa.ttsFolderEntry,
		sa.chatDelayEntry, sa.playTimeoutEntry,
	} {
		if enabled {
			e.Enable()
		} else {
			e.Disable()
		}
	}
}

func parseIntOrDefault(s string, def int) int {
	v, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || v <= 0 {
		return def
	}
	return v
}

func fallbackStr(s, fallback string) string {
	if s = strings.TrimSpace(s); s == "" {
		return fallback
	}
	return s
}

func isStaleMessage(timestampMs, thresholdMs int64) bool {
	return time.Now().UnixMilli()-timestampMs > thresholdMs
}