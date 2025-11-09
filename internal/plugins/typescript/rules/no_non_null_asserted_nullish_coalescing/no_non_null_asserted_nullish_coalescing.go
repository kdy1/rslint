package no_non_null_asserted_nullish_coalescing

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildNoNonNullAssertedNullishCoalescingMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noNonNullAssertedNullishCoalescing",
		Description: "Non-null assertion operators should not be used with the nullish coalescing operator.",
	}
}

func buildSuggestRemovingNonNullMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "suggestRemovingNonNull",
		Description: "Remove the non-null assertion.",
	}
}

var NoNonNullAssertedNullishCoalescingRule = rule.CreateRule(rule.Rule{
	Name: "no-non-null-asserted-nullish-coalescing",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				expr := node.AsBinaryExpression()

				// Check if this is a nullish coalescing operator (??)
				if expr.OperatorToken.Kind != ast.KindQuestionQuestionToken {
					return
				}

				// Check if the left operand is a non-null assertion expression
				if !ast.IsNonNullExpression(expr.Left) {
					return
				}

				nonNullExpr := expr.Left.AsNonNullExpression()

				// Get the range of the entire non-null expression (including the !)
				nonNullExprRange := utils.TrimNodeTextRange(ctx.SourceFile, expr.Left)

				// Get the range of just the expression without the !
				innerExprRange := utils.TrimNodeTextRange(ctx.SourceFile, nonNullExpr.Expression)

				// Find the position of the ! token (exclamation mark)
				// The ! is at the end of the non-null expression, right before the operator
				exclamationRange := scanner.TextRange{
					Pos: innerExprRange.End(),
					End: nonNullExprRange.End(),
				}

				// Report the error with a suggestion to remove the !
				ctx.ReportNodeWithSuggestions(expr.Left, buildNoNonNullAssertedNullishCoalescingMessage(), rule.RuleSuggestion{
					Message: buildSuggestRemovingNonNullMessage(),
					FixesArr: []rule.RuleFix{
						rule.RuleFixRemoveRange(exclamationRange),
					},
				})
			},
		}
	},
})
