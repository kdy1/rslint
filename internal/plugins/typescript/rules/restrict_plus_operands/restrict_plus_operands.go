package restrict_plus_operands

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// RestrictPlusOperandsRule implements the restrict-plus-operands rule
// Require both operands of addition to be the same type and be `bigint`, `number`, or `string`
var RestrictPlusOperandsRule = rule.CreateRule(rule.Rule{
	Name: "restrict-plus-operands",
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
			//         Id:          "bigintAndNumber",
			//         Description: "TODO: Add error message",
			//     })
			// }
		},
	}
}
