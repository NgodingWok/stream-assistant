//go:build embed_ytdlp && !(linux && (amd64 || arm64)) && !darwin && !(windows && amd64)

package ytdlp

// EmbeddedBinary is empty on platforms not covered by a dedicated embed file.
// yt-dlp must be available in PATH on this platform.
var EmbeddedBinary []byte
