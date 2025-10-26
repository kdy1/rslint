package no_delete_var

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func buildUnexpectedMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpected",
		Description: "Variables should not be deleted.",
	}
}

// NoDeleteVarRule implements the no-delete-var rule
// Disallow deleting variables
var NoDeleteVarRule = rule.CreateRule(rule.Rule{
	Name: "no-delete-var",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindDeleteExpression: func(node *ast.Node) {
				if node.Kind != ast.KindDeleteExpression {
					return
				}

				deleteExpression := node.AsDeleteExpression()
				if deleteExpression == nil {
					return
				}

				// Get the expression being deleted
				expr := ast.SkipParentheses(deleteExpression.Expression)
				if expr == nil {
					return
				}

				// Check if it's an identifier (variable reference)
				// Deleting identifiers is not allowed, but deleting properties is fine
				if expr.Kind == ast.KindIdentifier {
					ctx.ReportNode(node, buildUnexpectedMessage())
				}
			},
		}
	},
})
