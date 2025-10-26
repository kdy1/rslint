package no_extra_non_null_assertion

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// NoExtraNonNullAssertionRule implements the no-extra-non-null-assertion rule
// Disallow extra non-null assertions
var NoExtraNonNullAssertionRule = rule.CreateRule(rule.Rule{
	Name: "no-extra-non-null-assertion",
	Run:  run,
})

func buildNoExtraNonNullAssertionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noExtraNonNullAssertion",
		Description: "Forbidden extra non-null assertion.",
	}
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	return rule.RuleListeners{
		// Detect double non-null assertions: foo!!.bar
		ast.KindNonNullExpression: func(node *ast.Node) {
			nonNullExpr := node.AsNonNullExpression()
			expression := nonNullExpr.Expression

			// Check if the expression is also a non-null expression (double assertion)
			if expression.Kind == ast.KindNonNullExpression {
				message := buildNoExtraNonNullAssertionMessage()

				// Create a fix that removes the extra ! by replacing the outer expression
				// with just the inner non-null expression
				sourceRange := utils.TrimNodeTextRange(ctx.SourceFile, expression)
				sourceText := ctx.SourceFile.Text()[sourceRange.Pos():sourceRange.End()]
				fix := rule.RuleFixReplace(ctx.SourceFile, node, sourceText)

				ctx.ReportNodeWithFixes(node, message, fix)
			}
		},

		// Detect non-null assertion before optional call: foo!?.()
		ast.KindCallExpression: func(node *ast.Node) {
			callExpr := node.AsCallExpression()

			// Check if this is an optional call expression
			if callExpr.QuestionDotToken == nil {
				return
			}

			// Check if the callee is a non-null expression
			expression := callExpr.Expression
			if expression.Kind == ast.KindNonNullExpression {
				message := buildNoExtraNonNullAssertionMessage()

				// Create a fix that removes the non-null assertion
				// We need to get the expression inside the non-null assertion
				nonNullExpr := expression.AsNonNullExpression()
				innerExpr := nonNullExpr.Expression

				// Build the fixed code by replacing the non-null expression with its inner expression
				innerRange := utils.TrimNodeTextRange(ctx.SourceFile, innerExpr)
				innerText := ctx.SourceFile.Text()[innerRange.Pos():innerRange.End()]

				// We need to replace just the expression part, keeping the rest of the call
				fix := rule.RuleFixReplace(ctx.SourceFile, expression, innerText)

				ctx.ReportNodeWithFixes(expression, message, fix)
			}
		},

		// Detect non-null assertion before optional member access: foo!?.bar
		ast.KindPropertyAccessExpression: func(node *ast.Node) {
			propAccess := node.AsPropertyAccessExpression()

			// Check if this is an optional property access
			if propAccess.QuestionDotToken == nil {
				return
			}

			// Check if the expression is a non-null expression
			expression := propAccess.Expression
			if expression.Kind == ast.KindNonNullExpression {
				message := buildNoExtraNonNullAssertionMessage()

				// Create a fix that removes the non-null assertion
				nonNullExpr := expression.AsNonNullExpression()
				innerExpr := nonNullExpr.Expression
				innerRange := utils.TrimNodeTextRange(ctx.SourceFile, innerExpr)
				innerText := ctx.SourceFile.Text()[innerRange.Pos():innerRange.End()]

				fix := rule.RuleFixReplace(ctx.SourceFile, expression, innerText)

				ctx.ReportNodeWithFixes(expression, message, fix)
			}
		},

		// Detect non-null assertion before optional element access: foo!?.[bar]
		ast.KindElementAccessExpression: func(node *ast.Node) {
			elemAccess := node.AsElementAccessExpression()

			// Check if this is an optional element access
			if elemAccess.QuestionDotToken == nil {
				return
			}

			// Check if the expression is a non-null expression
			expression := elemAccess.Expression
			if expression.Kind == ast.KindNonNullExpression {
				message := buildNoExtraNonNullAssertionMessage()

				// Create a fix that removes the non-null assertion
				nonNullExpr := expression.AsNonNullExpression()
				innerExpr := nonNullExpr.Expression
				innerRange := utils.TrimNodeTextRange(ctx.SourceFile, innerExpr)
				innerText := ctx.SourceFile.Text()[innerRange.Pos():innerRange.End()]

				fix := rule.RuleFixReplace(ctx.SourceFile, expression, innerText)

				ctx.ReportNodeWithFixes(expression, message, fix)
			}
		},
	}
}
