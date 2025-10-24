package prefer_string_starts_ends_with

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// PreferStringStartsEndsWithRule implements the prefer-string-starts-ends-with rule
// Enforce using `String#startsWith` and `String#endsWith` over other equivalent methods of checking substrings
var PreferStringStartsEndsWithRule = rule.CreateRule(rule.Rule{
	Name: "prefer-string-starts-ends-with",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {

	return rule.RuleListeners{
		ast.KindFunctionDeclaration: func(node *ast.Node) {
			// TODO: Implement rule logic for FunctionDeclaration

			// Example: Check some condition and report
			// if violatesRule(node) {
			//     ctx.ReportNode(node, rule.RuleMessage{
			//         Id:          "preferEndsWith",
			//         Description: "TODO: Add error message",
			//     })
			// }
		},
	}
}
