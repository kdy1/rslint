package no_throw_literal

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoThrowLiteralRule implements the no-throw-literal rule
// TODO: Add description for no-throw-literal rule
var NoThrowLiteralRule = rule.CreateRule(rule.Rule{
	Name: "no-throw-literal",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {

	return rule.RuleListeners{
		ast.KindFunctionDeclaration: func(node *ast.Node) {
			// TODO: Implement rule logic for FunctionDeclaration

			// Example: Check some condition and report
			// if violatesRule(node) {
			//     ctx.ReportNode(node, rule.RuleMessage{
			//         Id:          "default",
			//         Description: "TODO: Add error message",
			//     })
			// }
		},
	}
}
