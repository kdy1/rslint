package no_invalid_this

import (
	"unicode"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type NoInvalidThisOptions struct {
	CapIsConstructor bool `json:"capIsConstructor"`
}

// Scope tracking for 'this' validity
type scopeInfo struct {
	init       bool
	valid      bool
	thisAllowed bool
	node       *ast.Node
	upper      *scopeInfo
}

// Check if a function name starts with uppercase (potential constructor)
func startsWithUpperCase(name string) bool {
	if len(name) == 0 {
		return false
	}
	firstRune := []rune(name)[0]
	return unicode.IsUpper(firstRune)
}

// Check if a function has explicit 'this' parameter (TypeScript feature)
func hasExplicitThisParameter(node *ast.Node) bool {
	var params []*ast.Node

	switch node.Kind {
	case ast.KindFunctionDeclaration:
		if fn := node.AsFunctionDeclaration(); fn != nil && fn.Parameters != nil {
			params = fn.Parameters.Nodes
		}
	case ast.KindFunctionExpression:
		if fn := node.AsFunctionExpression(); fn != nil && fn.Parameters != nil {
			params = fn.Parameters.Nodes
		}
	case ast.KindArrowFunction:
		if fn := node.AsArrowFunction(); fn != nil && fn.Parameters != nil {
			params = fn.Parameters.Nodes
		}
	case ast.KindMethodDeclaration:
		if fn := node.AsMethodDeclaration(); fn != nil && fn.Parameters != nil {
			params = fn.Parameters.Nodes
		}
	}

	if len(params) == 0 {
		return false
	}

	// Check if first parameter is a 'this' parameter
	firstParam := params[0]
	if param := firstParam.AsParameterDeclaration(); param != nil {
		if param.Name() != nil && ast.IsIdentifier(param.Name()) {
			if ident := param.Name().AsIdentifier(); ident != nil {
				return ident.Text == "this"
			}
		}
	}

	return false
}

// Check if function has JSDoc @this annotation
// Note: Simplified implementation - full JSDoc parsing not yet available
func hasJSDocThisAnnotation(node *ast.Node) bool {
	// TODO: Implement when JSDoc APIs are available
	return false
}

// Get the name of a function for checking capitalization
func getFunctionName(node *ast.Node) string {
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		if fn := node.AsFunctionDeclaration(); fn != nil && fn.Name() != nil {
			if ast.IsIdentifier(fn.Name()) {
				if ident := fn.Name().AsIdentifier(); ident != nil {
					return ident.Text
				}
			}
		}
	case ast.KindFunctionExpression:
		if fn := node.AsFunctionExpression(); fn != nil && fn.Name() != nil {
			if ast.IsIdentifier(fn.Name()) {
				if ident := fn.Name().AsIdentifier(); ident != nil {
					return ident.Text
				}
			}
		}
	}
	return ""
}

// Check if node is a method definition (object literal or class method)
func isMethodDefinition(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}

	parent := node.Parent

	// Check if it's a class method
	if node.Kind == ast.KindMethodDeclaration ||
	   node.Kind == ast.KindConstructor ||
	   node.Kind == ast.KindGetAccessor ||
	   node.Kind == ast.KindSetAccessor {
		return true
	}

	// Check if it's a property in an object literal
	if parent.Kind == ast.KindPropertyAssignment {
		if prop := parent.AsPropertyAssignment(); prop != nil {
			if prop.Initializer == node {
				grandparent := parent.Parent
				if grandparent != nil && grandparent.Kind == ast.KindObjectLiteralExpression {
					return true
				}
			}
		}
	}

	// Check if it's a method in an object literal
	if parent.Kind == ast.KindMethodDeclaration {
		grandparent := parent.Parent
		if grandparent != nil && grandparent.Kind == ast.KindObjectLiteralExpression {
			return true
		}
	}

	// Check shorthand method in object literal
	if node.Kind == ast.KindFunctionExpression && parent.Kind == ast.KindPropertyAssignment {
		grandparent := parent.Parent
		if grandparent != nil && grandparent.Kind == ast.KindObjectLiteralExpression {
			return true
		}
	}

	return false
}

