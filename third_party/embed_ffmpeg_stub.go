//go:build !embed_ffmpeg

package ytdlp

// EmbeddedFFmpeg holds the ffmpeg binary embedded at build time.
// Empty in the default build; populate by building with: -tags embed_ffmpeg
var EmbeddedFFmpeg []byte
