// Package types provides ESTree-compliant type definitions for TypeScript AST nodes.
package types

// Position represents a single position in the source code.
// Positions are 1-indexed for line numbers and 0-indexed for column numbers,
// following the ESTree specification.
type Position struct {
	Line   int `json:"line"`   // 1-indexed line number
	Column int `json:"column"` // 0-indexed column number
}

// SourceLocation represents the location of a node in the source code.
// It contains the start and end positions of a node.
type SourceLocation struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Range represents the character range of a node in the source code.
// It is a tuple of two numbers: [start, end], where both are 0-indexed
// character offsets from the start of the file.
type Range [2]int
