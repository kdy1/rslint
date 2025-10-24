package no_redeclare

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoRedeclareRule implements the no-redeclare rule
// Disallow variable redeclaration
var NoRedeclareRule = rule.CreateRule(rule.Rule{
	Name: "no-redeclare",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {

	return rule.RuleListeners{
		ast.KindFunctionDeclaration: func(node *ast.Node) {
			// TODO: Implement rule logic for FunctionDeclaration

			// Example: Check some condition and report
			// if violatesRule(node) {
			//     ctx.ReportNode(node, rule.RuleMessage{
			//         Id:          "redeclared",
			//         Description: "TODO: Add error message",
			//     })
			// }
		},
	}
}
