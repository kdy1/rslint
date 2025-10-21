//go:build js

package main

import "log"

func runLSP(args []string) int {
	log.Println("LSP mode is not supported in WASM builds")
	return 1
}
