package no_array_constructor

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildUseLiteralMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useLiteral",
		Description: "The array literal notation [] is preferable.",
	}
}

var NoArrayConstructorRule = rule.CreateRule(rule.Rule{
	Name: "no-array-constructor",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		checkArrayConstructor := func(node *ast.Node, args []*ast.Node, isOptionalChain bool) {
			// Get the callee node - it could be a simple identifier or a member expression
			var callee *ast.Node
			var hasTypeArguments bool

			if node.Kind == ast.KindNewExpression {
				newExpr := node.AsNewExpression()
				callee = newExpr.Expression
				hasTypeArguments = newExpr.TypeArguments != nil
			} else if node.Kind == ast.KindCallExpression {
				callExpr := node.AsCallExpression()
				callee = callExpr.Expression
				hasTypeArguments = callExpr.TypeArguments != nil
			} else {
				return
			}

			// Skip if it's not an identifier or if the identifier is not "Array"
			if callee.Kind != ast.KindIdentifier {
				return
			}

			identifier := callee.AsIdentifier()
			if identifier.Text() != "Array" {
				return
			}

			// Skip if there are type arguments - TypeScript generic syntax like Array<Foo>()
			if hasTypeArguments {
				return
			}

			// Skip if there is exactly one argument (could be specifying array length)
			// Exception: if it's optional chaining, we still report it
			if len(args) == 1 && !isOptionalChain {
				return
			}

			// Create the fix
			fix := createArrayLiteralFix(ctx, node, args)
			ctx.ReportNodeWithFixes(node, buildUseLiteralMessage(), fix)
		}

		return rule.RuleListeners{
			ast.KindNewExpression: func(node *ast.Node) {
				if node.Kind != ast.KindNewExpression {
					return
				}
				newExpr := node.AsNewExpression()
				args := []*ast.Node{}
				if newExpr.Arguments != nil {
					args = newExpr.Arguments
				}
				checkArrayConstructor(node, args, false)
			},
			ast.KindCallExpression: func(node *ast.Node) {
				if node.Kind != ast.KindCallExpression {
					return
				}
				callExpr := node.AsCallExpression()

				// Check if this is an optional chain call (e.g., Array?.())
				isOptionalChain := callExpr.QuestionDotToken.Pos() > 0

				args := []*ast.Node{}
				if callExpr.Arguments != nil {
					args = callExpr.Arguments
				}
				checkArrayConstructor(node, args, isOptionalChain)
			},
		}
	},
})

func createArrayLiteralFix(ctx rule.RuleContext, node *ast.Node, args []*ast.Node) rule.RuleFix {
	sourceText := ctx.SourceFile.Text
	nodeRange := utils.TrimNodeTextRange(ctx.SourceFile, node)

	// Handle empty array case
	if len(args) == 0 {
		// For empty arrays, replace everything with []
		return rule.RuleFixReplaceRange(nodeRange, "[]")
	}

	// Get the original node text
	originalText := sourceText[nodeRange.Pos():nodeRange.End()]

	// Find the opening parenthesis
	openParenIdx := strings.Index(originalText, "(")
	if openParenIdx == -1 {
		// Fallback: just replace the whole thing
		return rule.RuleFixReplaceRange(nodeRange, "[]")
	}

	// Find the closing parenthesis (last one in the text)
	closeParenIdx := strings.LastIndex(originalText, ")")
	if closeParenIdx == -1 {
		// Fallback: just replace the whole thing
		return rule.RuleFixReplaceRange(nodeRange, "[]")
	}

	// Extract the content between parentheses
	argsContent := originalText[openParenIdx+1 : closeParenIdx]

	// Build replacement
	replacement := "[" + argsContent + "]"

	return rule.RuleFixReplaceRange(nodeRange, replacement)
}

func getLeadingCommentsRange(ctx rule.RuleContext, node *ast.Node) (core.TextRange, bool) {
	sourceText := ctx.SourceFile.Text
	nodeStart := node.Pos()

	// Scan backwards to find the start of leading comments
	startPos := nodeStart

	for startPos > 0 {
		// Skip whitespace
		for startPos > 0 && (sourceText[startPos-1] == ' ' || sourceText[startPos-1] == '\t') {
			startPos--
		}

		// Check if we're at a comment
		if startPos >= 2 && sourceText[startPos-2:startPos] == "*/" {
			// Found end of block comment, scan to find start
			endPos := startPos
			startPos -= 2
			foundStart := false
			for startPos >= 2 {
				if sourceText[startPos-2:startPos] == "/*" {
					startPos -= 2
					foundStart = true
					break
				}
				startPos--
			}
			if !foundStart {
				break
			}
		} else {
			break
		}
	}

	if startPos < nodeStart {
		return core.CreateTextRange(startPos, nodeStart), true
	}

	return core.CreateTextRange(0, 0), false
}

func getTrailingCommentsRange(ctx rule.RuleContext, node *ast.Node) (core.TextRange, bool) {
	sourceText := ctx.SourceFile.Text
	nodeEnd := node.End()

	// Scan forward to find trailing comments
	endPos := nodeEnd

	// Skip whitespace
	for endPos < len(sourceText) && (sourceText[endPos] == ' ' || sourceText[endPos] == '\t') {
		endPos++
	}

	// Check for comments
	if endPos+1 < len(sourceText) && sourceText[endPos:endPos+2] == "/*" {
		// Found start of block comment, scan to find end
		endPos += 2
		for endPos+1 < len(sourceText) {
			if sourceText[endPos:endPos+2] == "*/" {
				endPos += 2
				break
			}
			endPos++
		}
	}

	if endPos > nodeEnd {
		return core.CreateTextRange(nodeEnd, endPos), true
	}

	return core.CreateTextRange(0, 0), false
}
