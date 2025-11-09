package no_dynamic_delete

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func buildDynamicDeleteMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "dynamicDelete",
		Description: "Do not delete dynamically computed property keys.",
	}
}

var NoDynamicDeleteRule = rule.CreateRule(rule.Rule{
	Name: "no-dynamic-delete",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindDeleteExpression: func(node *ast.Node) {
				if node.Kind != ast.KindDeleteExpression {
					return
				}
				deleteExpression := ast.SkipParentheses(node.AsDeleteExpression().Expression)

				// Only check element access expressions (e.g., obj[key])
				if !ast.IsElementAccessExpression(deleteExpression) {
					return
				}

				expression := deleteExpression.AsElementAccessExpression()
				argumentExpression := expression.ArgumentExpression

				if argumentExpression == nil {
					return
				}

				// Skip parentheses to get the actual argument
				arg := ast.SkipParentheses(argumentExpression)

				// Allow numeric literals (e.g., obj[7], obj[-7])
				if arg.Kind == ast.KindNumericLiteral {
					return
				}

				// Allow string literals (e.g., obj['aaa'])
				if arg.Kind == ast.KindStringLiteral {
					return
				}

				// Everything else is considered dynamic and should be reported
				// This includes:
				// - Identifiers (e.g., obj[name])
				// - Binary expressions (e.g., obj['aa' + 'b'])
				// - Prefix unary expressions (e.g., obj[+7], obj[-Infinity])
				// - Call expressions (e.g., obj[getName()])
				// - Property access expressions (e.g., obj[name.foo.bar])
				// - etc.
				ctx.ReportNode(node, buildDynamicDeleteMessage())
			},
		}
	},
})
