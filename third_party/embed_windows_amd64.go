//go:build embed_ytdlp && windows && amd64

package ytdlp

import _ "embed"

//go:embed bin/yt-dlp.exe
var EmbeddedBinary []byte
