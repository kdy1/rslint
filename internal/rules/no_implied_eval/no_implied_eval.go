package no_implied_eval

import (
	"github.com/microsoft/typescript-go/shim/ast"

	"github.com/web-infra-dev/rslint/internal/rule"
)

func buildImpliedEvalMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "impliedEval",
		Description: "Implied eval. Consider passing a function instead of a string.",
	}
}

func buildExecScriptMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "execScript",
		Description: "Expected a function, instead saw execScript.",
	}
}

// isStringArgument checks if a node is a string literal or string-like expression
func isStringArgument(node *ast.Node) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return true
	case ast.KindTemplateExpression:
		// Template literals with substitutions like `foo${bar}` are not flagged
		// because the substitution might resolve to a safe value
		return false
	case ast.KindBinaryExpression:
		// Check for string concatenation
		binExpr := node.AsBinaryExpression()
		if binExpr != nil && binExpr.OperatorToken.Kind == ast.KindPlusToken {
			// If either side is a string, this could be string concatenation
			return isStringArgument(binExpr.Left) || isStringArgument(binExpr.Right)
		}
	case ast.KindParenthesizedExpression:
		parenExpr := node.AsParenthesizedExpression()
		if parenExpr != nil {
			return isStringArgument(parenExpr.Expression)
		}
	}

	return false
}

// isGlobalReference checks if a name is a global reference (window, global, globalThis)
func isGlobalReference(node *ast.Node) bool {
	if node == nil {
		return false
	}

	if node.Kind == ast.KindIdentifier {
		ident := node.AsIdentifier()
		if ident != nil {
			name := ident.Text
			return name == "window" || name == "global" || name == "globalThis"
		}
	}

	return false
}

// isShadowedGlobal checks if a global identifier (window, global, globalThis) is shadowed by a local declaration
func isShadowedGlobal(node *ast.Node, sourceFile *ast.SourceFile) bool {
	if node == nil || node.Kind != ast.KindIdentifier {
		return false
	}

	ident := node.AsIdentifier()
	if ident == nil {
		return false
	}

	name := ident.Text
	if name != "window" && name != "global" && name != "globalThis" {
		return false
	}

	// Walk up to find the containing function or source file to check for variable declarations
	current := node.Parent
	for current != nil {
		// Check if there's a variable declaration with this name in the current scope
		if current.Kind == ast.KindSourceFile {
			sf := current.AsSourceFile()
			if sf != nil && sf.Statements != nil {
				for _, stmt := range sf.Statements.Nodes {
					if stmt.Kind == ast.KindVariableStatement {
						varStmt := stmt.AsVariableStatement()
						if varStmt != nil && varStmt.DeclarationList != nil {
							declList := varStmt.DeclarationList.AsVariableDeclarationList()
							if declList != nil {
								for _, decl := range declList.Declarations.Nodes {
									if decl.Name() != nil && decl.Name().Kind == ast.KindIdentifier {
										declIdent := decl.Name().AsIdentifier()
										if declIdent != nil && declIdent.Text == name {
											return true
										}
									}
								}
							}
						}
					}
				}
			}
			break
		}
		current = current.Parent
	}

	return false
}

// getCalleeName extracts the function name from a call expression and returns the node to report
func getCalleeName(callExpr *ast.CallExpression, sourceFile *ast.SourceFile) (string, *ast.Node, bool) {
	if callExpr == nil || callExpr.Expression == nil {
		return "", nil, false
	}

	switch callExpr.Expression.Kind {
	case ast.KindIdentifier:
		ident := callExpr.Expression.AsIdentifier()
		if ident != nil {
			return ident.Text, callExpr.Expression, true
		}
	case ast.KindPropertyAccessExpression:
		propAccess := callExpr.Expression.AsPropertyAccessExpression()
		if propAccess != nil && propAccess.Name() != nil {
			// For window.setTimeout or global.setTimeout
			if isGlobalReference(propAccess.Expression) && !isShadowedGlobal(propAccess.Expression, sourceFile) {
				return propAccess.Name().Text(), propAccess.Name(), true
			}
		}
	}

	return "", nil, false
}

// NoImpliedEvalRule disallows implied eval via setTimeout, setInterval, or execScript
var NoImpliedEvalRule = rule.CreateRule(rule.Rule{
	Name: "no-implied-eval",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		listeners := rule.RuleListeners{}

		listeners[ast.KindCallExpression] = func(node *ast.Node) {
			callExpr := node.AsCallExpression()
			if callExpr == nil || callExpr.Expression == nil {
				return
			}

			funcName, reportNode, ok := getCalleeName(callExpr, ctx.SourceFile)
			if !ok {
				return
			}

			// Check for execScript - it's bad unless the first argument is clearly a function
			if funcName == "execScript" {
				// Only allow execScript if the first argument is a function expression or arrow function
				if callExpr.Arguments != nil && len(callExpr.Arguments.Nodes) > 0 {
					firstArg := callExpr.Arguments.Nodes[0]
					// Allow function expressions and arrow functions
					if firstArg.Kind != ast.KindFunctionExpression && firstArg.Kind != ast.KindArrowFunction {
						ctx.ReportNode(reportNode, buildExecScriptMessage())
					}
				} else {
					// No arguments, still bad
					ctx.ReportNode(reportNode, buildExecScriptMessage())
				}
				return
			}

			// Check for setTimeout and setInterval with string arguments
			if funcName == "setTimeout" || funcName == "setInterval" {
				// Check if first argument is a string
				if callExpr.Arguments != nil && len(callExpr.Arguments.Nodes) > 0 {
					firstArg := callExpr.Arguments.Nodes[0]
					if isStringArgument(firstArg) {
						ctx.ReportNode(reportNode, buildImpliedEvalMessage())
					}
				}
			}
		}

		return listeners
	},
})
