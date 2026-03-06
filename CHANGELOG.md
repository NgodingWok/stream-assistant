# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

* CLI entry point for monitoring TikTok live streams via `cmd/stream-assistant`
* TikTok live stream connection using `gotiktoklive` client
* Real-time chat event processing with configurable stale message filtering
* Text-to-speech playback for chat messages using `htgo-tts` with configurable language support
* YouTube audio playback via `!play <query>` chat command using `yt-dlp` and `ffplay`
* YouTube search by scraping video IDs with regex-based extraction
* Direct audio streaming from YouTube without full download
* Non-blocking YouTube playback using goroutines
* Runtime configuration via environment variables (`CHAT_DELAY_MS`, `PLAY_TIMEOUT_MIN`, `TTS_LANGUAGE`, `TTS_FOLDER`)
* Graceful shutdown handling via SIGINT/SIGTERM signals with context cancellation
* User join/leave and viewer count event logging
* Unit tests for configuration loading, handler staleness logic, and YouTube ID extraction
* Integration tests for TikTok connection, TTS playback, and YouTube search/play
