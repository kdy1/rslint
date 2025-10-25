package no_unexpected_multiline

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoUnexpectedMultilineRule implements the no-unexpected-multiline rule
// Disallow confusing multiline expressions
var NoUnexpectedMultilineRule = rule.Rule{
	Name: "no-unexpected-multiline",
	Run:  run,
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	sourceFile := ctx.SourceFile
	if sourceFile == nil {
		return rule.RuleListeners{}
	}

	return rule.RuleListeners{
		ast.KindCallExpression: func(node *ast.Node) {
			callExpr := node.AsCallExpression()
			if callExpr == nil || callExpr.Expression == nil {
				return
			}

			// Check if the opening paren is on a new line from the callee
			if isOnNewLine(sourceFile, callExpr.Expression, node) {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "function",
					Description: "Unexpected newline between function and ( of function call.",
				})
			}
		},
		ast.KindPropertyAccessExpression: func(node *ast.Node) {
			propAccess := node.AsPropertyAccessExpression()
			if propAccess == nil || propAccess.Expression == nil {
				return
			}

			// Check if this is an element access (bracket notation)
			// In TypeScript AST, PropertyAccessExpression is for dot notation
			// ElementAccessExpression is for bracket notation
			// This handler is for PropertyAccessExpression, skip it
		},
		ast.KindElementAccessExpression: func(node *ast.Node) {
			elemAccess := node.AsElementAccessExpression()
			if elemAccess == nil || elemAccess.Expression == nil {
				return
			}

			// Check if the opening bracket is on a new line from the object
			if isOnNewLine(sourceFile, elemAccess.Expression, node) {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "property",
					Description: "Unexpected newline between object and [ of property access.",
				})
			}
		},
		ast.KindTaggedTemplateExpression: func(node *ast.Node) {
			taggedTemplate := node.AsTaggedTemplateExpression()
			if taggedTemplate == nil || taggedTemplate.Tag == nil || taggedTemplate.Template == nil {
				return
			}

			// Check if the template is on a new line from the tag
			if isOnNewLine(sourceFile, taggedTemplate.Tag, taggedTemplate.Template) {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "taggedTemplate",
					Description: "Unexpected newline between template tag and template literal.",
				})
			}
		},
		ast.KindBinaryExpression: func(node *ast.Node) {
			binExpr := node.AsBinaryExpression()
			if binExpr == nil || binExpr.Left == nil || binExpr.Right == nil {
				return
			}

			// Check for division operator that might be confused with regex
			if binExpr.OperatorToken != nil && binExpr.OperatorToken.Kind == ast.KindSlashToken {
				// Check if the right side is on a new line
				if isOnNewLine(sourceFile, binExpr.Left, binExpr.Right) {
					ctx.ReportNode(binExpr.Right, rule.RuleMessage{
						Id:          "division",
						Description: "Unexpected newline between numerator and division operator.",
					})
				}
			}
		},
	}
}

// isOnNewLine checks if node2 starts on a different line than node1 ends
func isOnNewLine(sourceFile *ast.SourceFile, node1 *ast.Node, node2 interface{}) bool {
	if sourceFile == nil || node1 == nil || node2 == nil {
		return false
	}

	// Get end position of node1
	end1 := node1.End()

	// Get start position of node2
	var start2 int
	switch v := node2.(type) {
	case *ast.Node:
		start2 = v.Pos()
	default:
		return false
	}

	// Get line and character for both positions
	line1, _ := scanner.GetLineAndCharacterOfPosition(sourceFile, end1)
	line2, _ := scanner.GetLineAndCharacterOfPosition(sourceFile, start2)

	return line2 > line1
}
