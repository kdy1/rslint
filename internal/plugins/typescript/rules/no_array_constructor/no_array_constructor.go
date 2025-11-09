package no_array_constructor

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
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
			if identifier.Text != "Array" {
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
					args = newExpr.Arguments.Nodes
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
					args = callExpr.Arguments.Nodes
				}
				checkArrayConstructor(node, args, isOptionalChain)
			},
		}
	},
})

func createArrayLiteralFix(ctx rule.RuleContext, node *ast.Node, args []*ast.Node) rule.RuleFix {
	sourceText := ctx.SourceFile.Text()
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
