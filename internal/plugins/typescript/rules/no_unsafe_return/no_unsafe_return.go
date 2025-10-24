package no_unsafe_return

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoUnsafeReturnRule implements the no-unsafe-return rule
// Disallow returning a value with type `any` from a function
var NoUnsafeReturnRule = rule.CreateRule(rule.Rule{
	Name: "no-unsafe-return",
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
			//         Id:          "unsafeReturn",
			//         Description: "TODO: Add error message",
			//     })
			// }
		},
	}
}
