package prefer_readonly

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// PreferReadonlyRule implements the prefer-readonly rule
// Require private members to be marked as `readonly` if they
var PreferReadonlyRule = rule.CreateRule(rule.Rule{
	Name: "prefer-readonly",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {

	return rule.RuleListeners{
		ast.KindFunctionDeclaration: func(node *ast.Node) {
			// TODO: Implement rule logic for FunctionDeclaration

			// Example: Check some condition and report
			// if violatesRule(node) {
			//     ctx.ReportNode(node, rule.RuleMessage{
			//         Id:          "preferReadonly",
			//         Description: "TODO: Add error message",
			//     })
			// }
		},
	}
}
