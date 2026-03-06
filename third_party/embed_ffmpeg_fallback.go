//go:build embed_ffmpeg && !(windows && amd64)

package ytdlp

// EmbeddedFFmpeg is empty on platforms without a dedicated embed file.
// ffmpeg must be available in PATH on this platform.
var EmbeddedFFmpeg []byte
