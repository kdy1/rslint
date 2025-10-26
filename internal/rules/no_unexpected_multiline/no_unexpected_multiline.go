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

func buildFunctionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "function",
		Description: "Unexpected newline between function and ( of function call.",
	}
}

func buildPropertyMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "property",
		Description: "Unexpected newline between object and [ of property access.",
	}
}

func buildTaggedTemplateMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "taggedTemplate",
		Description: "Unexpected newline between template tag and template literal.",
	}
}

// getLine gets the line number for a node position
func getLine(sourceFile *ast.SourceFile, pos int) int {
	line, _ := scanner.GetLineAndCharacterOfPosition(sourceFile, pos)
	return line
}

// hasLineBreakBetween checks if there's a line break between two positions
func hasLineBreakBetween(sourceFile *ast.SourceFile, start int, end int) bool {
	return getLine(sourceFile, start) != getLine(sourceFile, end)
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	sourceFile := ctx.SourceFile

	return rule.RuleListeners{
		// Check for unexpected multiline in call expressions: foo\n()
		ast.KindCallExpression: func(node *ast.Node) {
			if node == nil {
				return
			}

			expr := node.Expression()
			if expr == nil {
				return
			}

			// Check for newline between function and opening paren
			// The opening paren would be right after the expression ends
			exprEnd := expr.End()
			nodeStart := node.Pos()

			// For call expressions, check if there's a newline before the arguments
			args := node.Arguments()
			if args != nil && len(args) > 0 {
				// Check between expression end and first argument
				if hasLineBreakBetween(sourceFile, exprEnd, nodeStart) {
					ctx.ReportNode(node, buildFunctionMessage())
				}
			} else {
				// Empty call - check for newline between expression and parentheses
				if hasLineBreakBetween(sourceFile, exprEnd, node.End()) {
					// Additional check: make sure it's actually multiline
					if getLine(sourceFile, exprEnd) != getLine(sourceFile, node.End()) {
						ctx.ReportNode(node, buildFunctionMessage())
					}
				}
			}
		},

		// Check for unexpected multiline in element access: obj\n[prop]
		ast.KindElementAccessExpression: func(node *ast.Node) {
			if node == nil {
				return
			}

			expr := node.Expression()
			argExpr := node.ArgumentExpression()

			if expr == nil || argExpr == nil {
				return
			}

			// Check for newline between object and opening bracket
			exprEnd := expr.End()
			argStart := argExpr.Pos()

			if hasLineBreakBetween(sourceFile, exprEnd, argStart) {
				ctx.ReportNode(node, buildPropertyMessage())
			}
		},

		// Check for unexpected multiline in tagged templates: tag\n`template`
		ast.KindTaggedTemplateExpression: func(node *ast.Node) {
			if node == nil {
				return
			}

			tag := node.Tag()
			template := node.Template()

			if tag == nil || template == nil {
				return
			}

			// Check for newline between tag and template
			tagEnd := tag.End()
			templateStart := template.Pos()

			if hasLineBreakBetween(sourceFile, tagEnd, templateStart) {
				ctx.ReportNode(node, buildTaggedTemplateMessage())
			}
		},
	}
}
