package prefer_includes

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// PreferIncludesRule implements the prefer-includes rule
// Enforce `includes` method over `indexOf` method
var PreferIncludesRule = rule.CreateRule(rule.Rule{
	Name: "prefer-includes",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {

	return rule.RuleListeners{
		ast.KindFunctionDeclaration: func(node *ast.Node) {
			// TODO: Implement rule logic for FunctionDeclaration

			// Example: Check some condition and report
			// if violatesRule(node) {
			//     ctx.ReportNode(node, rule.RuleMessage{
			//         Id:          "preferIncludes",
			//         Description: "TODO: Add error message",
			//     })
			// }
		},
	}
}
