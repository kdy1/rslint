package explicit_module_boundary_types

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// ExplicitModuleBoundaryTypesRule implements the explicit-module-boundary-types rule
// Require explicit return and argument types on exported functions
var ExplicitModuleBoundaryTypesRule = rule.CreateRule(rule.Rule{
	Name: "explicit-module-boundary-types",
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
			//         Id:          "anyTypedArg",
			//         Description: "TODO: Add error message",
			//     })
			// }
		},
	}
}
