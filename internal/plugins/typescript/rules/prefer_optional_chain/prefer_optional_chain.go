package prefer_optional_chain

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildPreferOptionalChainMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferOptionalChain",
		Description: "Prefer using an optional chain expression instead, as it's more concise and easier to read.",
	}
}

func buildOptionalChainSuggestMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "optionalChainSuggest",
		Description: "Change to an optional chain.",
	}
}

// Check if a node is an empty object literal {}
func isEmptyObjectLiteral(node *ast.Node) bool {
	if node == nil || !ast.IsObjectLiteralExpression(node) {
		return false
	}
	objLit := node.AsObjectLiteralExpression()
	return objLit.Properties == nil || len(objLit.Properties.Slice()) == 0
}

// Check if node needs parentheses when converting to optional chain
func needsParentheses(node *ast.Node) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindBinaryExpression, ast.KindConditionalExpression:
		return true
	case ast.KindAwaitExpression:
		return true
	default:
		return false
	}
}

// Get text for a node, wrapping in parentheses if needed
func getNodeTextMaybeWrapped(ctx rule.RuleContext, node *ast.Node) string {
	text := ctx.SourceFile.Text()[utils.TrimNodeTextRange(ctx.SourceFile, node).Pos():utils.TrimNodeTextRange(ctx.SourceFile, node).End()]
	if needsParentheses(node) {
		return "(" + text + ")"
	}
	return text
}

// Check if a binary expression is a comparison (==, ===, !=, !==)
func isComparison(node *ast.Node) bool {
	if node == nil || !ast.IsBinaryExpression(node) {
		return false
	}
	binExpr := node.AsBinaryExpression()
	op := binExpr.OperatorToken.Kind

	return op == ast.KindEqualsEqualsToken ||
		op == ast.KindEqualsEqualsEqualsToken ||
		op == ast.KindExclamationEqualsToken ||
		op == ast.KindExclamationEqualsEqualsToken
}

// Check if expression is a nullish check (=== null, === undefined, == null, == undefined, etc.)
func isNullishCheck(node *ast.Node) bool {
	if node == nil || !ast.IsBinaryExpression(node) {
		return false
	}

	binExpr := node.AsBinaryExpression()
	op := binExpr.OperatorToken.Kind

	// Check for ==, ===, !=, !== operators
	if op != ast.KindEqualsEqualsToken &&
		op != ast.KindEqualsEqualsEqualsToken &&
		op != ast.KindExclamationEqualsToken &&
		op != ast.KindExclamationEqualsEqualsToken {
		return false
	}

	// Check if one side is null or undefined
	left := binExpr.Left
	right := binExpr.Right

	isLeftNullish := (left.Kind == ast.KindNullKeyword) ||
		(ast.IsIdentifier(left) && left.AsIdentifier().EscapedText == "undefined")
	isRightNullish := (right.Kind == ast.KindNullKeyword) ||
		(ast.IsIdentifier(right) && right.AsIdentifier().EscapedText == "undefined")

	return isLeftNullish || isRightNullish
}

// Extract the expression being checked from a nullish check
func getCheckedExpression(node *ast.Node) *ast.Node {
	if !isNullishCheck(node) {
		return nil
	}

	binExpr := node.AsBinaryExpression()
	left := binExpr.Left
	right := binExpr.Right

	// Return the non-nullish side
	isLeftNullish := (left.Kind == ast.KindNullKeyword) ||
		(ast.IsIdentifier(left) && left.AsIdentifier().EscapedText == "undefined")

	if isLeftNullish {
		return right
	}
	return left
}

// Check if two expressions are equivalent (simple text comparison)
func areExpressionsEquivalent(ctx rule.RuleContext, expr1, expr2 *ast.Node) bool {
	if expr1 == nil || expr2 == nil {
		return false
	}

	text1 := ctx.SourceFile.Text()[utils.TrimNodeTextRange(ctx.SourceFile, expr1).Pos():utils.TrimNodeTextRange(ctx.SourceFile, expr1).End()]
	text2 := ctx.SourceFile.Text()[utils.TrimNodeTextRange(ctx.SourceFile, expr2).Pos():utils.TrimNodeTextRange(ctx.SourceFile, expr2).End()]

	return text1 == text2
}

