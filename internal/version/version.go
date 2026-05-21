package version

// Version is the kitsune release version. It defaults to "dev" for local
// builds and is overridden via ldflags at release time by goreleaser.
var Version = "dev"

// Schema is the JSON output schema version. Bump only on breaking changes
// to the shape of the report emitted by `kitsune --json`.
const Schema = "1"
