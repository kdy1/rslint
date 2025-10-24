package no_unsafe_argument

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoUnsafeArgumentRule implements the no-unsafe-argument rule
// Disallow calling a function with a value with type `any`
var NoUnsafeArgumentRule = rule.CreateRule(rule.Rule{
	Name: "no-unsafe-argument",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {

	return rule.RuleListeners{
		ast.KindFunctionDeclaration: func(node *ast.Node) {
			// TODO: Implement rule logic for FunctionDeclaration
			// This rule requires type information
			if ctx.TypeChecker == nil {
				return
			}

			// Example: Check some condition and report
			// if violatesRule(node) {
			//     ctx.ReportNode(node, rule.RuleMessage{
			//         Id:          "unsafeArgument",
			//         Description: "TODO: Add error message",
			//     })
			// }
		},
	}
}
