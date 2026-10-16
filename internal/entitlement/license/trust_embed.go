package license

import "embed"

//go:embed vendorkeys/*.pub
var vendorKeyFS embed.FS
