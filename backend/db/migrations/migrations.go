package migrations

import "embed"

// Files contains the SQL migration files bundled with the application.
//
//go:embed *.sql
var Files embed.FS
