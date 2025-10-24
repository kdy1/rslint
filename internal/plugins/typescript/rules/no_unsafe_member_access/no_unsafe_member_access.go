package no_unsafe_member_access

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoUnsafeMemberAccessRule implements the no-unsafe-member-access rule
// Disallow member access on a value with type `any`
var NoUnsafeMemberAccessRule = rule.CreateRule(rule.Rule{
	Name: "no-unsafe-member-access",
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
			//         Id:          "unsafeComputedMemberAccess",
			//         Description: "TODO: Add error message",
			//     })
			// }
		},
	}
}
