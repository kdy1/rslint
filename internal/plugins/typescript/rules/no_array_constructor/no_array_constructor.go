package no_array_constructor

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoArrayConstructorRule implements the no-array-constructor rule
// Disallow generic `Array` constructors
var NoArrayConstructorRule = rule.CreateRule(rule.Rule{
	Name: "no-array-constructor",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {

	return rule.RuleListeners{
		ast.KindFunctionDeclaration: func(node *ast.Node) {
			// TODO: Implement rule logic for FunctionDeclaration

			// Example: Check some condition and report
			// if violatesRule(node) {
			//     ctx.ReportNode(node, rule.RuleMessage{
			//         Id:          "useLiteral",
			//         Description: "TODO: Add error message",
			//     })
			// }
		},
	}
}
