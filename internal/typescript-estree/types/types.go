// Package types provides ESTree-compliant type definitions for TypeScript AST nodes.
// This package defines the core data structures used to represent JavaScript/TypeScript
// syntax trees in the ESTree format, which is the standard AST format used by ESLint
// and other JavaScript tooling.
package types

// Node represents the base interface for all ESTree nodes.
// Every AST node must implement this interface.
type Node interface {
	// Type returns the type of the node (e.g., "Program", "Identifier", etc.)
	Type() string

	// Loc returns the source location information for this node
	Loc() *SourceLocation

	// Range returns the start and end positions of this node in the source
	Range() [2]int
}

// SourceLocation represents the location of a node in the source code.
type SourceLocation struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Position represents a single position in the source code.
type Position struct {
	Line   int `json:"line"`   // 1-indexed line number
	Column int `json:"column"` // 0-indexed column number
}

// BaseNode provides a basic implementation of common node fields.
// Concrete node types should embed this struct.
type BaseNode struct {
	NodeType string          `json:"type"`
	Location *SourceLocation `json:"loc,omitempty"`
	Span     [2]int          `json:"range,omitempty"`
}

// Type implements the Node interface.
func (n *BaseNode) Type() string {
	return n.NodeType
}

// Loc implements the Node interface.
func (n *BaseNode) Loc() *SourceLocation {
	return n.Location
}

// Range implements the Node interface.
func (n *BaseNode) Range() [2]int {
	return n.Span
}

// Program represents the top-level program node.
type Program struct {
	BaseNode
	SourceType string `json:"sourceType"` // "script" or "module"
	Body       []Node `json:"body"`
}

// Identifier represents an identifier node.
type Identifier struct {
	BaseNode
	Name string `json:"name"`
}