// Check if function is bound via .bind(), .call(), or .apply()
func isCallWithNonNullContext(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}

	parent := node.Parent

	// Check for .bind(obj), .call(obj), .apply(obj)
	if parent.Kind == ast.KindCallExpression {
		if call := parent.AsCallExpression(); call != nil {
			// Check if the expression is a property access (.bind, .call, .apply)
			if ast.IsPropertyAccessExpression(call.Expression) {
				if propAccess := call.Expression.AsPropertyAccessExpression(); propAccess != nil {
					methodName := ""
					if ast.IsIdentifier(propAccess.Name()) {
						if ident := propAccess.Name().AsIdentifier(); ident != nil {
							methodName = ident.Text
						}
					}

					// Check if it's one of the binding methods
					if methodName == "bind" || methodName == "call" || methodName == "apply" {
						// Check if the bound function is our node
						if propAccess.Expression == node {
							// Check if there's a thisArg parameter and it's not null/undefined
							if call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
								firstArg := call.Arguments.Nodes[0]
								// If it's null or undefined, return false
								if firstArg.Kind == ast.KindNullKeyword {
									return false
								}
								if ast.IsIdentifier(firstArg) {
									if ident := firstArg.AsIdentifier(); ident != nil {
										if ident.Text == "undefined" {
											return false
										}
									}
								}
								return true
							}
						}
					}
				}
			}
		}
	}

	return false
}

// Check if function is passed to array method with thisArg
func isArrayMethodWithThisArg(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}

	parent := node.Parent

	// Check if parent is a call expression
	if parent.Kind == ast.KindCallExpression {
		if call := parent.AsCallExpression(); call != nil {
			// Check if the expression is a property access (e.g., array.forEach)
			if ast.IsPropertyAccessExpression(call.Expression) {
				if propAccess := call.Expression.AsPropertyAccessExpression(); propAccess != nil {
					methodName := ""
					if ast.IsIdentifier(propAccess.Name()) {
						if ident := propAccess.Name().AsIdentifier(); ident != nil {
							methodName = ident.Text
						}
					}

					// Array methods that accept thisArg as second parameter
					arrayMethods := map[string]bool{
						"forEach":   true,
						"filter":    true,
						"map":       true,
						"every":     true,
						"some":      true,
						"find":      true,
						"findIndex": true,
						"flatMap":   true,
					}

					if arrayMethods[methodName] {
						// Check if our node is the callback (first argument)
						if call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
							if call.Arguments.Nodes[0] == node {
								// Check if there's a thisArg (second argument)
								if len(call.Arguments.Nodes) > 1 {
									return true
								}
							}
						}
					}
				}
			}
		}
	}

	return false
}

// Check if inside a class (including class expressions)
func isInClass(node *ast.Node) bool {
	current := node.Parent
	for current != nil {
		if current.Kind == ast.KindClassDeclaration || current.Kind == ast.KindClassExpression {
			return true
		}
		current = current.Parent
	}
	return false
}

// Check if node is a constructor
func isConstructor(node *ast.Node) bool {
	return node.Kind == ast.KindConstructor
}

// Check if inside a class field initializer or static block
func isInClassFieldOrStaticBlock(node *ast.Node) bool {
	current := node.Parent
	for current != nil {
		if current.Kind == ast.KindPropertyDeclaration {
			return true
		}
		if current.Kind == ast.KindClassStaticBlockDeclaration {
			return true
		}
		// Stop at class boundary
		if current.Kind == ast.KindClassDeclaration || current.Kind == ast.KindClassExpression {
			return false
		}
		current = current.Parent
	}
	return false
}

