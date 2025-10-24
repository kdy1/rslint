package explicit_member_accessibility

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// ExplicitMemberAccessibilityRule implements the explicit-member-accessibility rule
// Require explicit accessibility modifiers on class properties and methods
var ExplicitMemberAccessibilityRule = rule.CreateRule(rule.Rule{
	Name: "explicit-member-accessibility",
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
			//         Id:          "addExplicitAccessibility",
			//         Description: "TODO: Add error message",
			//     })
			// }
		},
	}
}
