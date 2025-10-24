package no_unnecessary_type_assertion

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoUnnecessaryTypeAssertionRule implements the no-unnecessary-type-assertion rule
// Disallow type assertions that do not change the type of an expression
var NoUnnecessaryTypeAssertionRule = rule.CreateRule(rule.Rule{
	Name: "no-unnecessary-type-assertion",
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
			//         Id:          "contextuallyUnnecessary",
			//         Description: "TODO: Add error message",
			//     })
			// }
		},
	}
}