// Check if a property access chain matches a base expression
func isPropertyAccessOf(ctx rule.RuleContext, propAccess, base *ast.Node) bool {
	if propAccess == nil || base == nil {
		return false
	}

	if ast.IsPropertyAccessExpression(propAccess) {
		expr := propAccess.AsPropertyAccessExpression().Expression
		return areExpressionsEquivalent(ctx, expr, base)
	}
	if ast.IsElementAccessExpression(propAccess) {
		expr := propAccess.AsElementAccessExpression().Expression
		return areExpressionsEquivalent(ctx, expr, base)
	}
	if ast.IsCallExpression(propAccess) {
		expr := propAccess.AsCallExpression().Expression
		// For call expressions, check if the expression itself is a property access of base
		return isPropertyAccessOf(ctx, expr, base)
	}

	return false
}

// Analyze a logical && expression for optional chain opportunities
func analyzeAndExpression(ctx rule.RuleContext, node *ast.Node) {
	if !ast.IsBinaryExpression(node) {
		return
	}

	binExpr := node.AsBinaryExpression()
	if binExpr.OperatorToken.Kind != ast.KindAmpersandAmpersandToken {
		return
	}

	left := binExpr.Left
	right := binExpr.Right

	// Pattern: foo && foo.bar
	// Pattern: foo !== null && foo.bar
	// Pattern: foo && foo.bar && foo.bar.baz

	var baseExpr *ast.Node
	var accessExpr *ast.Node

	// Check if left is a nullish check
	if isNullishCheck(left) {
		baseExpr = getCheckedExpression(left)
		accessExpr = right
	} else {
		baseExpr = left
		accessExpr = right
	}

	// Check if right is a property access of the left expression
	if !isPropertyAccessOf(ctx, accessExpr, baseExpr) {
		return
	}

	// Check if the right side ends with a comparison - if so, we can auto-fix
	// Otherwise, we provide a suggestion
	var hasSuggestionOnly bool
	var endNode *ast.Node = node

	// Walk through the && chain to find all connected parts
	var parts []*ast.Node
	var current *ast.Node = node
	for current != nil && ast.IsBinaryExpression(current) && current.AsBinaryExpression().OperatorToken.Kind == ast.KindAmpersandAmpersandToken {
		currentBin := current.AsBinaryExpression()
		parts = append([]*ast.Node{currentBin.Left}, parts...)
		current = currentBin.Right
		endNode = current
	}
	if current != nil {
		parts = append(parts, current)
	}

	// Check if the last part is a comparison
	if len(parts) > 0 {
		lastPart := parts[len(parts)-1]
		if isComparison(lastPart) {
			hasSuggestionOnly = false
		} else {
			hasSuggestionOnly = true
		}
	}

	// Build the replacement text
	replacement := buildOptionalChainReplacement(ctx, node)

	if replacement == "" {
		return
	}

	// Create the fix/suggestion
	fix := rule.RuleFixReplaceRange(
		utils.TrimNodeTextRange(ctx.SourceFile, node),
		replacement,
	)

	if hasSuggestionOnly {
		// Provide as suggestion only
		ctx.ReportNodeWithSuggestions(
			node,
			buildPreferOptionalChainMessage(),
			rule.RuleSuggestion{
				Message:  buildOptionalChainSuggestMessage(),
				FixesArr: []rule.RuleFix{fix},
			},
		)
	} else {
		// Provide as auto-fix with suggestion
		ctx.ReportNodeWithSuggestions(
			node,
			buildPreferOptionalChainMessage(),
			rule.RuleSuggestion{
				Message:  buildOptionalChainSuggestMessage(),
				FixesArr: []rule.RuleFix{fix},
			},
		)
	}
}

