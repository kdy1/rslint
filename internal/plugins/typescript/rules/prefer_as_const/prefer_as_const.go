package prefer_as_const

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// PreferAsConstRule implements the prefer-as-const rule
// Enforce the use of `as const` over literal type
var PreferAsConstRule = rule.CreateRule(rule.Rule{
	Name: "prefer-as-const",
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
			//         Id:          "preferConstAssertion",
			//         Description: "TODO: Add error message",
			//     })
			// }
		},
	}
}