var NoInvalidThisRule = rule.CreateRule(rule.Rule{
	Name: "no-invalid-this",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := NoInvalidThisOptions{
			CapIsConstructor: true, // default
		}

		// Parse options
		if options != nil {
			var optsMap map[string]interface{}
			if optsArray, ok := options.([]interface{}); ok && len(optsArray) > 0 {
				if opts, ok := optsArray[0].(map[string]interface{}); ok {
					optsMap = opts
				}
			} else if opts, ok := options.(map[string]interface{}); ok {
				optsMap = opts
			}

			if optsMap != nil {
				if capIsConstructor, ok := optsMap["capIsConstructor"].(bool); ok {
					opts.CapIsConstructor = capIsConstructor
				}
			}
		}

		var currentScope *scopeInfo

		// Enter a function scope
		enterFunction := func(node *ast.Node, thisAllowed bool) {
			currentScope = &scopeInfo{
				init:       true,
				valid:      true,
				thisAllowed: thisAllowed,
				node:       node,
				upper:      currentScope,
			}
		}

		// Exit a function scope
		exitFunction := func(node *ast.Node) {
			if currentScope != nil && currentScope.node == node {
				currentScope = currentScope.upper
			}
		}

		// Determine if 'this' is allowed in this function
		isThisAllowed := func(node *ast.Node) bool {
			// TypeScript explicit 'this' parameter
			if hasExplicitThisParameter(node) {
				return true
			}

			// JSDoc @this annotation
			if hasJSDocThisAnnotation(node) {
				return true
			}

			// Constructors always allow 'this'
			if isConstructor(node) {
				return true
			}

			// Class methods allow 'this'
			if isMethodDefinition(node) {
				return true
			}

			// Functions in class field initializers or static blocks
			if isInClassFieldOrStaticBlock(node) {
				return true
			}

			// Check for capitalized function name (potential constructor)
			if opts.CapIsConstructor {
				name := getFunctionName(node)
				if name != "" && startsWithUpperCase(name) {
					return true
				}
			}

			// Bound functions (.bind, .call, .apply with context)
			if isCallWithNonNullContext(node) {
				return true
			}

			// Array methods with thisArg
			if isArrayMethodWithThisArg(node) {
				return true
			}

			return false
		}

		// Check 'this' keyword usage
		checkThis := func(node *ast.Node) {
			// If we're not tracking scopes, we're at module/global level
			if currentScope == nil {
				// Top-level 'this' in modules is always invalid
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "unexpectedThis",
					Description: "Unexpected 'this'.",
				})
				return
			}

			// Check if 'this' is allowed in current scope
			if !currentScope.thisAllowed {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "unexpectedThis",
					Description: "Unexpected 'this'.",
				})
			}
		}

		return rule.RuleListeners{
			// Function declarations
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				thisAllowed := isThisAllowed(node)
				enterFunction(node, thisAllowed)
			},
			rule.ListenerOnExit(ast.KindFunctionDeclaration): exitFunction,

			// Function expressions
			ast.KindFunctionExpression: func(node *ast.Node) {
				thisAllowed := isThisAllowed(node)
				enterFunction(node, thisAllowed)
			},
			rule.ListenerOnExit(ast.KindFunctionExpression): exitFunction,

			// Arrow functions (inherit 'this' from enclosing scope, don't create new scope)
			// Arrow functions don't change 'this' binding, so we don't create a new scope

			// Method declarations
			ast.KindMethodDeclaration: func(node *ast.Node) {
				enterFunction(node, true) // Methods always allow 'this'
			},
			rule.ListenerOnExit(ast.KindMethodDeclaration): exitFunction,

			// Constructors
			ast.KindConstructor: func(node *ast.Node) {
				enterFunction(node, true) // Constructors always allow 'this'
			},
			rule.ListenerOnExit(ast.KindConstructor): exitFunction,

			// Getters and setters
			ast.KindGetAccessor: func(node *ast.Node) {
				enterFunction(node, true)
			},
			rule.ListenerOnExit(ast.KindGetAccessor): exitFunction,

			ast.KindSetAccessor: func(node *ast.Node) {
				enterFunction(node, true)
			},
			rule.ListenerOnExit(ast.KindSetAccessor): exitFunction,

			// Check 'this' keyword usage
			ast.KindThisKeyword: checkThis,
		}
	},
})
