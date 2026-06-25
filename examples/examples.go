package examples

import "embed"

//go:embed *.conf *.toml *.yaml *.json *.hcl
var Files embed.FS