// Build the optional chain replacement for a logical expression
func buildOptionalChainReplacement(ctx rule.RuleContext, node *ast.Node) string {
	if node == nil {
		return ""
	}

	if !ast.IsBinaryExpression(node) {
		return ""
	}

	binExpr := node.AsBinaryExpression()
	op := binExpr.OperatorToken.Kind

	// Handle && chains
	if op == ast.KindAmpersandAmpersandToken {
		// Collect all parts of the && chain
		var parts []string
		var baseExpr *ast.Node

		// Simple pattern: foo && foo.bar
		left := binExpr.Left
		right := binExpr.Right

		// Determine base expression
		if isNullishCheck(left) {
			baseExpr = getCheckedExpression(left)
		} else {
			baseExpr = left
		}

		baseText := getNodeTextMaybeWrapped(ctx, baseExpr)

		// Try to convert the right side to optional chain
		rightText := convertToOptionalChain(ctx, right, baseExpr)
		if rightText == "" {
			return ""
		}

		return baseText + "?." + rightText
	}

	return ""
}

// Convert an expression to optional chain notation relative to a base
func convertToOptionalChain(ctx rule.RuleContext, node, base *ast.Node) string {
	if node == nil {
		return ""
	}

	// If this is a property access of base, convert it
	if ast.IsPropertyAccessExpression(node) {
		propAccess := node.AsPropertyAccessExpression()
		expr := propAccess.Expression

		if areExpressionsEquivalent(ctx, expr, base) {
			// foo.bar -> bar
			propName := propAccess.Name.AsIdentifier().EscapedText
			return propName
		}

		// Recursively handle nested accesses: foo.bar.baz
		prefix := convertToOptionalChain(ctx, expr, base)
		if prefix != "" {
			propName := propAccess.Name.AsIdentifier().EscapedText
			return prefix + "?." + propName
		}
	}

	if ast.IsElementAccessExpression(node) {
		elemAccess := node.AsElementAccessExpression()
		expr := elemAccess.Expression

		if areExpressionsEquivalent(ctx, expr, base) {
			// foo[bar] -> [bar]
			argText := ctx.SourceFile.Text()[utils.TrimNodeTextRange(ctx.SourceFile, elemAccess.ArgumentExpression).Pos():utils.TrimNodeTextRange(ctx.SourceFile, elemAccess.ArgumentExpression).End()]
			return "[" + argText + "]"
		}

		// Recursively handle nested accesses
		prefix := convertToOptionalChain(ctx, expr, base)
		if prefix != "" {
			argText := ctx.SourceFile.Text()[utils.TrimNodeTextRange(ctx.SourceFile, elemAccess.ArgumentExpression).Pos():utils.TrimNodeTextRange(ctx.SourceFile, elemAccess.ArgumentExpression).End()]
			return prefix + "?.[" + argText + "]"
		}
	}

	if ast.IsCallExpression(node) {
		callExpr := node.AsCallExpression()
		expr := callExpr.Expression

		// Handle foo.bar() where we're converting relative to foo
		if ast.IsPropertyAccessExpression(expr) || ast.IsElementAccessExpression(expr) {
			prefix := convertToOptionalChain(ctx, expr, base)
			if prefix != "" {
				// Get the arguments
				var args []string
				if callExpr.Arguments != nil {
					for _, arg := range callExpr.Arguments.Slice() {
						argText := ctx.SourceFile.Text()[utils.TrimNodeTextRange(ctx.SourceFile, arg).Pos():utils.TrimNodeTextRange(ctx.SourceFile, arg).End()]
						args = append(args, argText)
					}
				}
				argsText := strings.Join(args, ", ")
				return prefix + "(" + argsText + ")"
			}
		}
	}

	if ast.IsBinaryExpression(node) {
		binExpr := node.AsBinaryExpression()

		// If it's a comparison at the end, handle specially
		if isComparison(node) {
			left := binExpr.Left
			right := binExpr.Right
			op := binExpr.OperatorToken.Kind

			// Convert the left side
			leftConverted := convertToOptionalChain(ctx, left, base)
			if leftConverted != "" {
				// Get the operator text
				var opText string
				switch op {
				case ast.KindEqualsEqualsToken:
					opText = " == "
				case ast.KindEqualsEqualsEqualsToken:
					opText = " === "
				case ast.KindExclamationEqualsToken:
					opText = " != "
				case ast.KindExclamationEqualsEqualsToken:
					opText = " !== "
				}

				rightText := ctx.SourceFile.Text()[utils.TrimNodeTextRange(ctx.SourceFile, right).Pos():utils.TrimNodeTextRange(ctx.SourceFile, right).End()]
				return leftConverted + opText + rightText
			}
		}

		// Handle && chains recursively
		if binExpr.OperatorToken.Kind == ast.KindAmpersandAmpersandToken {
			left := binExpr.Left
			right := binExpr.Right

			// Try to build a chain from the parts
			var leftBase *ast.Node
			if isNullishCheck(left) {
				leftBase = getCheckedExpression(left)
			} else {
				leftBase = left
			}

			// Check if leftBase matches our base
			if areExpressionsEquivalent(ctx, leftBase, base) {
				rightConverted := convertToOptionalChain(ctx, right, base)
				return rightConverted
			}
		}
	}

	return ""
}

