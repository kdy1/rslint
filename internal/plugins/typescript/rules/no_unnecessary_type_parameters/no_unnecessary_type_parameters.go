package no_unnecessary_type_parameters

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoUnnecessaryTypeParametersRule implements the no-unnecessary-type-parameters rule
// Disallow type parameters that aren
var NoUnnecessaryTypeParametersRule = rule.CreateRule(rule.Rule{
	Name: "no-unnecessary-type-parameters",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {

	return rule.RuleListeners{
		ast.KindFunctionDeclaration: func(node *ast.Node) {
			// TODO: Implement rule logic for FunctionDeclaration

			// Example: Check some condition and report
			// if violatesRule(node) {
			//     ctx.ReportNode(node, rule.RuleMessage{
			//         Id:          "replaceUsagesWithConstraint",
			//         Description: "TODO: Add error message",
			//     })
			// }
		},
	}
}
