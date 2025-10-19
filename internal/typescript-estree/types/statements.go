// Package types provides ESTree-compliant type definitions for TypeScript AST nodes.
package types

// ExpressionStatement represents a statement consisting of a single expression.
type ExpressionStatement struct {
	BaseNode
	Expression Expression `json:"expression"`
	Directive  string     `json:"directive,omitempty"` // For "use strict" etc.
}

func (e *ExpressionStatement) statementNode() {}

// BlockStatement represents a block statement (a sequence of statements surrounded by braces).
type BlockStatement struct {
	BaseNode
	Body []Statement `json:"body"`
}

func (b *BlockStatement) statementNode() {}

// EmptyStatement represents an empty statement (a solitary semicolon).
type EmptyStatement struct {
	BaseNode
}

func (e *EmptyStatement) statementNode() {}

// DebuggerStatement represents a debugger statement.
type DebuggerStatement struct {
	BaseNode
}

func (d *DebuggerStatement) statementNode() {}

// WithStatement represents a with statement.
type WithStatement struct {
	BaseNode
	Object Expression `json:"object"`
	Body   Statement  `json:"body"`
}

func (w *WithStatement) statementNode() {}

// ReturnStatement represents a return statement.
type ReturnStatement struct {
	BaseNode
	Argument *Expression `json:"argument"`
}

func (r *ReturnStatement) statementNode() {}

// LabeledStatement represents a labeled statement.
type LabeledStatement struct {
	BaseNode
	Label *Identifier `json:"label"`
	Body  Statement   `json:"body"`
}

func (l *LabeledStatement) statementNode() {}

// BreakStatement represents a break statement.
type BreakStatement struct {
	BaseNode
	Label *Identifier `json:"label"`
}

func (b *BreakStatement) statementNode() {}

// ContinueStatement represents a continue statement.
type ContinueStatement struct {
	BaseNode
	Label *Identifier `json:"label"`
}

func (c *ContinueStatement) statementNode() {}

// IfStatement represents an if statement.
type IfStatement struct {
	BaseNode
	Test       Expression `json:"test"`
	Consequent Statement  `json:"consequent"`
	Alternate  *Statement `json:"alternate"`
}

func (i *IfStatement) statementNode() {}

// SwitchStatement represents a switch statement.
type SwitchStatement struct {
	BaseNode
	Discriminant Expression   `json:"discriminant"`
	Cases        []SwitchCase `json:"cases"`
}

func (s *SwitchStatement) statementNode() {}

// SwitchCase represents a case (or default) clause in a switch statement.
type SwitchCase struct {
	BaseNode
	Test       *Expression `json:"test"` // nil for default case
	Consequent []Statement `json:"consequent"`
}

// ThrowStatement represents a throw statement.
type ThrowStatement struct {
	BaseNode
	Argument Expression `json:"argument"`
}

func (t *ThrowStatement) statementNode() {}

// TryStatement represents a try statement.
type TryStatement struct {
	BaseNode
	Block     *BlockStatement `json:"block"`
	Handler   *CatchClause    `json:"handler"`
	Finalizer *BlockStatement `json:"finalizer"`
}

func (t *TryStatement) statementNode() {}

// CatchClause represents a catch clause in a try statement.
type CatchClause struct {
	BaseNode
	Param *Pattern        `json:"param"` // Can be nil in ES2019+
	Body  *BlockStatement `json:"body"`
}

// WhileStatement represents a while statement.
type WhileStatement struct {
	BaseNode
	Test Expression `json:"test"`
	Body Statement  `json:"body"`
}

func (w *WhileStatement) statementNode() {}

// DoWhileStatement represents a do-while statement.
type DoWhileStatement struct {
	BaseNode
	Body Statement  `json:"body"`
	Test Expression `json:"test"`
}

func (d *DoWhileStatement) statementNode() {}

// ForStatement represents a for statement.
type ForStatement struct {
	BaseNode
	Init   Node       `json:"init"` // VariableDeclaration or Expression or nil
	Test   *Expression `json:"test"`
	Update *Expression `json:"update"`
	Body   Statement  `json:"body"`
}

