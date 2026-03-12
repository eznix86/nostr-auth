package nostrauth

import "embed"

//go:embed all:public all:resources/views/app.html resources/images/images.json
var AssetsFS embed.FS
