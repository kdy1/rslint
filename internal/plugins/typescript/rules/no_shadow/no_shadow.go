package no_shadow

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoShadowRule implements the no-shadow rule
// Disallow variable declarations from shadowing variables declared in the outer scope
var NoShadowRule = rule.CreateRule(rule.Rule{
	Name: "no-shadow",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {

	return rule.RuleListeners{
		ast.KindFunctionDeclaration: func(node *ast.Node) {
			// TODO: Implement rule logic for FunctionDeclaration

			// Example: Check some condition and report
			// if violatesRule(node) {
			//     ctx.ReportNode(node, rule.RuleMessage{
			//         Id:          "noShadow",
			//         Description: "TODO: Add error message",
			//     })
			// }
		},
	}
}
