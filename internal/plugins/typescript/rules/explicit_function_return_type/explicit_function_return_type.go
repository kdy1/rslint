package explicit_function_return_type

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// ExplicitFunctionReturnTypeRule implements the explicit-function-return-type rule
// Require explicit return types on functions and class methods
var ExplicitFunctionReturnTypeRule = rule.CreateRule(rule.Rule{
	Name: "explicit-function-return-type",
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
			//         Id:          "missingReturnType",
			//         Description: "TODO: Add error message",
			//     })
			// }
		},
	}
}
