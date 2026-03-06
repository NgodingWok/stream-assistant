// Package ytdlp provides access to a yt-dlp executable, either from the system
// PATH or from a binary embedded at build time with -tags embed_ytdlp.
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
	extractOnce   sync.Once
	cachedPath    string
	cachedPathErr error
)

// Executable returns the path to a runnable yt-dlp binary.
// System-installed yt-dlp in PATH takes priority over the embedded binary,
// allowing users to override with a newer version without rebuilding.
// If neither is available, an error describing how to fix it is returned.
func Executable() (string, error) {
	if path, err := exec.LookPath("yt-dlp"); err == nil {
		return path, nil
	}
	if len(EmbeddedBinary) == 0 {
		return "", fmt.Errorf("yt-dlp not found in PATH and no binary was embedded; " +
			"install yt-dlp or rebuild with: go build -tags embed_ytdlp")
	}
	extractOnce.Do(func() {
		cachedPath, cachedPathErr = extractBinary()
	})
	return cachedPath, cachedPathErr
}

// extractBinary writes EmbeddedBinary to the user cache directory and marks it
// executable. A short SHA-256 prefix in the filename ensures a fresh binary is
// written whenever the embedded content changes (e.g. after a yt-dlp update).
func extractBinary() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = os.TempDir()
	}

	dir := filepath.Join(cacheDir, "stream-assistant")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("create cache dir: %w", err)
	}

	sum := sha256.Sum256(EmbeddedBinary)
	name := "yt-dlp-" + hex.EncodeToString(sum[:8])
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	dest := filepath.Join(dir, name)

	if _, statErr := os.Stat(dest); statErr == nil {
		return dest, nil // already extracted from a previous run
	}

	if err := os.WriteFile(dest, EmbeddedBinary, 0755); err != nil {
		return "", fmt.Errorf("extract embedded yt-dlp to %s: %w", dest, err)
	}
	return dest, nil
}
