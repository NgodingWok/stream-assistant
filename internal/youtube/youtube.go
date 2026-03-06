package youtube

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"time"

	ytdlp "NgodingWok/stream-assistant/third_party"
)

var videoIDPattern = regexp.MustCompile(`watch\?v=([A-Za-z0-9_-]{11})`)

// android_vr (Oculus Quest) InnerTube client constants.
// android_vr produces stream URLs that YouTube CDN accepts without restriction,
// unlike android_sdkless which CDN started blocking (see yt-dlp/yt-dlp#15726).
const (
	vrClientName    = "ANDROID_VR"
	vrClientVersion = "1.60.19"
	vrUserAgent     = "com.google.android.apps.youtube.vr.oculus/1.60.19 (Linux; U; Android 12L; eureka-user Build/SQ3A.220605.009.A1) gzip"
	vrAPIKey        = "AIzaSyAO_FJ2SlqU8Q4STEHLGCilw_Y9_11qcW8"
	vrSDKVersion    = 32

	// maxBotRetries is the number of times to retry with a fresh visitorData
	// when YouTube returns LOGIN_REQUIRED (bot detection triggered).
	maxBotRetries = 3
)

// errLoginRequired is returned when YouTube requires sign-in after all retries.
var errLoginRequired = errors.New("youtube login required")

type vrRequest struct {
	VideoID        string    `json:"videoId"`
	ContentCheckOK bool      `json:"contentCheckOk"`
	RacyCheckOk    bool      `json:"racyCheckOk"`
	Context        vrContext `json:"context"`
}

type vrContext struct {
	Client vrClient `json:"client"`
}

type vrClient struct {
	ClientName        string `json:"clientName"`
	ClientVersion     string `json:"clientVersion"`
	AndroidSDKVersion int    `json:"androidSdkVersion"`
	UserAgent         string `json:"userAgent"`
	VisitorData       string `json:"visitorData,omitempty"`
	HL                string `json:"hl"`
	GL                string `json:"gl"`
	TimeZone          string `json:"timeZone"`
}

type vrFormat struct {
	ItagNo       int    `json:"itag"`
	URL          string `json:"url"`
	MimeType     string `json:"mimeType"`
	Bitrate      int    `json:"bitrate"`
	AudioQuality string `json:"audioQuality"`
}

type vrResponse struct {
	PlayabilityStatus struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	} `json:"playabilityStatus"`
	StreamingData struct {
		AdaptiveFormats []vrFormat `json:"adaptiveFormats"`
	} `json:"streamingData"`
}

const visitorIDChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// generateVisitorData creates a random base64-encoded protobuf visitor token.
// YouTube uses this to distinguish sessions; providing a unique one helps bypass
// bot detection that triggers when no visitor identity is present in the request.
// Format: field 1 (LEN, 11 bytes visitor ID) | field 2 (VARINT, value 17).
func generateVisitorData() string {
	id := make([]byte, 11)
	for i := range id {
		id[i] = visitorIDChars[rand.IntN(len(visitorIDChars))]
	}
	data := append([]byte{0x0a, 0x0b}, id...)
	data = append(data, 0x10, 0x11)
	return base64.StdEncoding.EncodeToString(data)
}

// resolveVRAudioURL calls the InnerTube API using the android_vr client to get a
// CDN-allowed audio stream URL for the given video ID. Retries up to maxBotRetries
// times with fresh visitorData when YouTube returns LOGIN_REQUIRED (bot detection).
// Returns errLoginRequired if all attempts fail, allowing the caller to fall back.
func resolveVRAudioURL(ctx context.Context, videoID string) (string, error) {
	for attempt := 0; attempt < maxBotRetries; attempt++ {
		streamURL, status, err := tryResolveVRAudioURL(ctx, videoID, generateVisitorData())
		if err != nil {
			return "", err
		}
		if streamURL != "" {
			return streamURL, nil
		}
		if status != "LOGIN_REQUIRED" {
			return "", fmt.Errorf("video unavailable (%s): %s", videoID, status)
		}
		// LOGIN_REQUIRED likely means bot detection triggered; retry with fresh visitorData.
	}
	return "", errLoginRequired
}

