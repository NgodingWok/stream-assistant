//go:build embed_ytdlp && linux && amd64

package ytdlp

import _ "embed"

//go:embed bin/yt-dlp_linux
var EmbeddedBinary []byte
