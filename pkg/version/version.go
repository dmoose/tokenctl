// Package version is the one place tokenctl's version is stated.
//
// It used to be stated twice: a `main.version` var set by ldflags, and a
// hand-maintained `generators.TokenctlVersion` const that the catalog
// stamped into every export. The const drifted — it read "1.2.0" against
// binaries built from anything — and a version string that is wrong is
// worse than one that is absent, because a consumer reconciling catalog
// shapes across releases trusts it.
package version

import "runtime/debug"

// Value is the build-time version, injected via ldflags:
//
//	go build -ldflags "-X github.com/dmoose/tokenctl/pkg/version.Value=v1.2.0"
//
// Left at "dev" it is not used; String falls back to what the binary
// itself knows.
var Value = "dev"

// Commit and BuildTime are the other two ldflags slots, kept for the
// `version` subcommand. They are not stamped into generated artifacts —
// a catalog carrying a build timestamp is exactly the reproducibility
// problem this package exists next to.
var (
	Commit    = "unknown"
	BuildTime = "unknown"
)

// String reports the version of the running binary.
//
// An ldflags-injected Value wins. Otherwise the module's own build info
// answers: a `go install`ed binary knows its module version, and one
// built from a working tree knows its VCS revision. Only a binary that
// knows neither reports "dev".
func String() string {
	if Value != "" && Value != "dev" {
		return Value
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}
	if v := info.Main.Version; v != "" && v != "(devel)" {
		return v
	}
	var rev string
	var dirty bool
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			rev = s.Value
		case "vcs.modified":
			dirty = s.Value == "true"
		}
	}
	if rev == "" {
		return "dev"
	}
	if len(rev) > 12 {
		rev = rev[:12]
	}
	if dirty {
		return "dev+" + rev + ".dirty"
	}
	return "dev+" + rev
}
