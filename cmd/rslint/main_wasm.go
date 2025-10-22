//go:build (wasm || js)
// +build wasm js

package main

import (
	"fmt"
	"os"
)

// runLSP is not supported in WASM builds
func runLSP(args []string) int {
	fmt.Fprintln(os.Stderr, "LSP mode is not supported in WASM builds")
	return 1
}
