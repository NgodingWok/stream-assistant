# stream-assistant

An interactive stream assistant built with [Go](https://go.dev/). Available as a **CLI tool** and a **GUI application** (built with [Fyne](https://fyne.io/)). It connects to TikTok live streams using [gotiktoklive](https://github.com/steampoweredtaco/gotiktoklive), reads chat messages aloud via Google text-to-speech, and plays YouTube audio on demand through chat commands.


## Features

- Real-time TikTok live stream monitoring with chat event processing
- Text-to-speech playback for chat messages using [htgo-tts](https://github.com/hegedustibor/htgo-tts) with configurable language
- YouTube audio playback via `!play <query>` chat command; primary path uses the android_vr InnerTube client (requires `ffplay`, `mpv`, or `afplay`); yt-dlp fallback uses native go-mp3 + oto/v2 playback (requires `ffmpeg` for yt-dlp transcoding); sending a new `!play` while audio is playing cancels the current track and starts the new one
- User join and viewer count logging
- Graceful shutdown via SIGINT/SIGTERM (CLI) or Stop button (GUI)
- Runtime configuration through environment variables (CLI) or settings form (GUI)
- GUI built with [Fyne](https://fyne.io/) — live log view, viewer count, status bar

## Prerequisites

- Go 1.25+
- For YouTube InnerTube playback (primary path), one of:
  - `ffplay` — part of [FFmpeg](https://ffmpeg.org/) (recommended)
  - `mpv` — cross-platform media player
  - `afplay` — built-in on macOS
- For the yt-dlp fallback path: `ffmpeg` in PATH (used by yt-dlp to transcode to MP3; playback itself is native via go-mp3 + oto/v2 — no extra player needed)
  - On Windows this can be bundled into the binary via `--embed-ffmpeg` (see below)

## Installation

```bash
git clone https://github.com/NgodingWok/stream-assistant.git
cd stream-assistant
```

**Standard build** (requires `yt-dlp` in PATH at runtime):
```bash
make build
# or
bash scripts/build.sh
```

**Embedded build** (bundles yt-dlp into the binary, no runtime dependency):
```bash
make build-embedded
# or
bash scripts/build.sh --embedded
```

**Embedded build with FFmpeg for Windows** (bundles both yt-dlp and ffmpeg into the binary):
```bash
# 1. fetch the binaries first (only needed once)
bash scripts/fetch-ytdlp.sh
bash scripts/fetch-ffmpeg.sh    # downloads ffmpeg.exe from BtbN builds

# 2. build
bash scripts/build.sh --embedded --embed-ffmpeg
```

**GUI build**:
```bash
make build-gui
# or
bash scripts/build.sh --gui

# GUI + embedded yt-dlp:
make build-gui-embedded
# or
bash scripts/build.sh --gui --embedded
```

See `scripts/build.sh --help` for all build options (`--os`, `--arch`, `--version`, `--race`).

## Usage

### CLI

```bash
./bin/stream-assistant <tiktok-username>
```

### GUI

```bash
./bin/stream-assistant-gui
```

Fill in the settings form (username, TTS language, chat delay, etc.) and click **Start**.

### Environment Variables

| Variable | Default | Description |
|---|---|---|
| `CHAT_DELAY_MS` | `10000` | Ignore chat messages older than this threshold (ms) |
| `PLAY_TIMEOUT_MIN` | `15` | Maximum duration for YouTube playback (minutes) |
| `TTS_LANGUAGE` | `id` | Language code for text-to-speech (e.g., `en`, `id`, `es`) |
| `TTS_FOLDER` | `.tmp` | Directory for generated TTS and YouTube audio cache files |

### Chat Commands

- `!play <query>` — search YouTube and stream audio playback
- Any other message — read aloud via text-to-speech

## Development

### Project Structure

```
cmd/stream-assistant/   CLI entry point
main.go                 GUI entry point (Fyne)
internal/config/        Configuration loading from environment variables
internal/handler/       TikTok event processing and dispatch
internal/tts/           Text-to-speech wrapper
internal/youtube/       YouTube search and audio streaming
third_party/            Embedded yt-dlp and ffmpeg binary package (build tags: embed_ytdlp, embed_ffmpeg)
test/integration/       Integration tests (build tag: integration)
scripts/                Build, test, fetch-ytdlp, and fetch-ffmpeg shell/bat scripts
```

### Running Tests

```bash
# unit tests
make test
# or: bash scripts/test.sh --unit

# integration tests (requires network and external tools)
make test-integration
# or: bash scripts/test.sh --integration

# both unit and integration
make test-all

# with coverage report
bash scripts/test.sh --unit --coverage
```

See `scripts/test.sh --help` for all options (`--race`, `--run`, `--timeout`, `--verbose`).

## License

MIT License — see [LICENSE](LICENSE).

