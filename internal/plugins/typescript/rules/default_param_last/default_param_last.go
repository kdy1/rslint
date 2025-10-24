package default_param_last

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// DefaultParamLastRule implements the default-param-last rule
// Enforce default parameters to be last
var DefaultParamLastRule = rule.CreateRule(rule.Rule{
	Name: "default-param-last",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {

	return rule.RuleListeners{
		ast.KindFunctionDeclaration: func(node *ast.Node) {
			// TODO: Implement rule logic for FunctionDeclaration

			// Example: Check some condition and report
			// if violatesRule(node) {
			//     ctx.ReportNode(node, rule.RuleMessage{
			//         Id:          "shouldBeLast",
			//         Description: "TODO: Add error message",
			//     })
			// }
		},
	}
}
