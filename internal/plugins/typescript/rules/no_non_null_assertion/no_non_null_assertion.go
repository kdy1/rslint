package no_non_null_assertion

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func buildNoNonNullMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noNonNull",
		Description: "Forbidden non-null assertion.",
	}
}

func buildSuggestOptionalChainMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "suggestOptionalChain",
		Description: "Consider using optional chaining instead.",
	}
}

var NoNonNullAssertionRule = rule.CreateRule(rule.Rule{
	Name: "no-non-null-assertion",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindNonNullExpression: func(node *ast.Node) {
				// Get the expression being asserted
				expression := node.Expression()

				// Build a fix to remove the exclamation mark
				buildRemoveExclamationFix := func() rule.RuleFix {
					s := scanner.GetScannerForSourceFile(ctx.SourceFile, expression.End())
					return rule.RuleFixRemoveRange(s.TokenRange())
				}

				// Try to build a suggestion for optional chaining
				suggestion := tryBuildOptionalChainingSuggestion(ctx, node, expression)

				if suggestion != nil {
					ctx.ReportNodeWithSuggestions(node, buildNoNonNullMessage(), *suggestion)
				} else {
					ctx.ReportNodeWithFixes(node, buildNoNonNullMessage(), buildRemoveExclamationFix())
				}
			},
		}
	},
})

// tryBuildOptionalChainingSuggestion attempts to build an optional chaining suggestion
func tryBuildOptionalChainingSuggestion(ctx rule.RuleContext, node *ast.Node, expression *ast.Node) *rule.RuleSuggestion {
	// Check if this is followed by a property access, element access, or call expression
	parent := node.Parent

	if parent == nil {
		return nil
	}

	var fixes []rule.RuleFix

	// Get the exclamation mark range to remove it
	s := scanner.GetScannerForSourceFile(ctx.SourceFile, expression.End())
	exclamationRange := s.TokenRange()

	// Determine what kind of access follows the non-null assertion
	switch parent.Kind {
	case ast.KindPropertyAccessExpression:
		// x!.y -> x?.y
		propAccess := parent.AsPropertyAccessExpression()
		if propAccess.Expression == node {
			// Get the dot token after the expression
			dotRange := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, expression.End()+1) // +1 to skip the !

			fixes = []rule.RuleFix{
				rule.RuleFixRemoveRange(exclamationRange),
				rule.RuleFixReplaceRange(dotRange, "?."),
			}
		}

	case ast.KindElementAccessExpression:
		// x![y] -> x?.[y]
		elemAccess := parent.AsElementAccessExpression()
		if elemAccess.Expression == node {
			// Get the [ token after the expression
			bracketRange := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, expression.End()+1) // +1 to skip the !

			fixes = []rule.RuleFix{
				rule.RuleFixRemoveRange(exclamationRange),
				rule.RuleFixReplaceRange(bracketRange, "?.["),
			}
		}

	case ast.KindCallExpression:
		// x!() -> x?.()
		callExpr := parent.AsCallExpression()
		if callExpr.Expression == node {
			// Get the ( token after the expression
			parenRange := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, expression.End()+1) // +1 to skip the !

			fixes = []rule.RuleFix{
				rule.RuleFixRemoveRange(exclamationRange),
				rule.RuleFixReplaceRange(parenRange, "?.("),
			}
		}

	case ast.KindNonNullExpression:
		// x!! -> don't suggest, just remove one
		return nil

	default:
		// For other cases, check if we can convert the inner expression
		return tryConvertInnerExpression(ctx, node, expression, exclamationRange)
	}

	if len(fixes) > 0 {
		return &rule.RuleSuggestion{
			Message:  buildSuggestOptionalChainMessage(),
			FixesArr: fixes,
		}
	}

	return nil
}

// tryConvertInnerExpression tries to convert inner expressions to optional chaining
func tryConvertInnerExpression(ctx rule.RuleContext, node *ast.Node, expression *ast.Node, exclamationRange scanner.TextRange) *rule.RuleSuggestion {
	// Check if the expression itself is a property access, element access, or call
	// that we can convert to optional chaining
	// e.g., x.y! -> x.y (but we can suggest x?.y if x might be null)

	// Get the exclamation mark to remove
	fixes := []rule.RuleFix{
		rule.RuleFixRemoveRange(exclamationRange),
	}

	// Check the expression kind to see if we can suggest optional chaining
	switch expression.Kind {
	case ast.KindPropertyAccessExpression:
		// x.y! -> suggest x?.y
		propAccess := expression.AsPropertyAccessExpression()
		innerExpr := propAccess.Expression

		// Find the dot before the property
		dotPos := innerExpr.End()
		dotRange := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, dotPos)

		fixes = append(fixes, rule.RuleFixReplaceRange(dotRange, "?."))

		return &rule.RuleSuggestion{
			Message:  buildSuggestOptionalChainMessage(),
			FixesArr: fixes,
		}

	case ast.KindElementAccessExpression:
		// x[y]! -> suggest x?.[y]
		elemAccess := expression.AsElementAccessExpression()
		innerExpr := elemAccess.Expression

		// Find the [ before the element
		bracketPos := innerExpr.End()
		bracketRange := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, bracketPos)

		fixes = append(fixes, rule.RuleFixReplaceRange(bracketRange, "?.["))

		return &rule.RuleSuggestion{
			Message:  buildSuggestOptionalChainMessage(),
			FixesArr: fixes,
		}

	case ast.KindCallExpression:
		// x.y()! -> suggest x.y?.()
		callExpr := expression.AsCallExpression()

		// Check if the call expression's target is a property or element access
		if callExpr.Expression.Kind == ast.KindPropertyAccessExpression {
			propAccess := callExpr.Expression.AsPropertyAccessExpression()
			innerExpr := propAccess.Expression

			dotPos := innerExpr.End()
			dotRange := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, dotPos)

			fixes = append(fixes, rule.RuleFixReplaceRange(dotRange, "?."))

			return &rule.RuleSuggestion{
				Message:  buildSuggestOptionalChainMessage(),
				FixesArr: fixes,
			}
		} else if callExpr.Expression.Kind == ast.KindElementAccessExpression {
			elemAccess := callExpr.Expression.AsElementAccessExpression()
			innerExpr := elemAccess.Expression

			bracketPos := innerExpr.End()
			bracketRange := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, bracketPos)

			fixes = append(fixes, rule.RuleFixReplaceRange(bracketRange, "?.["))

			return &rule.RuleSuggestion{
				Message:  buildSuggestOptionalChainMessage(),
				FixesArr: fixes,
			}
		}
	}

	return nil
}

// Helper function to get the source text for a node
func getNodeText(ctx rule.RuleContext, node *ast.Node) string {
	text := ctx.SourceFile.Text()
	start := node.Pos()
	end := node.End()
	if start >= 0 && end <= len(text) && start < end {
		return text[start:end]
	}
	return ""
}

// Helper function to check if a character is an operator character
func isOperatorChar(c byte) bool {
	return strings.ContainsRune("!<>=&|+-*/%^~", rune(c))
}
