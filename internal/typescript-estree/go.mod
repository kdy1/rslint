module github.com/web-infra-dev/rslint/internal/typescript-estree

go 1.25.0

require (
	github.com/Masterminds/semver/v3 v3.3.1
	github.com/microsoft/typescript-go/shim/ast v0.0.0
)

require (
	github.com/go-json-experiment/json v0.0.0-20250811204210-4789234c3ea1 // indirect
	github.com/microsoft/typescript-go v0.0.0-20250829050502-5d1d69a77a4c // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/text v0.28.0 // indirect
)

replace (
	github.com/microsoft/typescript-go/shim/ast => ../../shim/ast
	github.com/microsoft/typescript-go/shim/core => ../../shim/core
	github.com/microsoft/typescript-go/shim/scanner => ../../shim/scanner
)
