// Package migrations embeds the .sql migration files so they ship inside the
// compiled binary and can be applied at startup without external files.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
