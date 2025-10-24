package prefer_regexp_exec

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// PreferRegexpExecRule implements the prefer-regexp-exec rule
// Enforce `RegExp#exec` over `String#match` if no global flag is provided
var PreferRegexpExecRule = rule.CreateRule(rule.Rule{
	Name: "prefer-regexp-exec",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {

	return rule.RuleListeners{
		ast.KindFunctionDeclaration: func(node *ast.Node) {
			// TODO: Implement rule logic for FunctionDeclaration

			// Example: Check some condition and report
			// if violatesRule(node) {
			//     ctx.ReportNode(node, rule.RuleMessage{
			//         Id:          "regExpExecOverStringMatch",
			//         Description: "TODO: Add error message",
			//     })
			// }
		},
	}
}
