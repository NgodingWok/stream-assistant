package ytdlp

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
)

var (
	ffmpegOnce   sync.Once
	cachedFFmpeg string
	cachedFFmpegErr error
)

// FFmpegExecutable returns the path to a runnable ffmpeg binary.
// System ffmpeg in PATH takes priority over the embedded binary, allowing users
// to override with a different version without rebuilding.
// On Linux and macOS, only PATH is checked (no embedded binary is provided).
// On Windows, if ffmpeg is absent from PATH and a binary was embedded with
// -tags embed_ffmpeg, the embedded binary is extracted to the user cache directory.
func FFmpegExecutable() (string, error) {
	if path, err := exec.LookPath("ffmpeg"); err == nil {
		return path, nil
	}
	if len(EmbeddedFFmpeg) == 0 {
		return "", fmt.Errorf("ffmpeg not found in PATH and no binary was embedded; " +
			"install ffmpeg or rebuild with: go build -tags embed_ffmpeg")
	}
	ffmpegOnce.Do(func() {
		cachedFFmpeg, cachedFFmpegErr = extractFFmpeg()
	})
	return cachedFFmpeg, cachedFFmpegErr
}

// extractFFmpeg writes EmbeddedFFmpeg to the user cache directory and marks it
// executable. A short SHA-256 prefix in the filename ensures a fresh binary is
// written whenever the embedded content changes.
func extractFFmpeg() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = os.TempDir()
	}

	dir := filepath.Join(cacheDir, "stream-assistant")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("create cache dir: %w", err)
	}

	sum := sha256.Sum256(EmbeddedFFmpeg)
	name := "ffmpeg-" + hex.EncodeToString(sum[:8])
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	dest := filepath.Join(dir, name)

	if _, statErr := os.Stat(dest); statErr == nil {
		return dest, nil // already extracted from a previous run
	}

	if err := os.WriteFile(dest, EmbeddedFFmpeg, 0755); err != nil {
		return "", fmt.Errorf("extract embedded ffmpeg to %s: %w", dest, err)
	}
	return dest, nil
}
