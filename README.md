# stream-assistant

An interactive stream assistant and CLI tool built with Node.js. It integrates [with tiktok-live-connector](https://github.com/zerodytrash/TikTok-Live-Connector) to listening the chat events and uses gTTS (text-to-speech) for audio feedback.

## Features

- Command-line interface executable (`stream-assistant`)
- Language selection via `-l`/`--lang` option
- Handlers for default behaviour, help, and version commands
- Text-to-speech utility (`src/utils/tts.js`) using `gtts` and `sound-play`
- Easily extendable with new command handlers under `src/handler/`

## Installation

```bash
# clone and install locally
git clone https://github.com/NgodingWok/stream-assistant.git
cd stream-assistant
npm install

# or install globally after publishing
npm i -g stream-assistant
```

## Usage

```bash
# run with language code (e.g., en, ar, es)
stream-assistant -l en
```

Available commands:

- `-help` — show usage information
- `-version` — display the package version
- any other input triggers the default handler

## Development

- Handlers live in `src/handler/`; each exports a function receiving `(args)`.
- `src/index.js` is the CLI entry point and dispatches to handlers.
- `utils/tts.js` provides `speak(text, lang)` returning a promise.

To add a handler, create a new file and update the `handlers` object in `index.js`.

## License

MIT License — see [LICENSE](LICENSE).

