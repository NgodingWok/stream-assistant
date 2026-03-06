# stream-assistant

An interactive stream assistant and CLI tool built with [Go](https://go.dev/). It connects to TikTok live streams using [gotiktoklive](https://github.com/steampoweredtaco/gotiktoklive), reads chat messages aloud via google text-to-speech, and plays YouTube audio on demand through chat commands.


## Features

- Real-time TikTok live stream monitoring with chat event processing
- Text-to-speech playback for chat messages using [htgo-tts](https://github.com/hegedustibor/htgo-tts) with configurable language
- YouTube audio playback via `!play <query>` chat command (requires `yt-dlp` and `ffplay`)
- User join and viewer count logging
- Graceful shutdown via SIGINT/SIGTERM
- Runtime configuration through environment variables

## Prerequisites

- Go 1.25+
- `yt-dlp` — for extracting YouTube audio stream URLs
- `ffplay` — for audio playback (part of FFmpeg)

## Installation

```bash
git clone https://github.com/NgodingWok/stream-assistant.git
cd stream-assistant
go build -o stream-assistant ./cmd/stream-assistant
```

## Usage

```bash
stream-assistant <tiktok-username>
```

### Environment Variables

| Variable | Default | Description |
|---|---|---|
| `CHAT_DELAY_MS` | `10000` | Ignore chat messages older than this threshold (ms) |
| `PLAY_TIMEOUT_MIN` | `15` | Maximum duration for YouTube playback (minutes) |
| `TTS_LANGUAGE` | `id` | Language code for text-to-speech (e.g., `en`, `id`, `es`) |
| `TTS_FOLDER` | `.tmp` | Directory for generated TTS audio files |

### Chat Commands

- `!play <query>` — search YouTube and stream audio playback
- Any other message — read aloud via text-to-speech

## Development

### Project Structure

```
cmd/stream-assistant/   CLI entry point
internal/config/        Configuration loading from environment variables
internal/handler/       TikTok event processing and dispatch
internal/tts/           Text-to-speech wrapper
internal/youtube/       YouTube search and audio streaming
test/integration/       Integration tests (build tag: integration)
```

### Running Tests

```bash
# unit tests
go test ./...

# integration tests (requires network and external tools)
go test -tags=integration ./test/integration/...
```

## License

MIT License — see [LICENSE](LICENSE).

