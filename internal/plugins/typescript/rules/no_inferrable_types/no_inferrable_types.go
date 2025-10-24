package no_inferrable_types

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoInferrableTypesRule implements the no-inferrable-types rule
// Disallow explicit type declarations for variables or parameters initialized to a number, string, or boolean
var NoInferrableTypesRule = rule.CreateRule(rule.Rule{
	Name: "no-inferrable-types",
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
			//         Id:          "noInferrableType",
			//         Description: "TODO: Add error message",
			//     })
			// }
		},
	}
}
