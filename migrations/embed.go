package migrations

import _ "embed"

//go:embed 001_initial_schema.sql
var schema001 string

//go:embed 002_additions.sql
var schema002 string

// Schema is the full ordered migration run on every startup.
// Each statement is idempotent so re-runs are safe.
var Schema = schema001 + "\n" + schema002
