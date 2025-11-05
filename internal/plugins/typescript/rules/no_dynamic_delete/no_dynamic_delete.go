package no_dynamic_delete

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildDynamicDeleteMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "dynamicDelete",
		Description: "Using the `delete` operator with a computed key expression is unsafe.",
	}
}

var NoDynamicDeleteRule = rule.CreateRule(rule.Rule{
	Name: "no-dynamic-delete",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		isStaticProperty := func(argumentExpression *ast.Node) bool {
			// Check if the argument is a static literal or identifier that can be computed at compile time
			switch argumentExpression.Kind {
			case ast.KindNumericLiteral:
				// Numeric literals are allowed (e.g., delete obj[7])
				return true
			case ast.KindStringLiteral:
				// String literals are allowed (e.g., delete obj['key'])
				return true
			case ast.KindPrefixUnaryExpression:
				// Check for negative numeric literals (e.g., delete obj[-7])
				unaryExpr := argumentExpression.AsPrefixUnaryExpression()
				if unaryExpr.Operator == scanner.TokenMinusToken {
					if unaryExpr.Operand.Kind == ast.KindNumericLiteral {
						return true
					}
				}
				// All other prefix unary expressions are dynamic (e.g., +7, +Infinity, typeof)
				return false
			case ast.KindIdentifier:
				// Identifiers that refer to runtime values are dynamic (e.g., delete obj[name])
				// However, we need to check for special cases like Infinity, NaN
				text := argumentExpression.AsIdentifier().Text.Text()
				// These are special identifiers that are considered static in the context of property access
				// Actually, based on the test cases, Infinity and NaN without operators are dynamic
				return false
			default:
				// All other expressions are considered dynamic
				return false
			}
		}

		return rule.RuleListeners{
			ast.KindDeleteExpression: func(node *ast.Node) {
				if node.Kind != ast.KindDeleteExpression {
					return
				}
				deleteExpression := ast.SkipParentheses(node.AsDeleteExpression().Expression)

				// Only check element access expressions (obj[key])
				if !ast.IsElementAccessExpression(deleteExpression) {
					return
				}

				expression := deleteExpression.AsElementAccessExpression()
				argumentExpression := ast.SkipParentheses(expression.ArgumentExpression)

				// If the argument is not a static property, report it
				if !isStaticProperty(argumentExpression) {
					ctx.ReportNode(node, buildDynamicDeleteMessage())
				}
			},
		}
	},
})
