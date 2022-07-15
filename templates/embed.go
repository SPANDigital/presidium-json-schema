package templates

import "embed"

var (
	//go:embed *.gohtml
	Files embed.FS
)
