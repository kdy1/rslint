package prefer_rest_params

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// PreferRestParamsRule implements the prefer-rest-params rule
// Require rest parameters instead of arguments
var PreferRestParamsRule = rule.Rule{
	Name: "prefer-rest-params",
	Run:  run,
}

func buildMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferRestParams",
		Description: "Use the rest parameters instead of 'arguments'.",
	}
}

// isInFunctionScope checks if we're inside a function that has its own arguments object
func isInFunctionScope(node *ast.Node) bool {
	current := node.Parent
	for current != nil {
		kind := current.Kind
		// Arrow functions don't have their own arguments
		if kind == ast.KindArrowFunction {
			return false
		}
		// Regular functions have arguments
		if kind == ast.KindFunctionDeclaration ||
			kind == ast.KindFunctionExpression ||
			kind == ast.KindMethodDeclaration ||
			kind == ast.KindConstructor ||
			kind == ast.KindGetAccessor ||
			kind == ast.KindSetAccessor {
			return true
		}
		current = current.Parent
	}
	return false
}

// isShadowedArguments checks if 'arguments' is shadowed by a parameter or local variable
func isShadowedArguments(node *ast.Node) bool {
	// Walk up to find the enclosing function
	current := node.Parent
	for current != nil {
		kind := current.Kind

		// Check if it's a parameter named 'arguments'
		if kind == ast.KindParameter {
			if param := current.AsParameterDeclaration(); param != nil && param.Name() != nil {
				if ident := param.Name().AsIdentifier(); ident != nil && ident.Text == "arguments" {
					return true
				}
			}
		}

		// Check if it's a variable declaration named 'arguments'
		if kind == ast.KindVariableDeclaration {
			if varDecl := current.AsVariableDeclaration(); varDecl != nil && varDecl.Name() != nil {
				if ident := varDecl.Name().AsIdentifier(); ident != nil && ident.Text == "arguments" {
					return true
				}
			}
		}

		// Stop at function boundary
		if kind == ast.KindFunctionDeclaration ||
			kind == ast.KindFunctionExpression ||
			kind == ast.KindArrowFunction ||
			kind == ast.KindMethodDeclaration ||
			kind == ast.KindConstructor {
			break
		}

		current = current.Parent
	}
	return false
}

// isArgumentsPropertyAccess checks if this is accessing a safe property of arguments
func isArgumentsPropertyAccess(node *ast.Node) bool {
	if node.Parent == nil {
		return false
	}

	parent := node.Parent
	// Check if parent is a property access expression
	if parent.Kind == ast.KindPropertyAccessExpression {
		if propAccess := parent.AsPropertyAccessExpression(); propAccess != nil {
			if propAccess.Expression == node && propAccess.Name() != nil {
				propName := propAccess.Name().Text()
				// Allow .length and .callee
				return propName == "length" || propName == "callee"
			}
		}
	}

	return false
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	return rule.RuleListeners{
		ast.KindIdentifier: func(node *ast.Node) {
			ident := node.AsIdentifier()
			if ident == nil || ident.Text != "arguments" {
				return
			}

			// Don't flag if shadowed by parameter or variable
			if isShadowedArguments(node) {
				return
			}

			// Don't flag if not in a function scope (or in arrow function)
			if !isInFunctionScope(node) {
				return
			}

			// Don't flag if accessing .length or .callee properties
			if isArgumentsPropertyAccess(node) {
				return
			}

			// Report the violation
			// Note: Auto-fix is complex (requires modifying function signature),
			// so we only report without providing a fix
			ctx.ReportNode(node, buildMessage())
		},
	}
}
