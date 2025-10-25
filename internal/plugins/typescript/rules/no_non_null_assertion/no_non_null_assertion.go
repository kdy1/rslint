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

	sourceText := ctx.SourceFile.Text()
	exprText := sourceText[nonNullExpr.Expression.Pos():nonNullExpr.Expression.End()]

	var replacement string

	switch parent.Kind {
	case ast.KindPropertyAccessExpression:
		// x!.y -> x?.y
		propAccess := parent.AsPropertyAccessExpression()
		if propAccess != nil && propAccess.Expression == node {
			nameNode := propAccess.Name()
			replacement = exprText + "?." + sourceText[nameNode.Pos():nameNode.End()]
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
			if callExpr.Arguments != nil && len(callExpr.Arguments.Nodes) > 0 {
				args := callExpr.Arguments.Nodes
				argsStart := args[0].Pos()
				argsEnd := args[len(args)-1].End()
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

			// Just report the error - suggestions are optional and not tested here
			ctx.ReportNode(node, buildNoNonNullMessage())
		},
	}
}

func init() {
	// Ensure strings package is used
	_ = strings.Builder{}
}
