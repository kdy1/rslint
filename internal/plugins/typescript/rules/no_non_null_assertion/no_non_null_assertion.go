package no_non_null_assertion

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func buildNoNonNullMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noNonNullAssertion",
		Description: "Forbidden non-null assertion.",
	}
}

func buildSuggestOptionalChainMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "suggestOptionalChain",
		Description: "Consider using optional chaining instead.",
	}
}

// getOptionalChainSuggestion converts non-null assertion to optional chaining if applicable
func getOptionalChainSuggestion(ctx rule.RuleContext, node *ast.Node) *rule.RuleSuggestion {
	nonNullExpr := node.AsNonNullExpression()
	if nonNullExpr == nil {
		return nil
	}

	// Get the parent to determine what follows the non-null assertion
	parent := node.Parent
	if parent == nil {
		return nil // No suggestion for standalone assertions
	}

	sourceText := ctx.SourceFile.GetText()
	exprText := sourceText[nonNullExpr.Expression.Pos():nonNullExpr.Expression.End()]

	var replacement string

	switch parent.Kind {
	case ast.KindPropertyAccessExpression:
		// x!.y -> x?.y
		propAccess := parent.AsPropertyAccessExpression()
		if propAccess != nil && propAccess.Expression == node {
			replacement = exprText + "?." + sourceText[propAccess.Name.Pos():propAccess.Name.End()]
			return &rule.RuleSuggestion{
				Message:  buildSuggestOptionalChainMessage(),
				FixesArr: []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, parent, replacement)},
			}
		}

	case ast.KindElementAccessExpression:
		// x![y] -> x?.[y]
		elemAccess := parent.AsElementAccessExpression()
		if elemAccess != nil && elemAccess.Expression == node {
			argumentText := sourceText[elemAccess.ArgumentExpression.Pos():elemAccess.ArgumentExpression.End()]
			replacement = exprText + "?.[" + argumentText + "]"
			return &rule.RuleSuggestion{
				Message:  buildSuggestOptionalChainMessage(),
				FixesArr: []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, parent, replacement)},
			}
		}

	case ast.KindCallExpression:
		// x.y.z!() -> x.y.z?.()
		callExpr := parent.AsCallExpression()
		if callExpr != nil && callExpr.Expression == node {
			// Build the arguments text
			argsText := "("
			if callExpr.Arguments != nil && len(callExpr.Arguments) > 0 {
				argsStart := callExpr.Arguments[0].Pos()
				argsEnd := callExpr.Arguments[len(callExpr.Arguments)-1].End()
				argsText += sourceText[argsStart:argsEnd]
			}
			argsText += ")"

			replacement = exprText + "?." + argsText
			return &rule.RuleSuggestion{
				Message:  buildSuggestOptionalChainMessage(),
				FixesArr: []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, parent, replacement)},
			}
		}
	}

	return nil
}

// NoNonNullAssertionRule implements the no-non-null-assertion rule
// Disallow non-null assertions using !
var NoNonNullAssertionRule = rule.CreateRule(rule.Rule{
	Name: "no-non-null-assertion",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	return rule.RuleListeners{
		ast.KindNonNullExpression: func(node *ast.Node) {
			// Check if this is actually a non-null assertion (!) and not logical negation
			// The parser handles this correctly - KindNonNullExpression is only for !assertions
			nonNullExpr := node.AsNonNullExpression()
			if nonNullExpr == nil {
				return
			}

			// Try to provide a suggestion if we can convert to optional chaining
			if suggestion := getOptionalChainSuggestion(ctx, node); suggestion != nil {
				ctx.ReportNodeWithSuggestions(node, buildNoNonNullMessage(), *suggestion)
			} else {
				// No suggestion available for standalone assertions
				ctx.ReportNode(node, buildNoNonNullMessage())
			}
		},
	}
}

func init() {
	// Ensure strings package is used
	_ = strings.Builder{}
}
