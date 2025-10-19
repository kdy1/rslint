// Package types provides ESTree-compliant type definitions for TypeScript AST nodes.
package types

// ThisExpression represents a 'this' expression.
type ThisExpression struct {
	BaseNode
}

func (t *ThisExpression) expressionNode() {}

// ArrayExpression represents an array literal expression.
type ArrayExpression struct {
	BaseNode
	Elements []Expression `json:"elements"` // Can contain nil for holes
}

func (a *ArrayExpression) expressionNode() {}

// ObjectExpression represents an object literal expression.
type ObjectExpression struct {
	BaseNode
	Properties []Property `json:"properties"`
}

func (o *ObjectExpression) expressionNode() {}

// Property represents a property in an object expression.
type Property struct {
	BaseNode
	Key       Expression `json:"key"`
	Value     Expression `json:"value"`
	Kind      string     `json:"kind"` // "init", "get", or "set"
	Method    bool       `json:"method"`
	Shorthand bool       `json:"shorthand"`
	Computed  bool       `json:"computed"`
}

// FunctionExpression represents a function expression.
type FunctionExpression struct {
	BaseNode
	ID         *Identifier     `json:"id"`
	Params     []Pattern       `json:"params"`
	Body       *BlockStatement `json:"body"`
	Generator  bool            `json:"generator"`
	Async      bool            `json:"async"`
	Expression bool            `json:"expression"`
}

func (f *FunctionExpression) expressionNode() {}

// ArrowFunctionExpression represents an arrow function expression (ES2015).
type ArrowFunctionExpression struct {
	BaseNode
	Params     []Pattern `json:"params"`
	Body       Node      `json:"body"` // BlockStatement or Expression
	Generator  bool      `json:"generator"`
	Async      bool      `json:"async"`
	Expression bool      `json:"expression"`
}

func (a *ArrowFunctionExpression) expressionNode() {}

// UnaryExpression represents a unary operation expression.
type UnaryExpression struct {
	BaseNode
	Operator string     `json:"operator"` // "-", "+", "!", "~", "typeof", "void", "delete"
	Prefix   bool       `json:"prefix"`
	Argument Expression `json:"argument"`
}

func (u *UnaryExpression) expressionNode() {}

// UpdateExpression represents an update (increment/decrement) expression.
type UpdateExpression struct {
	BaseNode
	Operator string     `json:"operator"` // "++" or "--"
	Argument Expression `json:"argument"`
	Prefix   bool       `json:"prefix"`
}

func (u *UpdateExpression) expressionNode() {}

// BinaryExpression represents a binary operation expression.
type BinaryExpression struct {
	BaseNode
	Operator string     `json:"operator"` // "==", "!=", "===", "!==", "<", "<=", ">", ">=", "<<", ">>", ">>>", "+", "-", "*", "/", "%", "|", "^", "&", "in", "instanceof"
	Left     Expression `json:"left"`
	Right    Expression `json:"right"`
}

func (b *BinaryExpression) expressionNode() {}

// AssignmentExpression represents an assignment expression.
type AssignmentExpression struct {
	BaseNode
	Operator string     `json:"operator"` // "=", "+=", "-=", "*=", "/=", "%=", "<<=", ">>=", ">>>=", "|=", "^=", "&=", "**=", "||=", "&&=", "??="
	Left     Pattern    `json:"left"`
	Right    Expression `json:"right"`
}

func (a *AssignmentExpression) expressionNode() {}

// LogicalExpression represents a logical operation expression.
type LogicalExpression struct {
	BaseNode
	Operator string     `json:"operator"` // "||", "&&", "??"
	Left     Expression `json:"left"`
	Right    Expression `json:"right"`
}

func (l *LogicalExpression) expressionNode() {}

// MemberExpression represents a member access expression.
type MemberExpression struct {
	BaseNode
	Object   Expression `json:"object"`
	Property Expression `json:"property"`
	Computed bool       `json:"computed"`
	Optional bool       `json:"optional"` // For optional chaining (?.)
}

func (m *MemberExpression) expressionNode() {}
func (m *MemberExpression) patternNode()    {}

// ConditionalExpression represents a ternary conditional expression.
type ConditionalExpression struct {
	BaseNode
	Test       Expression `json:"test"`
	Consequent Expression `json:"consequent"`
	Alternate  Expression `json:"alternate"`
}

