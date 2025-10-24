package no_unsafe_assignment

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoUnsafeAssignmentRule implements the no-unsafe-assignment rule
// Disallow assigning a value with type `any` to variables and properties
var NoUnsafeAssignmentRule = rule.CreateRule(rule.Rule{
	Name: "no-unsafe-assignment",
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
			//         Id:          "anyAssignment",
			//         Description: "TODO: Add error message",
			//     })
			// }
		},
	}
}
