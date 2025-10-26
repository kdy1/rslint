package no_obj_calls

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoObjCallsRule implements the no-obj-calls rule
// Disallow calling global object properties as functions
var NoObjCallsRule = rule.Rule{
	Name: "no-obj-calls",
	Run:  run,
}

// Non-callable global objects that should not be invoked as functions
var nonCallables = map[string]bool{
	"Math":    true,
	"JSON":    true,
	"Reflect": true,
	"Atomics": true,
	"Intl":    true,
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	return rule.RuleListeners{
		// Handle direct function calls like Math()
		ast.KindCallExpression: func(node *ast.Node) {
			checkCallExpression(ctx, node, false)
		},

		// Handle new expressions like new Math()
		ast.KindNewExpression: func(node *ast.Node) {
			checkCallExpression(ctx, node, true)
		},
	}
}

func checkCallExpression(ctx rule.RuleContext, node *ast.Node, isNew bool) {
	if node == nil {
		return
	}

	expr := node.Expression()
	if expr == nil {
		return
	}

	// Check if it's a direct call to a non-callable global
	// e.g., Math() or JSON()
	if expr.Kind == ast.KindIdentifier {
		name := expr.Text()
		if nonCallables[name] {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "unexpectedCall",
				Description: "'" + name + "' is not a function.",
			})
		}
		return
	}

	// Check for globalThis.Math() pattern
	// This includes both regular property access and optional chaining
	if expr.Kind == ast.KindPropertyAccessExpression {
		checkGlobalThisPropertyCall(ctx, node, expr)
	}
}

func checkGlobalThisPropertyCall(ctx rule.RuleContext, callNode *ast.Node, propAccess *ast.Node) {
	obj := propAccess.Expression()
	if obj == nil {
		return
	}

	name := propAccess.Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return
	}

	propertyName := name.Text()

	// Check if the property is a non-callable
	if !nonCallables[propertyName] {
		return
	}

	// Check if it's globalThis.NonCallable()
	if obj.Kind == ast.KindIdentifier && obj.Text() == "globalThis" {
		ctx.ReportNode(callNode, rule.RuleMessage{
			Id:          "unexpectedCall",
			Description: "'" + propertyName + "' is not a function.",
		})
	}
}
