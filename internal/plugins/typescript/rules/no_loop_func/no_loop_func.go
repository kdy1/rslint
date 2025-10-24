package no_loop_func

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoLoopFuncRule implements the no-loop-func rule
// Disallow function declarations that contain unsafe references inside loop statements
var NoLoopFuncRule = rule.CreateRule(rule.Rule{
	Name: "no-loop-func",
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
