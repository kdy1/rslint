package no_magic_numbers

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoMagicNumbersRule implements the no-magic-numbers rule
// Whether enums used in TypeScript are considered okay.
var NoMagicNumbersRule = rule.CreateRule(rule.Rule{
	Name: "no-magic-numbers",
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
