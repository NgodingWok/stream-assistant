//go:build embed_ytdlp && linux && arm64

package ytdlp

import _ "embed"

//go:embed bin/yt-dlp_linux_aarch64
var EmbeddedBinary []byte
