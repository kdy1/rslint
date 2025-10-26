package no_alert

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoAlertRule implements the no-alert rule
// Disallow the use of `alert`, `confirm`, and `prompt`
var NoAlertRule = rule.CreateRule(rule.Rule{
	Name: "no-alert",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	// Map of forbidden methods
	forbiddenMethods := map[string]bool{
		"alert":   true,
		"confirm": true,
		"prompt":  true,
	}

	isGlobalReference := func(node *ast.Node) bool {
		// Check if this is a reference to window, globalThis, or this
		if node == nil {
			return false
		}

		switch node.Kind {
		case ast.KindIdentifier:
			text := node.Text()
			return text == "window" || text == "globalThis"
		case ast.KindThisKeyword:
			return true
		}

		return false
	}

	checkCallExpression := func(node *ast.Node) {
		if node == nil {
			return
		}

		expr := node.Expression()
		if expr == nil {
			return
		}

		var methodName string
		var isForbidden bool

		switch expr.Kind {
		case ast.KindIdentifier:
			// Direct call: alert(), confirm(), prompt()
			methodName = expr.Text()
			isForbidden = forbiddenMethods[methodName]

		case ast.KindPropertyAccessExpression:
			// Property access: window.alert(), this.alert(), etc.
			propName := expr.Name()
			if propName != nil && propName.Kind == ast.KindIdentifier {
				methodName = propName.Text()
				if forbiddenMethods[methodName] {
					objExpr := expr.Expression()
					isForbidden = isGlobalReference(objExpr)
				}
			}

		case ast.KindElementAccessExpression:
			// Element access: window['alert'](), etc.
			elemAccess := expr.AsElementAccessExpression()
			if elemAccess != nil {
				objExpr := elemAccess.Expression
				if isGlobalReference(objExpr) && elemAccess.ArgumentExpression != nil {
					argExpr := elemAccess.ArgumentExpression
					if argExpr.Kind == ast.KindStringLiteral {
						// Remove quotes from string literal
						text := argExpr.Text()
						if len(text) >= 2 {
							methodName = text[1 : len(text)-1]
							isForbidden = forbiddenMethods[methodName]
						}
					}
				}
			}

		case ast.KindParenthesizedExpression:
			// Unwrap parentheses for optional chaining: (window?.alert)(foo)
			innerExpr := expr.Expression()
			if innerExpr != nil && innerExpr.Kind == ast.KindPropertyAccessExpression {
				propName := innerExpr.Name()
				if propName != nil && propName.Kind == ast.KindIdentifier {
					methodName = propName.Text()
					if forbiddenMethods[methodName] {
						objExpr := innerExpr.Expression()
						isForbidden = isGlobalReference(objExpr)
					}
				}
			}
		}

		if isForbidden {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "unexpected-" + methodName,
				Description: "Unexpected " + methodName + ".",
			})
		}
	}

	return rule.RuleListeners{
		ast.KindCallExpression: checkCallExpression,
	}
}
