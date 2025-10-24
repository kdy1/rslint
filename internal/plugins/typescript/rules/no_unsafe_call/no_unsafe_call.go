package no_unsafe_call

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoUnsafeCallRule implements the no-unsafe-call rule
// Disallow calling a value with type `any`
var NoUnsafeCallRule = rule.CreateRule(rule.Rule{
	Name: "no-unsafe-call",
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
			//         Id:          "unsafeCall",
			//         Description: "TODO: Add error message",
			//     })
			// }
		},
	}
}
