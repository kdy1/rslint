package no_unnecessary_type_constraint

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoUnnecessaryTypeConstraintRule implements the no-unnecessary-type-constraint rule
// Disallow unnecessary constraints on generic types
var NoUnnecessaryTypeConstraintRule = rule.CreateRule(rule.Rule{
	Name: "no-unnecessary-type-constraint",
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
			//         Id:          "removeUnnecessaryConstraint",
			//         Description: "TODO: Add error message",
			//     })
			// }
		},
	}
}
