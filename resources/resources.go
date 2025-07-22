package resources

import "embed"

var (
	//go:embed gen/ui
	UIFs embed.FS
)
