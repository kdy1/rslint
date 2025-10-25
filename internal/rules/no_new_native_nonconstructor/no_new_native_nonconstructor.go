package no_new_native_nonconstructor

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoNewNativeNonconstructorRule implements the no-new-native-nonconstructor rule
// Disallow new operators with global non-constructor functions
var NoNewNativeNonconstructorRule = rule.Rule{
	Name: "no-new-native-nonconstructor",
	Run:  run,
}

var nonConstructorGlobals = map[string]string{
	"Symbol": "`Symbol` cannot be called as a constructor.",
	"BigInt": "`BigInt` cannot be called as a constructor.",
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	return rule.RuleListeners{
		ast.KindNewExpression: func(node *ast.Node) {
			newExpr := node.AsNewExpression()
			if newExpr == nil {
				return
			}

			// Check if the expression is an identifier
			if newExpr.Expression == nil || newExpr.Expression.Kind != ast.KindIdentifier {
				return
			}

			identifier := newExpr.Expression.AsIdentifier()
			if identifier == nil {
				return
			}

			name := identifier.Text
			errorMessage, isNonConstructor := nonConstructorGlobals[name]
			if !isNonConstructor {
				return
			}

			// Check if the identifier is shadowed by a local variable
			if ctx.TypeChecker != nil {
				symbol := ctx.TypeChecker.GetSymbolAtLocation(newExpr.Expression)
				if symbol != nil && len(symbol.Declarations) > 0 {
					// Check if any declaration is a parameter or local variable
					for _, decl := range symbol.Declarations {
						// If it's a parameter or variable declaration, it's shadowing the global
						if decl.Kind == ast.KindParameter ||
							decl.Kind == ast.KindVariableDeclaration ||
							decl.Kind == ast.KindFunctionDeclaration {
							// This is a local binding, not the global Symbol/BigInt
							return
						}
					}
				}
			}

			// Report the error
			ctx.ReportNode(newExpr.Expression, rule.RuleMessage{
				Id:          "noNewNativeNonconstructor",
				Description: errorMessage,
			})
		},
	}
}
