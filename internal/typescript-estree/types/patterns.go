// Package types provides ESTree-compliant type definitions for TypeScript AST nodes.
package types

// ArrayPattern represents an array destructuring pattern (ES2015).
type ArrayPattern struct {
	BaseNode
	Elements []Pattern `json:"elements"` // Can contain nil for holes
}

func (a *ArrayPattern) patternNode() {}

// ObjectPattern represents an object destructuring pattern (ES2015).
type ObjectPattern struct {
	BaseNode
	Properties []Node `json:"properties"` // AssignmentProperty or RestElement
}

func (o *ObjectPattern) patternNode() {}

// AssignmentPattern represents an assignment pattern with a default value (ES2015).
type AssignmentPattern struct {
	BaseNode
	Left  Pattern    `json:"left"`
	Right Expression `json:"right"`
}

func (a *AssignmentPattern) patternNode() {}

// RestElement represents a rest element in destructuring (...pattern) (ES2015).
type RestElement struct {
	BaseNode
	Argument Pattern `json:"argument"`
}

func (r *RestElement) patternNode() {}

// AssignmentProperty represents a property in an object pattern.
type AssignmentProperty struct {
	BaseNode
	Key       Expression `json:"key"`
	Value     Pattern    `json:"value"`
	Kind      string     `json:"kind"` // Always "init"
	Method    bool       `json:"method"`
	Shorthand bool       `json:"shorthand"`
	Computed  bool       `json:"computed"`
}
