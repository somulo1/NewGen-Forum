// Package seed carries the demo data (exported from the frontend's mock data) and the tones, inside the binary.
package seed

import "embed"

//go:embed data/*.json tones.json
var FS embed.FS
