package assets

import (
	"embed"
	"io/fs"
)

//go:embed * **/*
var staticAssets embed.FS

func StaticAssets() fs.FS {
	return staticAssets
}
