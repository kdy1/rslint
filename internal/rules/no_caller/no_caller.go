package no_caller

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoCallerRule implements the no-caller rule
// Disallow the use of `arguments.caller` or `arguments.callee`
var NoCallerRule = rule.CreateRule(rule.Rule{
	Name: "no-caller",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	checkPropertyAccess := func(node *ast.Node) {
		if node == nil {
			return
		}

		// Get the object being accessed
		expr := node.Expression()
		if expr == nil || expr.Kind != ast.KindIdentifier {
			return
		}

		// Check if it's accessing the arguments object
		if expr.Text() != "arguments" {
			return
		}

		// Get the property being accessed
		propName := node.Name()
		if propName == nil || propName.Kind != ast.KindIdentifier {
			return
		}

		prop := propName.Text()

		// Check if it's accessing caller or callee
		if prop == "caller" || prop == "callee" {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "unexpected",
				Description: "Avoid arguments." + prop + ".",
			})
		}
	}

	return rule.RuleListeners{
		ast.KindPropertyAccessExpression: checkPropertyAccess,
	}
}
