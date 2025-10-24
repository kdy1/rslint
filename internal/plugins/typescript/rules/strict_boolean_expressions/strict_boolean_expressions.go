package strict_boolean_expressions

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// StrictBooleanExpressionsRule implements the strict-boolean-expressions rule
// Disallow certain types in boolean expressions
var StrictBooleanExpressionsRule = rule.CreateRule(rule.Rule{
	Name: "strict-boolean-expressions",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {

	return rule.RuleListeners{
		ast.KindFunctionDeclaration: func(node *ast.Node) {
			// TODO: Implement rule logic for FunctionDeclaration

			// Example: Check some condition and report
			// if violatesRule(node) {
			//     ctx.ReportNode(node, rule.RuleMessage{
			//         Id:          "conditionErrorAny",
			//         Description: "TODO: Add error message",
			//     })
			// }
		},
	}
}
