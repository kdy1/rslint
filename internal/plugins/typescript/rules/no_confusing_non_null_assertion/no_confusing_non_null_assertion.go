package no_confusing_non_null_assertion

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildConfusingEqualMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "confusingEqual",
		Description: "Confusing combinations of non-null assertion and equality test like `a! == b` may be mistaken for `a != b`.",
	}
}

func buildConfusingAssignMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "confusingAssign",
		Description: "Confusing combinations of non-null assertion and assignment like `a! = b` may be mistaken for `a != b`.",
	}
}

func buildConfusingOperatorMessage(operator string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "confusingOperator",
		Description: "Confusing combinations of non-null assertion and " + operator + " test like `a! " + operator + " b` may be mistaken for negation.",
	}
}

func buildNotNeedInEqualTestMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "notNeedInEqualTest",
		Description: "Remove the non-null assertion operator.",
	}
}

func buildNotNeedInAssignMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "notNeedInAssign",
		Description: "Remove the non-null assertion operator.",
	}
}

func buildNotNeedInOperatorMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "notNeedInOperator",
		Description: "Remove the non-null assertion operator.",
	}
}

func buildWrapUpLeftMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "wrapUpLeft",
		Description: "Wrap the left side with parentheses to clarify the non-null assertion.",
	}
}

var NoConfusingNonNullAssertionRule = rule.CreateRule(rule.Rule{
	Name: "no-confusing-non-null-assertion",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				if node.Kind != ast.KindBinaryExpression {
					return
				}

				binaryExpr := node.AsBinaryExpression()
				left := binaryExpr.Left
				operator := binaryExpr.OperatorToken.Kind

				// Skip parenthesized expressions to get to the actual left expression
				leftExpr := ast.SkipParentheses(left)

				// Check if the left side ends with a non-null assertion
				if leftExpr.Kind != ast.KindNonNullExpression {
					return
				}

				nonNullExpr := leftExpr.AsNonNullExpression()

				// Get the text range for the non-null assertion
				exprRange := utils.TrimNodeTextRange(ctx.SourceFile, nonNullExpr.Expression)
				exclamationRange := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, exprRange.End())

				// Handle different operator types
				switch operator {
				case ast.KindEqualsEqualsToken, ast.KindEqualsEqualsEqualsToken:
					// Handle == and ===
					// Determine if this is a simple identifier or complex expression
					needsParens := !ast.IsIdentifier(nonNullExpr.Expression) &&
						!ast.IsPropertyAccessExpression(nonNullExpr.Expression) &&
						!ast.IsElementAccessExpression(nonNullExpr.Expression) &&
						nonNullExpr.Expression.Kind != ast.KindParenthesizedExpression

					if needsParens {
						// For complex expressions like "a + b! == c", suggest wrapping
						ctx.ReportNodeWithSuggestions(node, buildConfusingEqualMessage(), rule.RuleSuggestion{
							Message: buildWrapUpLeftMessage(),
							FixesArr: []rule.RuleFix{
								rule.RuleFixInsertBefore(ctx.SourceFile, nonNullExpr.Expression, "("),
								rule.RuleFixInsertAfter(left, ")"),
							},
						})
					} else {
						// For simple expressions, suggest removing the !
						ctx.ReportNodeWithSuggestions(node, buildConfusingEqualMessage(), rule.RuleSuggestion{
							Message: buildNotNeedInEqualTestMessage(),
							FixesArr: []rule.RuleFix{
								rule.RuleFixRemoveRange(exclamationRange),
							},
						})
					}

				case ast.KindEqualsToken:
					// Handle =
					ctx.ReportNodeWithSuggestions(node, buildConfusingAssignMessage(), rule.RuleSuggestion{
						Message: buildNotNeedInAssignMessage(),
						FixesArr: []rule.RuleFix{
							rule.RuleFixRemoveRange(exclamationRange),
						},
					})

				case ast.KindInKeyword:
					// Handle 'in'
					ctx.ReportNodeWithSuggestions(
						node,
						buildConfusingOperatorMessage("in"),
						rule.RuleSuggestion{
							Message: buildNotNeedInOperatorMessage(),
							FixesArr: []rule.RuleFix{
								rule.RuleFixRemoveRange(exclamationRange),
							},
						},
						rule.RuleSuggestion{
							Message: buildWrapUpLeftMessage(),
							FixesArr: []rule.RuleFix{
								rule.RuleFixInsertBefore(ctx.SourceFile, left, "("),
								rule.RuleFixInsertAfter(left, ")"),
							},
						},
					)

				case ast.KindInstanceOfKeyword:
					// Handle 'instanceof'
					ctx.ReportNodeWithSuggestions(
						node,
						buildConfusingOperatorMessage("instanceof"),
						rule.RuleSuggestion{
							Message: buildNotNeedInOperatorMessage(),
							FixesArr: []rule.RuleFix{
								rule.RuleFixRemoveRange(exclamationRange),
							},
						},
						rule.RuleSuggestion{
							Message: buildWrapUpLeftMessage(),
							FixesArr: []rule.RuleFix{
								rule.RuleFixInsertBefore(ctx.SourceFile, left, "("),
								rule.RuleFixInsertAfter(left, ")"),
							},
						},
					)
				}
			},
		}
	},
})
