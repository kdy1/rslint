package no_extra_non_null_assertion

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func buildNoExtraNonNullAssertionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noExtraNonNullAssertion",
		Description: "Forbidden extra non-null assertion.",
	}
}

var NoExtraNonNullAssertionRule = rule.CreateRule(rule.Rule{
	Name: "no-extra-non-null-assertion",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindNonNullExpression: func(node *ast.Node) {
				expression := node.Expression()

				// Check for double non-null assertion: foo!!
				if ast.IsNonNullExpression(expression) {
					// Report the outer non-null assertion
					ctx.ReportNodeWithFixes(
						node,
						buildNoExtraNonNullAssertionMessage(),
						// Fix: replace the outer expression with the inner expression
						rule.RuleFixReplace(ctx.SourceFile, node, ctx.SourceFile.Text()[expression.Pos():expression.End()]),
					)
					return
				}

				// Check for non-null assertion before optional chaining: foo!?.bar or foo!?.()
				// The parent of the NonNullExpression should be a PropertyAccessExpression or CallExpression
				// with a QuestionDotToken
				if expression != nil {
					parent := node.Parent
					if parent != nil {
						// Check if parent has QuestionDotToken
						hasQuestionDot := false

						// For property access: obj!?.prop
						if ast.IsPropertyAccessExpression(parent) {
							propAccess := parent.AsPropertyAccessExpression()
							if propAccess != nil && propAccess.QuestionDotToken != nil {
								hasQuestionDot = true
							}
						}

						// For call expression: obj!?.()
						if ast.IsCallExpression(parent) {
							callExpr := parent.AsCallExpression()
							if callExpr != nil && callExpr.QuestionDotToken != nil {
								hasQuestionDot = true
							}
						}

						// For element access: obj!?.[prop]
						if ast.IsElementAccessExpression(parent) {
							elemAccess := parent.AsElementAccessExpression()
							if elemAccess != nil && elemAccess.QuestionDotToken != nil {
								hasQuestionDot = true
							}
						}

						if hasQuestionDot {
							// Report the non-null assertion as unnecessary
							ctx.ReportNodeWithFixes(
								node,
								buildNoExtraNonNullAssertionMessage(),
								// Fix: remove the non-null assertion, keeping just the expression
								rule.RuleFixReplace(ctx.SourceFile, node, ctx.SourceFile.Text()[expression.Pos():expression.End()]),
							)
						}
					}
				}
			},
		}
	},
})
