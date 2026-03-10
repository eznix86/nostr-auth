package nostrauth

import "embed"

//go:embed all:public all:resources/views/app.html
var AssetsFS embed.FS
