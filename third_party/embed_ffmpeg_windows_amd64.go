//go:build embed_ffmpeg && windows && amd64

package ytdlp

import _ "embed"

//go:embed bin/ffmpeg.exe
var EmbeddedFFmpeg []byte
