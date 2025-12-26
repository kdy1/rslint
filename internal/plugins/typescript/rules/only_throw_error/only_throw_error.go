package only_throw_error

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type OnlyThrowErrorOptions struct {
	Allow                []utils.TypeOrValueSpecifier
	AllowInline          []string
	AllowThrowingAny     *bool
	AllowThrowingUnknown *bool
	AllowRethrowing      *bool
}

func buildObjectMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "object",
		Description: "Expected an error object to be thrown.",
	}
}
func buildUndefMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "undef",
		Description: "Do not throw undefined.",
	}
}

// getIdentifierText extracts text from an identifier node
func getIdentifierText(node *ast.Node) string {
	if node == nil || node.Kind != ast.KindIdentifier {
		return ""
	}
	return node.Text()
}

// isRethrownError checks if an identifier represents a caught error being rethrown
func isRethrownError(node *ast.Node) bool {
	if node.Kind != ast.KindIdentifier {
		return false
	}

	throwName := getIdentifierText(node)
	if throwName == "" {
		return false
	}

	// Walk up the tree to find if this identifier is from a catch clause
	current := node.Parent
	for current != nil {
		// Check if we're in a catch clause
		if current.Kind == ast.KindCatchClause {
			catchClause := current.AsCatchClause()
			if catchClause != nil && catchClause.VariableDeclaration != nil {
				// Check if the thrown identifier matches the catch variable
				varDecl := catchClause.VariableDeclaration.AsVariableDeclaration()
				if varDecl != nil && varDecl.Name() != nil {
					catchName := getIdentifierText(varDecl.Name())
					if catchName == throwName {
						return true
					}
				}
			}
		}

		// Check if we're in a promise catch/then handler
		// Looking for patterns like: .catch(e => { throw e })
		if current.Kind == ast.KindArrowFunction || current.Kind == ast.KindFunctionExpression {
			var params []*ast.Node
			if current.Kind == ast.KindArrowFunction {
				arrowFunc := current.AsArrowFunction()
				if arrowFunc != nil && arrowFunc.Parameters != nil {
					params = arrowFunc.Parameters.Nodes
				}
			} else {
				funcExpr := current.AsFunctionExpression()
				if funcExpr != nil && funcExpr.Parameters != nil {
					params = funcExpr.Parameters.Nodes
				}
			}

			if len(params) > 0 {
				// Check if the first parameter is a rest parameter - this is not allowed for rethrowing
				firstParam := params[0]
				if firstParam != nil && firstParam.Kind == ast.KindParameter {
					paramDecl := firstParam.AsParameterDeclaration()
					if paramDecl != nil && paramDecl.DotDotDotToken != nil {
						// Rest parameter (...e) - not allowed for rethrowing
						return false
					}
				}

				// Check if this is a callback to .catch() or .then()
				callExpr := current.Parent
				if callExpr != nil && callExpr.Kind == ast.KindCallExpression {
					call := callExpr.AsCallExpression()
					if call != nil && call.Expression != nil {
						// Check if it's a property access like promise.catch
						if call.Expression.Kind == ast.KindPropertyAccessExpression {
							propAccess := call.Expression.AsPropertyAccessExpression()
							if propAccess != nil && propAccess.Name() != nil {
								methodName := getIdentifierText(propAccess.Name())

								// Check for spread arguments before the handler - not allowed
								if call.Arguments != nil {
									args := call.Arguments.Nodes
									for i, arg := range args {
										if arg == current {
											// Check if any argument before this one is a spread element
											for j := 0; j < i; j++ {
												if args[j].Kind == ast.KindSpreadElement {
													return false
												}
											}
											break
										}
									}
								}

								// Only allow rethrowing from .catch() handlers
								// For .then(), we need to check it's the second parameter (rejection handler)
								if methodName == "catch" {
									// Check if the thrown identifier is the catch parameter
									if firstParam != nil && firstParam.Kind == ast.KindParameter {
										paramDecl := firstParam.AsParameterDeclaration()
										if paramDecl != nil && paramDecl.Name() != nil {
											paramName := getIdentifierText(paramDecl.Name())
											if paramName == throwName {
												return true
											}
										}
									}
								} else if methodName == "then" {
									// For .then(), only the second parameter (rejection handler) should allow rethrowing
									// Find which parameter index this function is
									if call.Arguments != nil {
										args := call.Arguments.Nodes
										isRejectionHandler := false
										for i, arg := range args {
											if arg == current && i == 1 {
												isRejectionHandler = true
												break
											}
										}
										if isRejectionHandler {
											// Check if the thrown identifier is the rejection parameter
											if firstParam != nil && firstParam.Kind == ast.KindParameter {
												paramDecl := firstParam.AsParameterDeclaration()
												if paramDecl != nil && paramDecl.Name() != nil {
													paramName := getIdentifierText(paramDecl.Name())
													if paramName == throwName {
														return true
													}
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}

		current = current.Parent
	}

	return false
}

var OnlyThrowErrorRule = rule.CreateRule(rule.Rule{
	Name: "only-throw-error",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts, ok := options.(OnlyThrowErrorOptions)
		if !ok {
			opts = OnlyThrowErrorOptions{}
		}
		if opts.Allow == nil {
			opts.Allow = []utils.TypeOrValueSpecifier{}
		}
		if opts.AllowInline == nil {
			opts.AllowInline = []string{}
		}
		if opts.AllowThrowingAny == nil {
			opts.AllowThrowingAny = utils.Ref(true)
		}
		if opts.AllowThrowingUnknown == nil {
			opts.AllowThrowingUnknown = utils.Ref(true)
		}
		if opts.AllowRethrowing == nil {
			opts.AllowRethrowing = utils.Ref(true)
		}

		return rule.RuleListeners{
			ast.KindThrowStatement: func(node *ast.Node) {
				expr := node.Expression()
				// TODO(port): why do we ignore await and yield here??
				// if (
				//   node.type === AST_NODE_TYPES.AwaitExpression ||
				//   node.type === AST_NODE_TYPES.YieldExpression
				// ) {
				//   return;
				// }

				// Check if rethrowing is allowed and this is a rethrown error
				if *opts.AllowRethrowing && isRethrownError(expr) {
					return
				}

				t := ctx.TypeChecker.GetTypeAtLocation(expr)

				if utils.TypeMatchesSomeSpecifier(t, opts.Allow, opts.AllowInline, ctx.Program) {
					return
				}

				if utils.IsTypeFlagSet(t, checker.TypeFlagsUndefined) {
					ctx.ReportNode(node, buildUndefMessage())
					return
				}

				if *opts.AllowThrowingAny && utils.IsTypeAnyType(t) {
					return
				}

				if *opts.AllowThrowingUnknown && utils.IsTypeUnknownType(t) {
					return
				}

				if utils.IsErrorLike(ctx.Program, ctx.TypeChecker, t) {
					return
				}

				ctx.ReportNode(expr, buildObjectMessage())
			},
		}
	},
})
