package no_new_func

import (
	"github.com/microsoft/typescript-go/shim/ast"

	"github.com/web-infra-dev/rslint/internal/rule"
)

func buildNoFunctionConstructorMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noFunctionConstructor",
		Description: "The Function constructor is eval.",
	}
}

// isFunctionIdentifier checks if a node is the global Function identifier
func isFunctionIdentifier(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindIdentifier {
		return false
	}

	ident := node.AsIdentifier()
	if ident == nil {
		return false
	}

	return ident.Text == "Function"
}

// NoNewFuncRule disallows new Function(...)
var NoNewFuncRule = rule.CreateRule(rule.Rule{
	Name: "no-new-func",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		listeners := rule.RuleListeners{}

		// Check for new Function() and Function() calls
		listeners[ast.KindNewExpression] = func(node *ast.Node) {
			newExpr := node.AsNewExpression()
			if newExpr == nil || newExpr.Expression == nil {
				return
			}

			if isFunctionIdentifier(newExpr.Expression) {
				ctx.ReportNode(newExpr.Expression, buildNoFunctionConstructorMessage())
			}
		}

		listeners[ast.KindCallExpression] = func(node *ast.Node) {
			callExpr := node.AsCallExpression()
			if callExpr == nil || callExpr.Expression == nil {
				return
			}

			// Direct Function() call
			if isFunctionIdentifier(callExpr.Expression) {
				ctx.ReportNode(callExpr.Expression, buildNoFunctionConstructorMessage())
				return
			}

			// Function.call(), Function.apply(), Function.bind()
			if callExpr.Expression.Kind == ast.KindPropertyAccessExpression {
				propAccess := callExpr.Expression.AsPropertyAccessExpression()
				if propAccess != nil && propAccess.Expression != nil && propAccess.Name() != nil {
					if isFunctionIdentifier(propAccess.Expression) {
						methodName := propAccess.Name().Text()
						if methodName == "call" || methodName == "apply" || methodName == "bind" {
							ctx.ReportNode(propAccess.Expression, buildNoFunctionConstructorMessage())
						}
					}
				}
			}

			// Function?.call() with optional chaining
			if callExpr.Expression.Kind == ast.KindPropertyAccessExpression {
				propAccess := callExpr.Expression.AsPropertyAccessExpression()
				if propAccess != nil && propAccess.QuestionDotToken != nil && propAccess.Expression != nil {
					if isFunctionIdentifier(propAccess.Expression) {
						ctx.ReportNode(propAccess.Expression, buildNoFunctionConstructorMessage())
					}
				}
			}

			// Check for parenthesized expressions like (Function?.call)
			if callExpr.Expression.Kind == ast.KindParenthesizedExpression {
				parenExpr := callExpr.Expression.AsParenthesizedExpression()
				if parenExpr != nil && parenExpr.Expression != nil {
					if parenExpr.Expression.Kind == ast.KindPropertyAccessExpression {
						propAccess := parenExpr.Expression.AsPropertyAccessExpression()
						if propAccess != nil && propAccess.Expression != nil {
							if isFunctionIdentifier(propAccess.Expression) {
								methodName := ""
								if propAccess.Name() != nil {
									methodName = propAccess.Name().Text()
								}
								if methodName == "call" || methodName == "apply" || methodName == "bind" {
									ctx.ReportNode(propAccess.Expression, buildNoFunctionConstructorMessage())
								}
							}
						}
					}
				}
			}
		}

		// Check for Function["call"] and similar bracket notation
		listeners[ast.KindElementAccessExpression] = func(node *ast.Node) {
			elemAccess := node.AsElementAccessExpression()
			if elemAccess == nil || elemAccess.Expression == nil || elemAccess.ArgumentExpression == nil {
				return
			}

			// Check if object is Function identifier
			if !isFunctionIdentifier(elemAccess.Expression) {
				return
			}

			// Check if we're accessing call, apply, or bind
			if elemAccess.ArgumentExpression.Kind == ast.KindStringLiteral {
				strLit := elemAccess.ArgumentExpression.AsStringLiteral()
				if strLit != nil {
					text := strLit.Text
					// Strip quotes
					if len(text) >= 2 && (text[0] == '"' || text[0] == '\'') {
						text = text[1 : len(text)-1]
					}
					if text == "call" || text == "apply" || text == "bind" {
						// This is Function["call"] etc., check if it's being called
						parent := node.Parent
						if parent != nil && parent.Kind == ast.KindCallExpression {
							ctx.ReportNode(elemAccess.Expression, buildNoFunctionConstructorMessage())
						}
					}
				}
			}
		}

		return listeners
	},
})
