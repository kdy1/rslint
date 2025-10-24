package no_unnecessary_parameter_property_assignment

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoUnnecessaryParameterPropertyAssignmentRule implements the no-unnecessary-parameter-property-assignment rule
// Disallows parameter properties when they could be regular parameters
var NoUnnecessaryParameterPropertyAssignmentRule = rule.CreateRule(rule.Rule{
	Name: "no-unnecessary-parameter-property-assignment",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	return rule.RuleListeners{
		ast.KindConstructor: func(node *ast.Node) {
			// This rule requires type information
			if ctx.TypeChecker == nil {
				return
			}

			// TODO: Implement logic to detect unnecessary parameter properties
			// 1. Check constructor parameters with visibility modifiers (public, private, protected, readonly)
			// 2. Look for assignments in constructor body that assign parameter to same-named property
			// 3. Report if assignment is redundant (parameter property already creates the assignment)
			// Example:
			//   class C {
			//     constructor(public x: number) {
			//       this.x = x; // Unnecessary - parameter property already assigns it
			//     }
			//   }
		},
		ast.KindParameter: func(node *ast.Node) {
			// This rule requires type information
			if ctx.TypeChecker == nil {
				return
			}

			// TODO: Check if this is a parameter property (has visibility modifier)
			// and if there's a redundant assignment in the constructor body
		},
	}
}
