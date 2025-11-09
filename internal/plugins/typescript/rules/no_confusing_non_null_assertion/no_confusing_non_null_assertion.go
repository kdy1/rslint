package no_confusing_non_null_assertion

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var NoConfusingNonNullAssertionRule = rule.CreateRule(rule.Rule{
	Name: "no-confusing-non-null-assertion",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				binaryExpr := node.AsBinaryExpression()
				if binaryExpr == nil {
					return
				}

				operator := binaryExpr.OperatorToken.Kind

				// Check for confusing operators: ==, ===, =, in, instanceof
				isConfusingOperator := false
				var messageId string
				var operatorText string

				switch operator {
				case ast.KindEqualsEqualsToken:
					isConfusingOperator = true
					messageId = "confusingEqual"
					operatorText = "=="
				case ast.KindEqualsEqualsEqualsToken:
					isConfusingOperator = true
					messageId = "confusingEqual"
					operatorText = "==="
				case ast.KindEqualsToken:
					isConfusingOperator = true
					messageId = "confusingAssign"
					operatorText = "="
				case ast.KindInKeyword:
					isConfusingOperator = true
					messageId = "confusingOperator"
					operatorText = "in"
				case ast.KindInstanceOfKeyword:
					isConfusingOperator = true
					messageId = "confusingOperator"
					operatorText = "instanceof"
				}

				if !isConfusingOperator {
					return
				}

				// Check if we need to report an error by examining the left side
				// Two cases:
				// 1. Left side is directly a non-null expression (e.g., a! == b)
				// 2. Left side when parentheses are skipped reveals a complex expression where
				//    only part of it has a non-null assertion (e.g., a + b! == c)

				leftNode := binaryExpr.Left
				leftWithoutParens := ast.SkipParentheses(leftNode)

				// Case 1: Direct non-null expression on the left
				if leftWithoutParens.Kind == ast.KindNonNullExpression {
					nonNullExpr := leftWithoutParens.AsNonNullExpression()
					if nonNullExpr == nil {
						return
					}

					// Get the expression inside the non-null assertion
					innerExpr := nonNullExpr.Expression
					innerRange := utils.TrimNodeTextRange(ctx.SourceFile, innerExpr)
					innerText := ctx.SourceFile.Text()[innerRange.Pos():innerRange.End()]

					// Build the appropriate message
					var message rule.RuleMessage
					if messageId == "confusingEqual" {
						message = rule.RuleMessage{
							Id:          messageId,
							Description: "Confusing non-null assertion next to equal sign. Use explicit parentheses to clarify.",
						}
					} else if messageId == "confusingAssign" {
						message = rule.RuleMessage{
							Id:          messageId,
							Description: "Confusing non-null assertion next to assign sign. Use explicit parentheses to clarify.",
						}
					} else {
						message = rule.RuleMessage{
							Id:          messageId,
							Description: fmt.Sprintf("Confusing non-null assertion next to '%s' operator. Use explicit parentheses to clarify.", operatorText),
						}
					}

					// For equality and assignment, suggest removing the !
					if messageId == "confusingEqual" || messageId == "confusingAssign" {
						suggestionId := "notNeedInEqualTest"
						if messageId == "confusingAssign" {
							suggestionId = "notNeedInAssign"
						}

						ctx.ReportNodeWithSuggestions(leftWithoutParens, message,
							rule.RuleSuggestion{
								MessageId:   suggestionId,
								Description: "Remove the non-null assertion operator",
								Fix:         rule.RuleFixReplace(ctx.SourceFile, leftWithoutParens, innerText),
							})
					} else {
						// For 'in' and 'instanceof', provide two suggestions:
						// 1. Remove the !
						// 2. Wrap the expression in parentheses
						fullRange := utils.TrimNodeTextRange(ctx.SourceFile, leftWithoutParens)
						fullText := ctx.SourceFile.Text()[fullRange.Pos():fullRange.End()]

						ctx.ReportNodeWithSuggestions(leftWithoutParens, message,
							rule.RuleSuggestion{
								MessageId:   "notNeedInOperator",
								Description: "Remove the non-null assertion operator",
								Fix:         rule.RuleFixReplace(ctx.SourceFile, leftWithoutParens, innerText),
							},
							rule.RuleSuggestion{
								MessageId:   "wrapUpLeft",
								Description: "Wrap the left-hand side in parentheses",
								Fix:         rule.RuleFixReplace(ctx.SourceFile, leftWithoutParens, "("+fullText+")"),
							})
					}
					return
				}

				// Case 2: Check if the left side contains a non-null assertion somewhere
				// (e.g., a + b! == c)
				// We need to check if there's a non-null assertion in the left expression tree
				hasNonNullAssertion := false
				var checkForNonNull func(*ast.Node) bool
				checkForNonNull = func(n *ast.Node) bool {
					if n == nil {
						return false
					}
					if n.Kind == ast.KindNonNullExpression {
						return true
					}
					// Check children - for binary expressions, check both sides
					if binExpr := n.AsBinaryExpression(); binExpr != nil {
						return checkForNonNull(binExpr.Left) || checkForNonNull(binExpr.Right)
					}
					// For other expressions, we'd need to check their children
					// For now, we'll keep it simple and only check binary expressions
					return false
				}

				hasNonNullAssertion = checkForNonNull(leftWithoutParens)
				if !hasNonNullAssertion {
					return
				}

				// Report the error with a suggestion to wrap the left side
				var message rule.RuleMessage
				if messageId == "confusingEqual" {
					message = rule.RuleMessage{
						Id:          messageId,
						Description: "Confusing non-null assertion next to equal sign. Use explicit parentheses to clarify.",
					}
				} else if messageId == "confusingAssign" {
					message = rule.RuleMessage{
						Id:          messageId,
						Description: "Confusing non-null assertion next to assign sign. Use explicit parentheses to clarify.",
					}
				} else {
					message = rule.RuleMessage{
						Id:          messageId,
						Description: fmt.Sprintf("Confusing non-null assertion next to '%s' operator. Use explicit parentheses to clarify.", operatorText),
					}
				}

				// Get the left side text and wrap it in parentheses
				leftRange := utils.TrimNodeTextRange(ctx.SourceFile, leftNode)
				leftText := ctx.SourceFile.Text()[leftRange.Pos():leftRange.End()]

				ctx.ReportNodeWithSuggestions(node, message,
					rule.RuleSuggestion{
						MessageId:   "wrapUpLeft",
						Description: "Wrap the left-hand side in parentheses",
						Fix:         rule.RuleFixReplace(ctx.SourceFile, leftNode, "("+leftText+")"),
					})
			},
		}
	},
})
