//go:build !embed_ytdlp

package ytdlp

// EmbeddedBinary holds the yt-dlp binary embedded at build time.
// Empty in the default build; populate by building with: -tags embed_ytdlp
var EmbeddedBinary []byte
