package prefer_template

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// PreferTemplateRule implements the prefer-template rule
// Require template literals instead of string concatenation
var PreferTemplateRule = rule.Rule{
	Name: "prefer-template",
	Run:  run,
}

func buildMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedStringConcatenation",
		Description: "Unexpected string concatenation.",
	}
}

// isStringType checks if a node is a string literal or template
func isStringType(node *ast.Node) bool {
	if node == nil {
		return false
	}
	kind := node.Kind
	return kind == ast.KindStringLiteral ||
		kind == ast.KindNoSubstitutionTemplateLiteral ||
		kind == ast.KindTemplateExpression
}

// getNodeText returns the trimmed text of a node
func getNodeText(srcFile *ast.SourceFile, node *ast.Node) string {
	rng := utils.TrimNodeTextRange(srcFile, node)
	return srcFile.Text()[rng.Pos():rng.End()]
}

// isStringLiteral checks if node is a plain string literal
func isStringLiteral(node *ast.Node) bool {
	return node != nil && node.Kind == ast.KindStringLiteral
}

// hasOctalOrNonOctalDecimalEscape checks if a string contains octal escapes or non-octal decimal escapes
func hasOctalOrNonOctalDecimalEscape(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			next := s[i+1]
			// Check for octal escapes (0-7)
			if next >= '0' && next <= '7' {
				return true
			}
			// Check for non-octal decimal escapes (8-9)
			if next == '8' || next == '9' {
				return true
			}
		}
	}
	return false
}

// unquoteString removes quotes from a string literal
func unquoteString(s string) string {
	if len(s) >= 2 && (s[0] == '"' || s[0] == '\'' || s[0] == '`') {
		return s[1 : len(s)-1]
	}
	return s
}

// convertToTemplatePart converts a node to a template literal part
func convertToTemplatePart(srcFile *ast.SourceFile, node *ast.Node) string {
	text := getNodeText(srcFile, node)

	if isStringLiteral(node) {
		// Remove quotes and escape backticks and ${
		inner := unquoteString(text)
		inner = strings.ReplaceAll(inner, "\\\"", "\"") // Unescape double quotes
		inner = strings.ReplaceAll(inner, "\\'", "'")   // Unescape single quotes
		inner = strings.ReplaceAll(inner, "`", "\\`")    // Escape backticks
		inner = strings.ReplaceAll(inner, "${", "\\${")  // Escape template expressions
		return inner
	}

	// For expressions, wrap in ${}
	return "${" + strings.TrimSpace(text) + "}"
}

// canConcat checks if this binary expression is part of string concatenation
func canConcat(ctx rule.RuleContext, node *ast.Node) bool {
	binExpr := node.AsBinaryExpression()
	if binExpr == nil {
		return false
	}

	// Must be + operator
	if binExpr.OperatorToken == nil || binExpr.OperatorToken.Kind != ast.KindPlusToken {
		return false
	}

	left := binExpr.Left
	right := binExpr.Right

	if left == nil || right == nil {
		return false
	}

	// At least one side must be a string
	return isStringType(left) || isStringType(right)
}

// buildTemplateLiteral recursively builds template literal from concatenation
func buildTemplateLiteral(srcFile *ast.SourceFile, node *ast.Node) (string, bool) {
	if isStringLiteral(node) {
		text := getNodeText(srcFile, node)
		// Check for octal/non-octal decimal escapes which can't be converted
		if hasOctalOrNonOctalDecimalEscape(text) {
			return "", false
		}
		return convertToTemplatePart(srcFile, node), true
	}

	if node.Kind == ast.KindBinaryExpression {
		binExpr := node.AsBinaryExpression()
		if binExpr != nil && binExpr.OperatorToken != nil && binExpr.OperatorToken.Kind == ast.KindPlusToken {
			left := binExpr.Left
			right := binExpr.Right

			leftPart, leftOk := buildTemplateLiteral(srcFile, left)
			rightPart, rightOk := buildTemplateLiteral(srcFile, right)

			if !leftOk || !rightOk {
				return "", false
			}

			return leftPart + rightPart, true
		}
	}

	// For non-string expressions
	return convertToTemplatePart(srcFile, node), true
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	return rule.RuleListeners{
		ast.KindBinaryExpression: func(node *ast.Node) {
			if !canConcat(ctx, node) {
				return
			}

			// Don't report if this is part of a larger concatenation (let parent handle it)
			if node.Parent != nil && node.Parent.Kind == ast.KindBinaryExpression {
				parentBin := node.Parent.AsBinaryExpression()
				if parentBin != nil && parentBin.OperatorToken != nil &&
					parentBin.OperatorToken.Kind == ast.KindPlusToken &&
					(isStringType(parentBin.Left) || isStringType(parentBin.Right)) {
					return
				}
			}

			// Try to build the template literal
			templateContent, ok := buildTemplateLiteral(ctx.SourceFile, node)
			if !ok {
				// Can't auto-fix (e.g., contains octal escapes), but still report
				ctx.ReportNode(node, buildMessage())
				return
			}

			replacement := "`" + templateContent + "`"
			ctx.ReportNodeWithFixes(node, buildMessage(), rule.RuleFixReplace(ctx.SourceFile, node, replacement))
		},
	}
}
