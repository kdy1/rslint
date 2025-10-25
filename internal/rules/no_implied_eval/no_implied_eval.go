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
		// Template literals with substitutions are also strings
		return true
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

// getCalleeName extracts the function name from a call expression
func getCalleeName(callExpr *ast.CallExpression) (string, bool) {
	if callExpr == nil || callExpr.Expression == nil {
		return "", false
	}

	switch callExpr.Expression.Kind {
	case ast.KindIdentifier:
		ident := callExpr.Expression.AsIdentifier()
		if ident != nil {
			return ident.Text, true
		}
	case ast.KindPropertyAccessExpression:
		propAccess := callExpr.Expression.AsPropertyAccessExpression()
		if propAccess != nil && propAccess.Name() != nil {
			// For window.setTimeout or global.setTimeout
			if isGlobalReference(propAccess.Expression) {
				return propAccess.Name().Text(), true
			}
		}
	}

	return "", false
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

			funcName, ok := getCalleeName(callExpr)
			if !ok {
				return
			}

			// Check for execScript - it's always bad
			if funcName == "execScript" {
				ctx.ReportNode(callExpr.Expression, buildExecScriptMessage())
				return
			}

			// Check for setTimeout and setInterval with string arguments
			if funcName == "setTimeout" || funcName == "setInterval" {
				// Check if first argument is a string
				if callExpr.Arguments != nil && len(callExpr.Arguments.Nodes) > 0 {
					firstArg := callExpr.Arguments.Nodes[0]
					if isStringArgument(firstArg) {
						ctx.ReportNode(callExpr.Expression, buildImpliedEvalMessage())
					}
				}
			}
		}

		return listeners
	},
})
