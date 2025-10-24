package no_unnecessary_qualifier

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoUnnecessaryQualifierRule implements the no-unnecessary-qualifier rule
// Disallows unnecessary namespace qualifiers
var NoUnnecessaryQualifierRule = rule.CreateRule(rule.Rule{
	Name: "no-unnecessary-qualifier",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	return rule.RuleListeners{
		ast.KindQualifiedName: func(node *ast.Node) {
			// This rule requires type information
			if ctx.TypeChecker == nil {
				return
			}

			// TODO: Implement logic to detect unnecessary namespace qualifiers
			// 1. Get the qualified name (e.g., Namespace.Member)
			// 2. Check if the member can be accessed without the qualifier
			// 3. Use type checker to verify the member is in scope
			// 4. Report if qualifier is unnecessary
		},
		ast.KindPropertyAccessExpression: func(node *ast.Node) {
			// This rule requires type information
			if ctx.TypeChecker == nil {
				return
			}

			// TODO: Check for unnecessary enum member qualifiers
			// e.g., MyEnum.Member when Member is already in scope
		},
	}
}
