# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

* GUI application (`main.go`) built with [Fyne v2](https://fyne.io/) — settings form, scrollable monospace log list (capped at 500 entries), live viewer count binding, status bar, and Start/Stop session control
* `--gui` flag in `scripts/build.sh` and `scripts/build.bat` — builds the root GUI package (`main.go`) instead of the CLI; defaults output to `bin/stream-assistant-gui[.exe]`; composable with all existing flags (`--embedded`, `--os`, `--arch`, `--version`, `--race`)
* `build-gui` and `build-gui-embedded` Makefile targets delegating to `scripts/build.sh --gui [--embedded]`
* `scripts/build.sh` and `scripts/build.bat` — cross-platform build scripts with options: `--embedded`, `--embed-ffmpeg`, `--output`, `--os`, `--arch`, `--version`, `--race`, `--help`
* `scripts/test.sh` and `scripts/test.bat` — test runner scripts with options: `--unit`, `--integration`, `--all`, `--coverage`, `--race`, `--run`, `--timeout`, `--verbose`, `--help`
* `scripts/fetch-ytdlp.sh` and `scripts/fetch-ytdlp.bat` — download yt-dlp release binaries for all supported platforms with options: `--version`, `--dir`, `--platform`
* `scripts/fetch-ffmpeg.sh` and `scripts/fetch-ffmpeg.bat` — download the FFmpeg Windows binary (gpl-essentials, statically linked) from BtbN FFmpeg Builds into `third_party/bin/ffmpeg.exe`
* `third_party/` package (`package ytdlp`) with platform-specific `//go:embed` files for bundling yt-dlp into the output binary at build time using `-tags embed_ytdlp`
* `third_party/ffmpeg.go` — `FFmpegExecutable()` function following the same pattern as `Executable()`: system PATH first, then embedded binary extracted to user cache on first use
* `embed_ffmpeg_windows_amd64.go`, `embed_ffmpeg_stub.go`, `embed_ffmpeg_fallback.go` — platform-specific embed files for `embed_ffmpeg` build tag (only Windows amd64 actually embeds; all other platforms use PATH)
* `Makefile` with `build`, `build-embedded`, `build-gui`, `build-gui-embedded`, `test`, `test-integration`, `test-all`, `fetch-ytdlp`, `fetch-ffmpeg`, and `clean` targets delegating to `scripts/`
* Embedded yt-dlp binaries in `third_party/bin/` for Linux amd64, Linux arm64, macOS universal, and Windows amd64
* `ytdlp.Executable()` function that resolves yt-dlp from system PATH (priority) or extracts the embedded binary to the user cache directory on first use
* CLI entry point for monitoring TikTok live streams via `cmd/stream-assistant`
* TikTok live stream connection using `gotiktoklive` client
* Real-time chat event processing with configurable stale message filtering
* Text-to-speech playback for chat messages using `htgo-tts` with configurable language support
* YouTube audio playback via `!play <query>` chat command using a custom android_vr InnerTube client
* YouTube search by scraping video IDs with regex-based extraction
* Cross-platform audio player fallback chain supporting `ffplay`, `mpv`, and `afplay` (macOS)
* Non-blocking YouTube playback using goroutines
* Polling loop that waits for the target user to go live before connecting, retrying every 30 seconds until live or interrupted

### Changed

* `handler.Handler` and `streamApp` now run a single dedicated player goroutine per session (`runPlayer` state machine over a `playCh chan string` channel); a new `!play` command drains any pending request and cancels in-progress playback before starting the new one — the latest request always wins
* Added `--embed-ffmpeg` build flag to `scripts/build.sh` and `scripts/build.bat`: embeds `ffmpeg.exe` (Windows amd64 only) into the binary using build tag `embed_ffmpeg`; composable with `--embedded` for a fully self-contained Windows binary
* `playViaYtDlp` now passes `--ffmpeg-location <path>` to yt-dlp when `FFmpegExecutable()` resolves a path (embedded or system), eliminating the PATH requirement on Windows when using `--embed-ffmpeg`
* Replaced stdout-pipe streaming in `playViaYtDlp` with file-based download: yt-dlp saves to `<tmpDir>/<videoID>.mp3` then plays natively via go-mp3 + oto/v2 — no external audio player required for the yt-dlp fallback path
* Added playback cache: if `<videoID>.mp3` already exists in `tmpDir`, yt-dlp download is skipped and the file is played directly
* `Play()` now accepts a `tmpDir` string parameter (reuses `TTSFolder`) for downloaded audio file storage, matching the TTS pattern; `TTS_FOLDER` now serves as the cache directory for both TTS audio files and downloaded YouTube MP3s
* Added oto/v2 context singleton (`sync.Once`) to comply with oto's single-context-per-process constraint
* Replaced `yt-dlp` external process and `github.com/kkdai/youtube/v2` library with a self-contained android_vr InnerTube client, removing all external dependencies for YouTube stream resolution
* YouTube stream URL resolution now POSTs directly to `youtubei/v1/player` with `ANDROID_VR` client parameters, streams the audio via Go's `net/http` with matching User-Agent, and pipes to the system player stdin
* Runtime configuration via environment variables (`CHAT_DELAY_MS`, `PLAY_TIMEOUT_MIN`, `TTS_LANGUAGE`, `TTS_FOLDER`)
* Graceful shutdown handling via SIGINT/SIGTERM signals with context cancellation
* User join/leave and viewer count event logging
* Unit tests for configuration loading, handler staleness logic, and YouTube ID extraction
* Integration tests for TikTok connection, TTS playback, and YouTube search/play

### Fixed

* Nil pointer dereference on startup when `live.Info` or `live.Info.Owner` is not populated after `TrackUser`; falls back to the configured username as display name
* WebSocket connection error (`unexpected HTTP response status: 200`) when the user is not live; replaced with a pre-flight liveness check via `GetLiveRoomUserInfo` and `IsLive` before calling `TrackUser`
* YouTube stream HTTP 403 errors caused by YouTube CDN blocking `android_sdkless` (`c=ANDROID`) signed URLs; resolved by implementing a minimal InnerTube client using the `ANDROID_VR` (`c=ANDROID_VR`) client identity which CDN accepts without restriction
