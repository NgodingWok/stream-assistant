.PHONY: build build-embedded fetch-ytdlp test test-integration test-all clean

## build: compile without embedded yt-dlp (requires yt-dlp in PATH at runtime)
build:
	@bash scripts/build.sh

## build-embedded: compile with yt-dlp binaries embedded in the output binary
build-embedded:
	@bash scripts/build.sh --embedded

## fetch-ytdlp: download yt-dlp binaries for all supported platforms into third_party/bin/
fetch-ytdlp:
	@bash scripts/fetch-ytdlp.sh

## test: run unit tests
test:
	@bash scripts/test.sh --unit

## test-integration: run integration tests (requires network access)
test-integration:
	@bash scripts/test.sh --integration

## test-all: run unit and integration tests
test-all:
	@bash scripts/test.sh --all

## clean: remove build artifacts and coverage reports
clean:
	rm -rf bin/ coverage/
