//go:build embed_ytdlp && darwin

package ytdlp

import _ "embed"

// yt-dlp_macos is a universal binary (x86_64 + arm64) covering both Intel and Apple Silicon.
//
//go:embed bin/yt-dlp_macos
var EmbeddedBinary []byte
