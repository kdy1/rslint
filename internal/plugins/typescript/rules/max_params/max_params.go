package max_params

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// MaxParamsRule implements the max-params rule
// Enforce a maximum number of parameters in function definitions
var MaxParamsRule = rule.CreateRule(rule.Rule{
	Name: "max-params",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {

	return rule.RuleListeners{
		ast.KindFunctionDeclaration: func(node *ast.Node) {
			// TODO: Implement rule logic for FunctionDeclaration

			// Example: Check some condition and report
			// if violatesRule(node) {
			//     ctx.ReportNode(node, rule.RuleMessage{
			//         Id:          "default",
			//         Description: "TODO: Add error message",
			//     })
			// }
		},
	}
}
