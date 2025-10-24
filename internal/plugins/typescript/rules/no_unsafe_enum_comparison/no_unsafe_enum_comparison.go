package no_unsafe_enum_comparison

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoUnsafeEnumComparisonRule implements the no-unsafe-enum-comparison rule
// Disallow comparing an enum value with a non-enum value
var NoUnsafeEnumComparisonRule = rule.CreateRule(rule.Rule{
	Name: "no-unsafe-enum-comparison",
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
			//         Id:          "mismatchedCase",
			//         Description: "TODO: Add error message",
			//     })
			// }
		},
	}
}
