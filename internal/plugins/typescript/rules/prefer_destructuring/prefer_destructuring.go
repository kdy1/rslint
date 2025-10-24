package prefer_destructuring

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// PreferDestructuringRule implements the prefer-destructuring rule
// Whether to enforce destructuring on variable declarations with type annotations.
var PreferDestructuringRule = rule.CreateRule(rule.Rule{
	Name: "prefer-destructuring",
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
