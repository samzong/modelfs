package static

import "embed"

//go:embed dist/* dist/assets/*
var Files embed.FS

