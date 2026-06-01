// Package migrations embeds the goose SQL migration files so they ship inside
// the compiled binary and can be applied at startup without external files.
package migrations

import "embed"

// FS holds every .sql migration in this directory.
//
//go:embed *.sql
var FS embed.FS
