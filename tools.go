//go:build tools

// This file is used to declare tool dependencies for the project, for tools that are only used by the `go generate`
// command. These tools would otherwise be unpinned, since the go.mod file would prune them on `go mod tidy`.

package tools

import (
	_ "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen"
)