// Handle the (foo || {}).bar pattern
func analyzeEmptyObjectPattern(ctx rule.RuleContext, node *ast.Node) {
	if !ast.IsBinaryExpression(node) {
		return
	}

	binExpr := node.AsBinaryExpression()
	if binExpr.OperatorToken.Kind != ast.KindBarBarToken && binExpr.OperatorToken.Kind != ast.KindQuestionQuestionToken {
		return
	}

	// Check if the right side is an empty object literal
	if !isEmptyObjectLiteral(binExpr.Right) {
		return
	}

	// Check if the parent is a property or element access
	parent := utils.GetParent(node)
	if parent == nil {
		return
	}

	var isOptionalAccess bool
	var propertyText string
	var isComputed bool

	if ast.IsPropertyAccessExpression(parent) {
		propAccess := parent.AsPropertyAccessExpression()
		// Make sure the logical expression is the base
		if propAccess.Expression != node {
			return
		}
		// Skip if already optional
		if propAccess.QuestionDotToken != nil {
			return
		}
		isOptionalAccess = false
		propertyText = propAccess.Name.AsIdentifier().EscapedText
		isComputed = false
	} else if ast.IsElementAccessExpression(parent) {
		elemAccess := parent.AsElementAccessExpression()
		// Make sure the logical expression is the base
		if elemAccess.Expression != node {
			return
		}
		// Skip if already optional
		if elemAccess.QuestionDotToken != nil {
			return
		}
		isOptionalAccess = false
		propertyText = ctx.SourceFile.Text()[utils.TrimNodeTextRange(ctx.SourceFile, elemAccess.ArgumentExpression).Pos():utils.TrimNodeTextRange(ctx.SourceFile, elemAccess.ArgumentExpression).End()]
		isComputed = true
	} else {
		return
	}

	// Build replacement
	leftText := getNodeTextMaybeWrapped(ctx, binExpr.Left)
	var replacement string
	if isComputed {
		replacement = fmt.Sprintf("%s?.[%s]", leftText, propertyText)
	} else {
		replacement = fmt.Sprintf("%s?.%s", leftText, propertyText)
	}

	// Create fix
	fix := rule.RuleFixReplaceRange(
		utils.TrimNodeTextRange(ctx.SourceFile, parent),
		replacement,
	)

	ctx.ReportNodeWithSuggestions(
		parent,
		buildPreferOptionalChainMessage(),
		rule.RuleSuggestion{
			Message:  buildOptionalChainSuggestMessage(),
			FixesArr: []rule.RuleFix{fix},
		},
	)
}

var PreferOptionalChainRule = rule.CreateRule(rule.Rule{
	Name: "prefer-optional-chain",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				if !ast.IsBinaryExpression(node) {
					return
				}

				binExpr := node.AsBinaryExpression()
				op := binExpr.OperatorToken.Kind

				// Handle && expressions
				if op == ast.KindAmpersandAmpersandToken {
					analyzeAndExpression(ctx, node)
				}

				// Handle || and ?? with empty object patterns
				if op == ast.KindBarBarToken || op == ast.KindQuestionQuestionToken {
					analyzeEmptyObjectPattern(ctx, node)
				}
			},
		}
	},
})
