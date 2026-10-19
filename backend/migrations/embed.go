// Package migrations embeds the SQL files so the binary carries exactly the schema it was built with.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
