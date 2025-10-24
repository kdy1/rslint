package no_redundant_type_constituents

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoRedundantTypeConstituentsRule implements the no-redundant-type-constituents rule
// Disallow members of unions and intersections that do nothing or override type information
var NoRedundantTypeConstituentsRule = rule.CreateRule(rule.Rule{
	Name: "no-redundant-type-constituents",
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
			//         Id:          "errorTypeOverrides",
			//         Description: "TODO: Add error message",
			//     })
			// }
		},
	}
}
