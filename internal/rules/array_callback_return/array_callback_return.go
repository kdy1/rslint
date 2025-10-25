package array_callback_return

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// ArrayCallbackReturnRule implements the array-callback-return rule
// Enforce `return` statements in callbacks of array methods
var ArrayCallbackReturnRule = rule.Rule{
	Name: "array-callback-return",
	Run:  run,
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	return rule.RuleListeners{
		ast.KindCallExpression: func(node *ast.Node) {
			callExpr := node.AsCallExpression()
			if callExpr == nil || callExpr.Expression == nil {
				return
			}

			// Basic implementation stub for array-callback-return rule
			// This is a minimal implementation that validates array method callbacks
			// TODO: Implement full logic to check:
			// 1. Identify array methods that require return statements (map, filter, reduce, etc.)
			// 2. Check if the callback function has appropriate return statements
			// 3. Validate that all code paths return a value
			// 4. Support options: { allowImplicit: boolean, checkForEach: boolean }

			// For now, just validate the node structure exists
			// A complete implementation would:
			// - Check if this is a call to an array method (e.g., arr.map(...))
			// - Analyze the callback function argument
			// - Ensure return statements are present in all branches
			// - Handle arrow functions with implicit returns
		},
	}
}
