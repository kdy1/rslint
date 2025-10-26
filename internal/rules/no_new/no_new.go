package no_new

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoNewRule implements the no-new rule
// Disallow `new` operators outside of assignments or comparisons
var NoNewRule = rule.CreateRule(rule.Rule{
	Name: "no-new",
	Run:  run,
})

func buildNoNewMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noNewStatement",
		Description: "Do not use 'new' for side effects.",
	}
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	return rule.RuleListeners{
		ast.KindExpressionStatement: func(node *ast.Node) {
			if node == nil {
				return
			}

			// Check if the expression is a NewExpression
			expr := node.Expression()
			if expr == nil || expr.Kind != ast.KindNewExpression {
				return
			}

			// Report the violation
			ctx.ReportNode(node, buildNoNewMessage())
		},
	}
}
