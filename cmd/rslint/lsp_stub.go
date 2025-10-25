//go:build !lsp
// +build !lsp

package main

import (
	"fmt"
	"os"
)

func runLSP(args []string) int {
	fmt.Fprintln(os.Stderr, "LSP support is disabled in this build. Please rebuild with -tags=lsp to enable LSP functionality.")
	return 1
}
