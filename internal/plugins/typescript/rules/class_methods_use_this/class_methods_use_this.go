package class_methods_use_this

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// ClassMethodsUseThisRule implements the class-methods-use-this rule
// Enforce that class methods utilize `this`
var ClassMethodsUseThisRule = rule.CreateRule(rule.Rule{
	Name: "class-methods-use-this",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {

	return rule.RuleListeners{
		ast.KindFunctionDeclaration: func(node *ast.Node) {
			// TODO: Implement rule logic for FunctionDeclaration

			// Example: Check some condition and report
			// if violatesRule(node) {
			//     ctx.ReportNode(node, rule.RuleMessage{
			//         Id:          "missingThis",
			//         Description: "TODO: Add error message",
			//     })
			// }
		},
	}
}
