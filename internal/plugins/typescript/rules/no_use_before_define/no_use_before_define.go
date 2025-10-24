package no_use_before_define

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoUseBeforeDefineRule implements the no-use-before-define rule
// Disallow the use of variables before they are defined
var NoUseBeforeDefineRule = rule.CreateRule(rule.Rule{
	Name: "no-use-before-define",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {

	return rule.RuleListeners{
		ast.KindFunctionDeclaration: func(node *ast.Node) {
			// TODO: Implement rule logic for FunctionDeclaration

			// Example: Check some condition and report
			// if violatesRule(node) {
			//     ctx.ReportNode(node, rule.RuleMessage{
			//         Id:          "noUseBeforeDefine",
			//         Description: "TODO: Add error message",
			//     })
			// }
		},
	}
}
