package no_unsafe_negation

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// NoUnsafeNegationOptions defines the configuration options for this rule
type NoUnsafeNegationOptions struct {
	EnforceForOrderingRelations bool `json:"enforceForOrderingRelations"`
}

// parseOptions parses and validates the rule options
func parseOptions(options any) NoUnsafeNegationOptions {
	opts := NoUnsafeNegationOptions{
		EnforceForOrderingRelations: false,
	}

	if options == nil {
		return opts
	}

	// Handle both array format [{ option: value }] and object format { option: value }
	var optsMap map[string]interface{}
	if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
		optsMap, _ = optArray[0].(map[string]interface{})
	} else {
		optsMap, _ = options.(map[string]interface{})
	}

	if optsMap != nil {
		if v, ok := optsMap["enforceForOrderingRelations"].(bool); ok {
			opts.EnforceForOrderingRelations = v
		}
	}

	return opts
}

// isInOrInstanceOfOperator checks if operator is "in" or "instanceof"
func isInOrInstanceOfOperator(op ast.SyntaxKind) bool {
	return op == ast.SyntaxKindInKeyword || op == ast.SyntaxKindInstanceOfKeyword
}

// isOrderingRelationalOperator checks if operator is <, >, <=, or >=
func isOrderingRelationalOperator(op ast.SyntaxKind) bool {
	return op == ast.SyntaxKindLessThanToken ||
		op == ast.SyntaxKindGreaterThanToken ||
		op == ast.SyntaxKindLessThanEqualsToken ||
		op == ast.SyntaxKindGreaterThanEqualsToken
}

// isNegation checks if node is a logical negation (!)
func isNegation(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindPrefixUnaryExpression {
		return false
	}
	prefix := node.AsPrefixUnaryExpression()
	return prefix != nil && prefix.Operator == ast.SyntaxKindExclamationToken
}

// getOperatorText returns the text representation of a binary operator
func getOperatorText(op ast.SyntaxKind) string {
	switch op {
	case ast.SyntaxKindInKeyword:
		return "in"
	case ast.SyntaxKindInstanceOfKeyword:
		return "instanceof"
	case ast.SyntaxKindLessThanToken:
		return "<"
	case ast.SyntaxKindGreaterThanToken:
		return ">"
	case ast.SyntaxKindLessThanEqualsToken:
		return "<="
	case ast.SyntaxKindGreaterThanEqualsToken:
		return ">="
	default:
		return ""
	}
}

// NoUnsafeNegationRule implements the no-unsafe-negation rule
// Disallow negating the left operand of relational operators
var NoUnsafeNegationRule = rule.Rule{
	Name: "no-unsafe-negation",
	Run:  run,
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := parseOptions(options)

	return rule.RuleListeners{
		ast.KindBinaryExpression: func(node *ast.Node) {
			binary := node.AsBinaryExpression()
			if binary == nil || binary.Left == nil {
				return
			}

			operator := binary.OperatorToken.Kind
			orderingRelationRuleApplies := opts.EnforceForOrderingRelations && isOrderingRelationalOperator(operator)

			// Check if operator is in/instanceof OR (ordering relation AND option enabled)
			if !isInOrInstanceOfOperator(operator) && !orderingRelationRuleApplies {
				return
			}

			// Check if left operand is a negation
			if !isNegation(binary.Left) {
				return
			}

			// Check if the negation is already parenthesized
			if utils.IsParenthesized(ctx.SourceFile, binary.Left) {
				return
			}

			// Report the violation with suggestions
			operatorText := getOperatorText(operator)
			msg := rule.RuleMessage{
				Id:          "unexpected",
				Description: fmt.Sprintf("Unexpected negating the left operand of '%s' operator.", operatorText),
			}

			// Create suggestion fixes
			text := ctx.SourceFile.Text()
			leftRange := utils.TrimNodeTextRange(ctx.SourceFile, binary.Left)
			nodeRange := utils.TrimNodeTextRange(ctx.SourceFile, node)

			// Get the operand being negated (the part after !)
			prefix := binary.Left.AsPrefixUnaryExpression()
			if prefix == nil || prefix.Operand == nil {
				ctx.ReportNode(node, msg)
				return
			}

			operandRange := utils.TrimNodeTextRange(ctx.SourceFile, prefix.Operand)

			// Suggestion 1: Negate entire expression - !(operand operator right)
			// Find the position after the ! token
			negationTokenEnd := leftRange.Pos() + 1 // Skip the '!' character
			restOfExpression := text[negationTokenEnd:nodeRange.End()]
			fix1 := rule.RuleFixReplace(ctx.SourceFile, node, fmt.Sprintf("!(%s)", restOfExpression))

			// Suggestion 2: Parenthesize the negation - (!operand) operator right
			leftText := text[leftRange.Pos():leftRange.End()]
			restText := text[leftRange.End():nodeRange.End()]
			fix2 := rule.RuleFixReplace(ctx.SourceFile, node, fmt.Sprintf("(%s)%s", leftText, restText))

			ctx.ReportNodeWithSuggestions(node, msg, []rule.RuleSuggestion{
				{
					Description: fmt.Sprintf("Negate '%s' expression instead of its left operand. This changes the current behavior.", operatorText),
					Fix:         fix1,
				},
				{
					Description: "Wrap negation in '()' to make the intention explicit. This preserves the current behavior.",
					Fix:         fix2,
				},
			})
		},
	}
}
