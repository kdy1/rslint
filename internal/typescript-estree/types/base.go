// Package types provides ESTree-compliant type definitions for TypeScript AST nodes.
package types

// Node represents the base interface for all ESTree nodes.
// Every AST node must implement this interface.
type Node interface {
	// Type returns the type of the node (e.g., "Program", "Identifier", etc.)
	Type() string

	// Loc returns the source location information for this node
	Loc() *SourceLocation

	// Range returns the start and end positions of this node in the source
	GetRange() Range
}

// BaseNode provides a basic implementation of common node fields.
// Concrete node types should embed this struct.
type BaseNode struct {
	NodeType string          `json:"type"`
	Location *SourceLocation `json:"loc,omitempty"`
	Span     Range           `json:"range,omitempty"`
}

// Type implements the Node interface.
func (n *BaseNode) Type() string {
	return n.NodeType
}

// Loc implements the Node interface.
func (n *BaseNode) Loc() *SourceLocation {
	return n.Location
}

// GetRange implements the Node interface.
func (n *BaseNode) GetRange() Range {
	return n.Span
}

// Statement is a marker interface for all statement nodes.
type Statement interface {
	Node
	statementNode()
}

// Expression is a marker interface for all expression nodes.
type Expression interface {
	Node
	expressionNode()
}

// Declaration is a marker interface for all declaration nodes.
type Declaration interface {
	Statement
	declarationNode()
}

// Pattern is a marker interface for all pattern nodes (used in destructuring).
type Pattern interface {
	Node
	patternNode()
}

// ModuleDeclaration is a marker interface for all module declaration nodes.
type ModuleDeclaration interface {
	Node
	moduleDeclarationNode()
}

// Program represents the top-level program node.
// It is the root of every ESTree-compliant AST.
type Program struct {
	BaseNode
	SourceType string      `json:"sourceType"` // "script" or "module"
	Body       []Statement `json:"body"`
	Comments   []Comment   `json:"comments,omitempty"`
	Tokens     []Token     `json:"tokens,omitempty"`
}

// Identifier represents an identifier.
type Identifier struct {
	BaseNode
	Name string `json:"name"`
}

func (i *Identifier) expressionNode() {}
func (i *Identifier) patternNode()    {}

// Literal represents a literal value.
type Literal interface {
	Expression
	literalNode()
}

// SimpleLiteral represents a simple literal (string, number, boolean, null).
type SimpleLiteral struct {
	BaseNode
	Value interface{} `json:"value"` // string, float64, bool, or nil
	Raw   string      `json:"raw"`
}

func (l *SimpleLiteral) expressionNode() {}
func (l *SimpleLiteral) literalNode()    {}

// RegExpLiteral represents a regular expression literal.
type RegExpLiteral struct {
	BaseNode
	Value *RegExpValue `json:"value"`
	Raw   string       `json:"raw"`
	Regex *RegExpValue `json:"regex"`
}

func (r *RegExpLiteral) expressionNode() {}
func (r *RegExpLiteral) literalNode()    {}

// RegExpValue represents the value of a regular expression.
type RegExpValue struct {
	Pattern string `json:"pattern"`
	Flags   string `json:"flags"`
}

// BigIntLiteral represents a BigInt literal (ES2020).
type BigIntLiteral struct {
	BaseNode
	Value interface{} `json:"value"` // Can be bigint or nil
	Raw   string      `json:"raw"`
	Bigint string     `json:"bigint"`
}

func (b *BigIntLiteral) expressionNode() {}
func (b *BigIntLiteral) literalNode()    {}

// Comment represents a comment in the source code.
type Comment struct {
	Type  string         `json:"type"`  // "Line" or "Block"
	Value string         `json:"value"` // The comment text
	Loc   *SourceLocation `json:"loc,omitempty"`
	Range Range          `json:"range,omitempty"`
}

// Token represents a token in the source code.
type Token struct {
	Type  string         `json:"type"`
	Value string         `json:"value"`
	Loc   *SourceLocation `json:"loc,omitempty"`
	Range Range          `json:"range,omitempty"`
}

// PrivateIdentifier represents a private identifier (# prefix).
type PrivateIdentifier struct {
	BaseNode
	Name string `json:"name"`
}

func (p *PrivateIdentifier) expressionNode() {}
func (p *PrivateIdentifier) patternNode()    {}