func (f *ForStatement) statementNode() {}

// ForInStatement represents a for-in statement.
type ForInStatement struct {
	BaseNode
	Left  Node       `json:"left"`  // VariableDeclaration or Pattern
	Right Expression `json:"right"`
	Body  Statement  `json:"body"`
}

func (f *ForInStatement) statementNode() {}

// ForOfStatement represents a for-of statement (ES2015).
type ForOfStatement struct {
	BaseNode
	Left  Node       `json:"left"`  // VariableDeclaration or Pattern
	Right Expression `json:"right"`
	Body  Statement  `json:"body"`
	Await bool       `json:"await"` // for await...of (ES2018)
}

func (f *ForOfStatement) statementNode() {}

// FunctionDeclaration represents a function declaration.
type FunctionDeclaration struct {
	BaseNode
	ID         *Identifier     `json:"id"`
	Params     []Pattern       `json:"params"`
	Body       *BlockStatement `json:"body"`
	Generator  bool            `json:"generator"`
	Async      bool            `json:"async"`
	Expression bool            `json:"expression"`
}

func (f *FunctionDeclaration) statementNode()    {}
func (f *FunctionDeclaration) declarationNode() {}

// VariableDeclaration represents a variable declaration.
type VariableDeclaration struct {
	BaseNode
	Declarations []VariableDeclarator `json:"declarations"`
	Kind         string               `json:"kind"` // "var", "let", or "const"
}

func (v *VariableDeclaration) statementNode()    {}
func (v *VariableDeclaration) declarationNode() {}

// VariableDeclarator represents a variable declarator.
type VariableDeclarator struct {
	BaseNode
	ID   Pattern     `json:"id"`
	Init *Expression `json:"init"`
}

// ClassDeclaration represents a class declaration (ES2015).
type ClassDeclaration struct {
	BaseNode
	ID         *Identifier `json:"id"`
	SuperClass *Expression `json:"superClass"`
	Body       *ClassBody  `json:"body"`
}

func (c *ClassDeclaration) statementNode()    {}
func (c *ClassDeclaration) declarationNode() {}

// ImportDeclaration represents an import declaration (ES2015).
type ImportDeclaration struct {
	BaseNode
	Specifiers []Node     `json:"specifiers"` // ImportSpecifier, ImportDefaultSpecifier, or ImportNamespaceSpecifier
	Source     *SimpleLiteral `json:"source"`
}

func (i *ImportDeclaration) moduleDeclarationNode() {}

// ImportSpecifier represents a named import specifier.
type ImportSpecifier struct {
	BaseNode
	Imported *Identifier `json:"imported"`
	Local    *Identifier `json:"local"`
}

// ImportDefaultSpecifier represents a default import specifier.
type ImportDefaultSpecifier struct {
	BaseNode
	Local *Identifier `json:"local"`
}

// ImportNamespaceSpecifier represents a namespace import specifier.
type ImportNamespaceSpecifier struct {
	BaseNode
	Local *Identifier `json:"local"`
}

// ExportNamedDeclaration represents a named export declaration (ES2015).
type ExportNamedDeclaration struct {
	BaseNode
	Declaration Declaration    `json:"declaration"`
	Specifiers  []ExportSpecifier `json:"specifiers"`
	Source      *SimpleLiteral `json:"source"`
}

func (e *ExportNamedDeclaration) moduleDeclarationNode() {}

// ExportSpecifier represents an export specifier.
type ExportSpecifier struct {
	BaseNode
	Exported *Identifier `json:"exported"`
	Local    *Identifier `json:"local"`
}

// ExportDefaultDeclaration represents a default export declaration (ES2015).
type ExportDefaultDeclaration struct {
	BaseNode
	Declaration Node `json:"declaration"` // Declaration or Expression
}

func (e *ExportDefaultDeclaration) moduleDeclarationNode() {}

// ExportAllDeclaration represents an export all declaration (ES2015).
type ExportAllDeclaration struct {
	BaseNode
	Source   *SimpleLiteral `json:"source"`
	Exported *Identifier    `json:"exported"` // ES2020
}

func (e *ExportAllDeclaration) moduleDeclarationNode() {}
