//go:build wasm
// +build wasm

package main

import (
	"fmt"
	"os"
)

// runLSP is a stub for WASM builds where LSP is not supported
func runLSP(args []string) int {
	fmt.Fprintln(os.Stderr, "LSP mode is not supported in WASM builds")
	return 1
}