func (c *ConditionalExpression) expressionNode() {}

// CallExpression represents a function call expression.
type CallExpression struct {
	BaseNode
	Callee    Expression   `json:"callee"`
	Arguments []Expression `json:"arguments"`
	Optional  bool         `json:"optional"` // For optional chaining (?.)
}

func (c *CallExpression) expressionNode() {}

// NewExpression represents a 'new' expression.
type NewExpression struct {
	BaseNode
	Callee    Expression   `json:"callee"`
	Arguments []Expression `json:"arguments"`
}

func (n *NewExpression) expressionNode() {}

// SequenceExpression represents a comma-separated sequence of expressions.
type SequenceExpression struct {
	BaseNode
	Expressions []Expression `json:"expressions"`
}

func (s *SequenceExpression) expressionNode() {}

// SpreadElement represents a spread element (...expression).
type SpreadElement struct {
	BaseNode
	Argument Expression `json:"argument"`
}

// YieldExpression represents a yield expression (ES2015).
type YieldExpression struct {
	BaseNode
	Argument *Expression `json:"argument"`
	Delegate bool        `json:"delegate"`
}

func (y *YieldExpression) expressionNode() {}

// AwaitExpression represents an await expression (ES2017).
type AwaitExpression struct {
	BaseNode
	Argument Expression `json:"argument"`
}

func (a *AwaitExpression) expressionNode() {}

// TemplateLiteral represents a template literal (ES2015).
type TemplateLiteral struct {
	BaseNode
	Quasis      []TemplateElement `json:"quasis"`
	Expressions []Expression      `json:"expressions"`
}

func (t *TemplateLiteral) expressionNode() {}

// TemplateElement represents an element in a template literal.
type TemplateElement struct {
	BaseNode
	Tail  bool                 `json:"tail"`
	Value TemplateElementValue `json:"value"`
}

// TemplateElementValue represents the value of a template element.
type TemplateElementValue struct {
	Cooked string `json:"cooked"` // Interpreted value
	Raw    string `json:"raw"`    // Raw source text
}

// TaggedTemplateExpression represents a tagged template expression (ES2015).
type TaggedTemplateExpression struct {
	BaseNode
	Tag   Expression       `json:"tag"`
	Quasi *TemplateLiteral `json:"quasi"`
}

func (t *TaggedTemplateExpression) expressionNode() {}

// ClassExpression represents a class expression (ES2015).
type ClassExpression struct {
	BaseNode
	ID         *Identifier `json:"id"`
	SuperClass *Expression `json:"superClass"`
	Body       *ClassBody  `json:"body"`
}

func (c *ClassExpression) expressionNode() {}

// ClassBody represents the body of a class.
type ClassBody struct {
	BaseNode
	Body []Node `json:"body"` // MethodDefinition, PropertyDefinition, or StaticBlock
}

// MethodDefinition represents a method in a class.
type MethodDefinition struct {
	BaseNode
	Key      Expression          `json:"key"`
	Value    *FunctionExpression `json:"value"`
	Kind     string              `json:"kind"` // "constructor", "method", "get", "set"
	Computed bool                `json:"computed"`
	Static   bool                `json:"static"`
}

// PropertyDefinition represents a property in a class (ES2022).
type PropertyDefinition struct {
	BaseNode
	Key      Expression  `json:"key"`
	Value    *Expression `json:"value"`
	Computed bool        `json:"computed"`
	Static   bool        `json:"static"`
}

// StaticBlock represents a static initialization block in a class (ES2022).
type StaticBlock struct {
	BaseNode
	Body []Statement `json:"body"`
}

// MetaProperty represents a meta property (e.g., new.target, import.meta).
type MetaProperty struct {
	BaseNode
	Meta     *Identifier `json:"meta"`
	Property *Identifier `json:"property"`
}

func (m *MetaProperty) expressionNode() {}

// ImportExpression represents a dynamic import() expression (ES2020).
type ImportExpression struct {
	BaseNode
	Source Expression `json:"source"`
}

func (i *ImportExpression) expressionNode() {}

// ChainExpression represents an optional chaining expression (ES2020).
type ChainExpression struct {
	BaseNode
	Expression Expression `json:"expression"` // Must be MemberExpression or CallExpression with optional: true
}

func (c *ChainExpression) expressionNode() {}
