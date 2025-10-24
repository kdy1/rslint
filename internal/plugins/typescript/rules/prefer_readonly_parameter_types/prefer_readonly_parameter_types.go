package prefer_readonly_parameter_types

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// PreferReadonlyParameterTypesRule implements the prefer-readonly-parameter-types rule
// Require function parameters to be typed as `readonly` to prevent accidental mutation of inputs
var PreferReadonlyParameterTypesRule = rule.CreateRule(rule.Rule{
	Name: "prefer-readonly-parameter-types",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {

	return rule.RuleListeners{
		ast.KindFunctionDeclaration: func(node *ast.Node) {
			// TODO: Implement rule logic for FunctionDeclaration

			// Example: Check some condition and report
			// if violatesRule(node) {
			//     ctx.ReportNode(node, rule.RuleMessage{
			//         Id:          "shouldBeReadonly",
			//         Description: "TODO: Add error message",
			//     })
			// }
		},
	}
}
