package no_unsafe_function_type

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// isGlobalFunction checks if a type reference to "Function" refers to the global Function type
// and not a locally scoped type alias.
func isGlobalFunction(ctx rule.RuleContext, node *ast.Node) bool {
	typeRef := node.AsTypeReferenceNode()
	if typeRef == nil || typeRef.TypeName == nil {
		return false
	}

	// Check if it's a simple identifier "Function"
	if !ast.IsIdentifier(typeRef.TypeName) {
		return false
	}

	ident := typeRef.TypeName.AsIdentifier()
	if ident == nil || ident.Text != "Function" {
		return false
	}

	// Use type checker to see if this references the global Function type
	if ctx.TypeChecker != nil {
		symbol := ctx.TypeChecker.GetSymbolAtLocation(typeRef.TypeName)
		if symbol != nil {
			// Check if this symbol has a local declaration (type alias)
			// If it has declarations in the same file, it's likely a local type
			for _, decl := range symbol.Declarations {
				// If there's a type alias declaration for "Function", it's not the global one
				if decl.Kind == ast.KindTypeAliasDeclaration {
					return false
				}
			}
		}
	}

	return true
}

// isGlobalFunctionInExpression checks if an expression references the global Function type
func isGlobalFunctionInExpression(ctx rule.RuleContext, expr *ast.Node) bool {
	if expr == nil {
		return false
	}

	// Check if it's a simple identifier "Function"
	if !ast.IsIdentifier(expr) {
		return false
	}

	ident := expr.AsIdentifier()
	if ident == nil || ident.Text != "Function" {
		return false
	}

	// Use type checker to see if this references the global Function type
	if ctx.TypeChecker != nil {
		symbol := ctx.TypeChecker.GetSymbolAtLocation(expr)
		if symbol != nil {
			// Check if this symbol has a local declaration (type alias)
			for _, decl := range symbol.Declarations {
				// If there's a type alias declaration for "Function", it's not the global one
				if decl.Kind == ast.KindTypeAliasDeclaration {
					return false
				}
			}
		}
	}

	return true
}

var NoUnsafeFunctionTypeRule = rule.CreateRule(rule.Rule{
	Name: "no-unsafe-function-type",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindTypeReference: func(node *ast.Node) {
				if isGlobalFunction(ctx, node) {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "bannedFunctionType",
						Description: "The `Function` type accepts any function-like value, including class declarations which will throw at runtime as they will not be called with `new`.",
					})
				}
			},
			ast.KindExpressionWithTypeArguments: func(node *ast.Node) {
				expr := node.AsExpressionWithTypeArguments()
				if expr == nil {
					return
				}
				// Check if the expression is "Function"
				if !ast.IsIdentifier(expr.Expression) {
					return
				}
				ident := expr.Expression.AsIdentifier()
				if ident == nil || ident.Text != "Function" {
					return
				}

				// Check if it's the global Function type, not a local alias
				isGlobal := true
				if ctx.TypeChecker != nil {
					symbol := ctx.TypeChecker.GetSymbolAtLocation(expr.Expression)
					if symbol != nil {
						for _, decl := range symbol.Declarations {
							if decl.Kind == ast.KindTypeAliasDeclaration {
								isGlobal = false
								break
							}
						}
					}
				}

				if isGlobal {
					// Report on the identifier for correct positioning
					ctx.ReportNode(expr.Expression, rule.RuleMessage{
						Id:          "bannedFunctionType",
						Description: "The `Function` type accepts any function-like value, including class declarations which will throw at runtime as they will not be called with `new`.",
					})
				}
			},
		}
	},
})