// tryResolveVRAudioURL performs a single InnerTube player API call.
// Returns the best audio URL (or empty string on LOGIN_REQUIRED), the playability status,
// and any transport/decode error.
func tryResolveVRAudioURL(ctx context.Context, videoID, visitorData string) (string, string, error) {
	body, err := json.Marshal(vrRequest{
		VideoID:        videoID,
		ContentCheckOK: true,
		RacyCheckOk:    true,
		Context: vrContext{
			Client: vrClient{
				ClientName:        vrClientName,
				ClientVersion:     vrClientVersion,
				AndroidSDKVersion: vrSDKVersion,
				UserAgent:         vrUserAgent,
				VisitorData:       visitorData,
				HL:                "en",
				GL:                "US",
				TimeZone:          "UTC",
			},
		},
	})
	if err != nil {
		return "", "", fmt.Errorf("marshal innertube request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://www.youtube.com/youtubei/v1/player?key="+vrAPIKey, bytes.NewReader(body))
	if err != nil {
		return "", "", fmt.Errorf("create innertube request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", vrUserAgent)
	req.Header.Set("X-YouTube-Client-Name", "28")
	req.Header.Set("X-YouTube-Client-Version", vrClientVersion)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("innertube request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("innertube returned %d", resp.StatusCode)
	}

	var result vrResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", fmt.Errorf("decode innertube response: %w", err)
	}

	status := result.PlayabilityStatus.Status
	if status == "LOGIN_REQUIRED" {
		return "", status, nil
	}

	audioFormats := make([]vrFormat, 0)
	for _, f := range result.StreamingData.AdaptiveFormats {
		if f.URL != "" && len(f.MimeType) >= 5 && f.MimeType[:5] == "audio" {
			audioFormats = append(audioFormats, f)
		}
	}
	if len(audioFormats) == 0 {
		return "", status, fmt.Errorf("no audio formats available for %s (status: %s)", videoID, status)
	}

	sort.Slice(audioFormats, func(i, j int) bool {
		return audioFormats[i].Bitrate > audioFormats[j].Bitrate
	})

	return audioFormats[0].URL, status, nil
}

// playViaYtDlp uses yt-dlp to extract and stream audio to the system player.
// This is the fallback when the InnerTube approach fails due to bot detection,
// since yt-dlp handles YouTube's authentication challenges internally.
// Uses the system yt-dlp if available; otherwise falls back to the embedded binary.
func playViaYtDlp(ctx context.Context, videoID string) error {
	ytdlpPath, err := ytdlp.Executable()
	if err != nil {
		return fmt.Errorf("yt-dlp unavailable: %w", err)
	}

	ytCmd := exec.CommandContext(ctx, ytdlpPath,
		"--no-playlist",
		"-f", "bestaudio[ext=m4a]/bestaudio",
		"-o", "-",
		"--quiet",
		"https://www.youtube.com/watch?v="+videoID,
	)
	ytCmd.Stderr = os.Stderr

	pipe, err := ytCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("create yt-dlp pipe: %w", err)
	}
	if err := ytCmd.Start(); err != nil {
		return fmt.Errorf("start yt-dlp: %w", err)
	}

	playerErr := pipeToPlayer(ctx, pipe)
	_ = ytCmd.Wait()
	return playerErr
}

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

/** Play streams audio for the given video ID.
 *  Primary: android_vr InnerTube client with random visitorData (retried up to maxBotRetries).
 *  Fallback: yt-dlp subprocess (when installed) for videos that require YouTube sign-in. */
func Play(ctx context.Context, videoID string, timeout time.Duration) error {
	playCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	streamURL, err := resolveVRAudioURL(playCtx, videoID)
	if err != nil {
		if !errors.Is(err, errLoginRequired) {
			return err
		}
		// InnerTube approach exhausted; fall back to yt-dlp if available.
		return playViaYtDlp(playCtx, videoID)
	}

	req, err := http.NewRequestWithContext(playCtx, http.MethodGet, streamURL, nil)
	if err != nil {
		return fmt.Errorf("create stream request: %w", err)
	}
	req.Header.Set("User-Agent", vrUserAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("fetch stream: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("stream returned %d", resp.StatusCode)
	}

	return pipeToPlayer(playCtx, resp.Body)
}

// pipeToPlayer starts the first available audio player with stdin as audio source.
func pipeToPlayer(ctx context.Context, r io.Reader) error {
	type player struct {
		bin  string
		args []string
	}

	players := []player{
		{"ffplay", []string{"-autoexit", "-nodisp", "-loglevel", "error", "-i", "pipe:0"}},
		{"mpv", []string{"--no-video", "-"}},
	}

	if runtime.GOOS == "darwin" {
		players = append(players, player{"afplay", []string{"-"}})
	}

	for _, p := range players {
		bin, err := exec.LookPath(p.bin)
		if err != nil {
			continue
		}
		cmd := exec.CommandContext(ctx, bin, p.args...)
		cmd.Stdin = r
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	return fmt.Errorf("no audio player found: install ffplay or mpv")
}
