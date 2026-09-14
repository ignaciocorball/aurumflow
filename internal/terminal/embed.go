package terminal

import (
	"embed"
	"io/fs"
)

//go:embed all:webdist
var webDist embed.FS

func WebFS() fs.FS {
	return webDist
}
