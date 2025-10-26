package no_useless_constructor

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoUselessConstructorRule implements the no-useless-constructor rule
// Disallow unnecessary constructors
var NoUselessConstructorRule = rule.CreateRule(rule.Rule{
	Name: "no-useless-constructor",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	return rule.RuleListeners{
		ast.KindConstructor: func(node *ast.Node) {
			constructor := node.AsConstructorDeclaration()
			if constructor == nil {
				return
			}

			// Skip if constructor has protected or private modifier
			modifierFlags := ast.GetCombinedModifierFlags(node)
			if modifierFlags&ast.ModifierFlagsProtected != 0 || modifierFlags&ast.ModifierFlagsPrivate != 0 {
				return
			}

			// Get parent class to check if it has a superclass
			parent := node.Parent
			if parent == nil || parent.Kind != ast.KindClassDeclaration {
				return
			}

			classDecl := parent.AsClassDeclaration()
			if classDecl == nil {
				return
			}

			hasSuperClass := classDecl.HeritageClauses != nil && len(classDecl.HeritageClauses.Nodes) > 0

			// If public constructor in class with superclass, skip
			if modifierFlags&ast.ModifierFlagsPublic != 0 && hasSuperClass {
				return
			}

			// Check if any parameter has modifiers (parameter properties) or decorators
			if constructor.Parameters != nil {
				for _, param := range constructor.Parameters.Nodes {
					if param.Kind == ast.KindParameter {
						paramDecl := param.AsParameterDeclaration()
						if paramDecl == nil {
							continue
						}

						// Check for parameter properties (public, private, protected, readonly)
						if paramDecl.Modifiers() != nil {
							for _, mod := range paramDecl.Modifiers().Nodes {
								if mod.Kind == ast.KindPublicKeyword ||
									mod.Kind == ast.KindPrivateKeyword ||
									mod.Kind == ast.KindProtectedKeyword ||
									mod.Kind == ast.KindReadonlyKeyword {
									return
								}
							}
						}

						// Check for parameter decorators - skip for now
						// Parameter decorators detection would go here
					}
				}
			}

			// Now check if the constructor is useless
			if constructor.Body == nil {
				// Constructor has no body - this is useless
				reportUselessConstructor(ctx, node)
				return
			}

			statements := constructor.Body.Statements()
			if statements == nil || len(statements) == 0 {
				// Empty constructor
				if !hasSuperClass {
					// Empty constructor in class without superclass is useless
					reportUselessConstructor(ctx, node)
				}
				return
			}

			// If there's more than one statement, it's not useless
			if len(statements) > 1 {
				return
			}

			// Single statement - check if it's just calling super with same params
			stmt := statements[0]
			if stmt.Kind != ast.KindExpressionStatement {
				return
			}

			exprStmt := stmt.AsExpressionStatement()
			if exprStmt == nil || exprStmt.Expression == nil {
				return
			}

			expr := exprStmt.Expression
			if expr.Kind != ast.KindCallExpression {
				return
			}

			callExpr := expr.AsCallExpression()
			if callExpr == nil || callExpr.Expression == nil {
				return
			}

			// Check if it's calling super
			if callExpr.Expression.Kind != ast.KindSuperKeyword {
				return
			}

			// Check if parameters are passed through identically
			if isPassingThroughParams(constructor, callExpr) {
				reportUselessConstructor(ctx, node)
			}
		},
	}
}

func reportUselessConstructor(ctx rule.RuleContext, node *ast.Node) {
	ctx.ReportNode(node, rule.RuleMessage{
		Id:          "noUselessConstructor",
		Description: "Useless constructor.",
	})
}

// isPassingThroughParams checks if the constructor is just passing its parameters to super
func isPassingThroughParams(constructor *ast.ConstructorDeclaration, callExpr *ast.CallExpression) bool {
	// Handle super() with no arguments
	if callExpr.Arguments == nil || len(callExpr.Arguments.Nodes) == 0 {
		if constructor.Parameters == nil || len(constructor.Parameters.Nodes) == 0 {
			return true
		}
		return false
	}

	// Handle super(...arguments)
	if len(callExpr.Arguments.Nodes) == 1 {
		arg := callExpr.Arguments.Nodes[0]
		if arg.Kind == ast.KindSpreadElement {
			spreadElem := arg.AsSpreadElement()
			if spreadElem != nil && spreadElem.Expression != nil {
				if spreadElem.Expression.Kind == ast.KindIdentifier {
					ident := spreadElem.Expression.AsIdentifier()
					if ident != nil && ident.Text == "arguments" {
						return true
					}
				}
			}
		}
	}

	// Check if parameters match arguments
	if constructor.Parameters == nil {
		return false
	}

	params := constructor.Parameters.Nodes
	args := callExpr.Arguments.Nodes

	// Handle rest parameter: constructor(...args) { super(...args); }
	if len(params) == 1 && len(args) == 1 {
		param := params[0]
		arg := args[0]

		if param.Kind == ast.KindParameter && arg.Kind == ast.KindSpreadElement {
			paramDecl := param.AsParameterDeclaration()
			spreadElem := arg.AsSpreadElement()

			if paramDecl != nil && paramDecl.DotDotDotToken != nil &&
				spreadElem != nil && spreadElem.Expression != nil {
				if spreadElem.Expression.Kind == ast.KindIdentifier && paramDecl.Name() != nil &&
					paramDecl.Name().Kind == ast.KindIdentifier {
					paramName := paramDecl.Name().AsIdentifier()
					argName := spreadElem.Expression.AsIdentifier()
					if paramName != nil && argName != nil && paramName.Text == argName.Text {
						return true
					}
				}
			}
		}
	}

	// Check if all parameters are passed through in order
	if len(params) != len(args) {
		return false
	}

	for i := 0; i < len(params); i++ {
		param := params[i]
		arg := args[i]

		if param.Kind != ast.KindParameter || arg.Kind != ast.KindIdentifier {
			return false
		}

		paramDecl := param.AsParameterDeclaration()
		argIdent := arg.AsIdentifier()

		if paramDecl == nil || argIdent == nil {
			return false
		}

		paramName := paramDecl.Name()
		if paramName == nil || paramName.Kind != ast.KindIdentifier {
			return false
		}

		paramNameIdent := paramName.AsIdentifier()
		if paramNameIdent == nil || paramNameIdent.Text != argIdent.Text {
			return false
		}
	}

	return true
}
