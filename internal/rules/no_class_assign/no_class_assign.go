package no_class_assign

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoClassAssignRule implements the no-class-assign rule
// Disallow reassigning class members
var NoClassAssignRule = rule.Rule{
	Name: "no-class-assign",
	Run:  run,
}

// classNames tracks class declaration names in the current scope
type scopeInfo struct {
	classNames map[string]*ast.Node // maps class name to the declaration node
	parent     *scopeInfo
}

func newScopeInfo(parent *scopeInfo) *scopeInfo {
	return &scopeInfo{
		classNames: make(map[string]*ast.Node),
		parent:     parent,
	}
}

func (s *scopeInfo) addClassName(name string, node *ast.Node) {
	s.classNames[name] = node
}

func (s *scopeInfo) isClassName(name string) (*ast.Node, bool) {
	// Check current scope
	if node, ok := s.classNames[name]; ok {
		return node, true
	}
	// Check parent scopes
	if s.parent != nil {
		return s.parent.isClassName(name)
	}
	return nil, false
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	// Track scopes
	var currentScope *scopeInfo

	// Helper to enter a new scope
	enterScope := func() {
		currentScope = newScopeInfo(currentScope)
	}

	// Helper to exit scope
	exitScope := func() {
		if currentScope != nil {
			currentScope = currentScope.parent
		}
	}

	// Helper to check if an identifier is being assigned
	isAssignmentTarget := func(node *ast.Node) bool {
		if node == nil || node.Parent == nil {
			return false
		}

		parent := node.Parent

		// Check for binary assignment (e.g., A = 0)
		if parent.Kind == ast.KindBinaryExpression {
			binExpr := parent.AsBinaryExpression()
			if binExpr != nil && binExpr.OperatorToken != nil {
				operatorKind := binExpr.OperatorToken.Kind
				// Check if it's an assignment operator
				if operatorKind == ast.KindEqualsToken ||
					operatorKind == ast.KindPlusEqualsToken ||
					operatorKind == ast.KindMinusEqualsToken ||
					operatorKind == ast.KindAsteriskEqualsToken ||
					operatorKind == ast.KindSlashEqualsToken ||
					operatorKind == ast.KindPercentEqualsToken ||
					operatorKind == ast.KindAsteriskAsteriskEqualsToken ||
					operatorKind == ast.KindLessThanLessThanEqualsToken ||
					operatorKind == ast.KindGreaterThanGreaterThanEqualsToken ||
					operatorKind == ast.KindGreaterThanGreaterThanGreaterThanEqualsToken ||
					operatorKind == ast.KindAmpersandEqualsToken ||
					operatorKind == ast.KindBarEqualsToken ||
					operatorKind == ast.KindCaretEqualsToken ||
					operatorKind == ast.KindBarBarEqualsToken ||
					operatorKind == ast.KindAmpersandAmpersandEqualsToken ||
					operatorKind == ast.KindQuestionQuestionEqualsToken {
					// Check if this node is the left side
					return binExpr.Left == node
				}
			}
		}

		// Check for object destructuring ({A} = 0, {b: A} = {})
		if parent.Kind == ast.KindPropertyAssignment ||
			parent.Kind == ast.KindShorthandPropertyAssignment ||
			parent.Kind == ast.KindBindingElement {
			return true
		}

		// Check for unary operators (++, --)
		if parent.Kind == ast.KindPrefixUnaryExpression || parent.Kind == ast.KindPostfixUnaryExpression {
			return true
		}

		return false
	}

	return rule.RuleListeners{
		// Track class declarations
		ast.KindClassDeclaration: func(node *ast.Node) {
			if node == nil {
				return
			}

			classDecl := node.AsClassDeclaration()
			if classDecl == nil || classDecl.Name == nil {
				return
			}

			nameNode := classDecl.Name
			if nameNode.Kind == ast.KindIdentifier {
				ident := nameNode.AsIdentifier()
				if ident != nil && ident.EscapedText != nil {
					className := *ident.EscapedText
					if currentScope != nil {
						currentScope.addClassName(className, node)
					}
				}
			}
		},

		// Track class expressions with names
		ast.KindClassExpression: func(node *ast.Node) {
			if node == nil {
				return
			}

			classExpr := node.AsClassExpression()
			if classExpr == nil || classExpr.Name == nil {
				return
			}

			nameNode := classExpr.Name
			if nameNode.Kind == ast.KindIdentifier {
				ident := nameNode.AsIdentifier()
				if ident != nil && ident.EscapedText != nil {
					className := *ident.EscapedText
					// For class expressions, we need to track them in scope too
					if currentScope != nil {
						currentScope.addClassName(className, node)
					}
				}
			}
		},

		// Check identifiers for assignments to class names
		ast.KindIdentifier: func(node *ast.Node) {
			if node == nil || currentScope == nil {
				return
			}

			ident := node.AsIdentifier()
			if ident == nil || ident.EscapedText == nil {
				return
			}

			name := *ident.EscapedText

			// Check if this identifier is a class name
			if classNode, isClass := currentScope.isClassName(name); isClass {
				// Check if it's being assigned
				if isAssignmentTarget(node) {
					// Don't report if this is the class declaration itself
					if node.Parent != classNode {
						ctx.ReportNode(node, rule.RuleMessage{
							Id:          "class",
							Description: "'" + name + "' is a class.",
						})
					}
				}
			}
		},

		// Enter new scopes
		ast.KindSourceFile: func(node *ast.Node) {
			enterScope()
		},
		ast.KindFunctionDeclaration: func(node *ast.Node) {
			enterScope()
		},
		ast.KindFunctionExpression: func(node *ast.Node) {
			enterScope()
		},
		ast.KindArrowFunction: func(node *ast.Node) {
			enterScope()
		},
		ast.KindBlock: func(node *ast.Node) {
			enterScope()
		},

		// Exit scopes
		rule.ListenerOnExit(ast.KindSourceFile): func(node *ast.Node) {
			exitScope()
		},
		rule.ListenerOnExit(ast.KindFunctionDeclaration): func(node *ast.Node) {
			exitScope()
		},
		rule.ListenerOnExit(ast.KindFunctionExpression): func(node *ast.Node) {
			exitScope()
		},
		rule.ListenerOnExit(ast.KindArrowFunction): func(node *ast.Node) {
			exitScope()
		},
		rule.ListenerOnExit(ast.KindBlock): func(node *ast.Node) {
			exitScope()
		},
	}
}
