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

// nonCallableGlobals lists the global objects that cannot be called as functions
var nonCallableGlobals = map[string]bool{
	"Math":    true,
	"JSON":    true,
	"Reflect": true,
	"Atomics": true,
	"Intl":    true,
}

// getIdentifierName extracts the identifier name from various expression types
func getIdentifierName(node *ast.Node) string {
	if node == nil {
		return ""
	}

	switch node.Kind {
	case ast.KindIdentifier:
		if ident := node.AsIdentifier(); ident != nil {
			return ident.Text
		}
	case ast.KindPropertyAccessExpression:
		// For globalThis.Math, etc.
		if pae := node.AsPropertyAccessExpression(); pae != nil {
			if expr := pae.Expression; expr != nil && expr.Kind == ast.KindIdentifier {
				if ident := expr.AsIdentifier(); ident != nil && ident.Text == "globalThis" {
					if name := pae.Name(); name != nil {
						return name.Text()
					}
				}
			}
		}
	}
	return ""
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	checkCall := func(node *ast.Node, isNew bool) {
		var expression *ast.Node

		if isNew {
			if newExpr := node.AsNewExpression(); newExpr != nil {
				expression = newExpr.Expression
			}
		} else {
			if callExpr := node.AsCallExpression(); callExpr != nil {
				expression = callExpr.Expression
			}
		}

		if expression == nil {
			return
		}

		name := getIdentifierName(expression)
		if name == "" {
			return
		}

		if !nonCallableGlobals[name] {
			return
		}

		// Report the error
		msgId := "unexpectedCall"
		description := "'" + name + "' is not a function."

		ctx.ReportNode(node, rule.RuleMessage{
			Id:          msgId,
			Description: description,
		})
	}

	return rule.RuleListeners{
		ast.KindCallExpression: func(node *ast.Node) {
			checkCall(node, false)
		},
		ast.KindNewExpression: func(node *ast.Node) {
			checkCall(node, true)
		},
	}
}
