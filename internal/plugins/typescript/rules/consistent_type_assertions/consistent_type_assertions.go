package consistent_type_assertions

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// ConsistentTypeAssertionsOptions defines the configuration
type ConsistentTypeAssertionsOptions struct {
	AssertionStyle               string `json:"assertionStyle"` // "as", "angle-bracket", or "never"
	ObjectLiteralTypeAssertions  string `json:"objectLiteralTypeAssertions"`
	ArrayLiteralTypeAssertions   string `json:"arrayLiteralTypeAssertions"`
}

func parseOptions(options interface{}) ConsistentTypeAssertionsOptions {
	opts := ConsistentTypeAssertionsOptions{
		AssertionStyle:              "as",      // Default
		ObjectLiteralTypeAssertions: "allow",   // Default
		ArrayLiteralTypeAssertions:  "allow",   // Default
	}

	if options == nil {
		return opts
	}

	switch v := options.(type) {
	case map[string]interface{}:
		if style, ok := v["assertionStyle"].(string); ok {
			if style == "as" || style == "angle-bracket" || style == "never" {
				opts.AssertionStyle = style
			}
		}
		if objLiteral, ok := v["objectLiteralTypeAssertions"].(string); ok {
			if objLiteral == "allow" || objLiteral == "allow-as-parameter" || objLiteral == "never" {
				opts.ObjectLiteralTypeAssertions = objLiteral
			}
		}
		if arrLiteral, ok := v["arrayLiteralTypeAssertions"].(string); ok {
			if arrLiteral == "allow" || arrLiteral == "allow-as-parameter" || arrLiteral == "never" {
				opts.ArrayLiteralTypeAssertions = arrLiteral
			}
		}
	}

	return opts
}

func buildUnexpectedTypeAssertionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedTypeAssertion",
		Description: "Use type annotation instead of type assertion.",
	}
}

func buildPreferAsMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "as",
		Description: "Use 'as' instead of angle-bracket type assertions.",
	}
}

func buildPreferAngleBracketMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "angle-bracket",
		Description: "Use angle-bracket type assertions instead of 'as'.",
	}
}

func buildNeverAssertionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "never",
		Description: "Type assertions are not allowed.",
	}
}

func buildObjectLiteralNeverMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedObjectTypeAssertion",
		Description: "Use type annotation instead of type assertion for object literals.",
	}
}

func buildArrayLiteralNeverMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedArrayTypeAssertion",
		Description: "Use type annotation instead of type assertion for array literals.",
	}
}

// Check if an expression is an object literal
func isObjectLiteral(node *ast.Node) bool {
	return node != nil && node.Kind == ast.KindObjectLiteralExpression
}

// Check if an expression is an array literal
func isArrayLiteral(node *ast.Node) bool {
	return node != nil && node.Kind == ast.KindArrayLiteralExpression
}

// Check if the assertion is to any or unknown (bypass restrictions)
func isAnyOrUnknownAssertion(typeNode *ast.Node) bool {
	if typeNode == nil {
		return false
	}

	if typeNode.Kind == ast.KindAnyKeyword || typeNode.Kind == ast.KindUnknownKeyword {
		return true
	}

	return false
}

// Check if we're inside a call expression parameter
func isInsideCallExpression(node *ast.Node) bool {
	parent := node.Parent
	for parent != nil {
		if parent.Kind == ast.KindCallExpression {
			return true
		}
		parent = parent.Parent
	}
	return false
}

// Convert angle-bracket assertion to 'as' assertion
func convertToAsAssertion(ctx rule.RuleContext, typeAssertion *ast.TypeAssertion) string {
	if typeAssertion == nil {
		return ""
	}

	exprRange := utils.TrimNodeTextRange(ctx.SourceFile, typeAssertion.Expression)
	exprText := ctx.SourceFile.Text()[exprRange.Pos():exprRange.End()]

	typeRange := utils.TrimNodeTextRange(ctx.SourceFile, typeAssertion.Type)
	typeText := ctx.SourceFile.Text()[typeRange.Pos():typeRange.End()]

	return exprText + " as " + typeText
}

// Convert 'as' assertion to angle-bracket assertion
func convertToAngleBracket(ctx rule.RuleContext, asExpr *ast.AsExpression) string {
	if asExpr == nil {
		return ""
	}

	exprRange := utils.TrimNodeTextRange(ctx.SourceFile, asExpr.Expression)
	exprText := ctx.SourceFile.Text()[exprRange.Pos():exprRange.End()]

	typeRange := utils.TrimNodeTextRange(ctx.SourceFile, asExpr.Type)
	typeText := ctx.SourceFile.Text()[typeRange.Pos():typeRange.End()]

	return "<" + typeText + ">" + exprText
}

var ConsistentTypeAssertionsRule = rule.CreateRule(rule.Rule{
	Name: "consistent-type-assertions",
	Run: func(ctx rule.RuleContext, options interface{}) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			// Handle 'as' expressions (value as Type)
			ast.KindAsExpression: func(node *ast.Node) {
				if node.Kind != ast.KindAsExpression {
					return
				}

				asExpr := node.AsAsExpression()
				if asExpr == nil {
					return
				}

				// Skip const assertions - always allowed
				if asExpr.Type != nil && asExpr.Type.Kind == ast.KindTypeOperator {
					typeOp := asExpr.Type.AsTypeOperatorNode()
					if typeOp != nil && typeOp.Operator == ast.KindConstKeyword {
						return
					}
				}

				// Check assertion style preference
				if opts.AssertionStyle == "angle-bracket" {
					replacement := convertToAngleBracket(ctx, asExpr)
					ctx.ReportNodeWithFixes(
						node,
						buildPreferAngleBracketMessage(),
						rule.RuleFixReplace(ctx.SourceFile, node, replacement),
					)
					return
				}

				if opts.AssertionStyle == "never" {
					// Skip any/unknown assertions
					if !isAnyOrUnknownAssertion(asExpr.Type) {
						ctx.ReportNode(node, buildNeverAssertionMessage())
						return
					}
				}

				// Check object literal assertions
				if opts.ObjectLiteralTypeAssertions == "never" && isObjectLiteral(asExpr.Expression) {
					if !isAnyOrUnknownAssertion(asExpr.Type) {
						ctx.ReportNode(node, buildObjectLiteralNeverMessage())
						return
					}
				}

				if opts.ObjectLiteralTypeAssertions == "allow-as-parameter" && isObjectLiteral(asExpr.Expression) {
					if !isInsideCallExpression(node) && !isAnyOrUnknownAssertion(asExpr.Type) {
						ctx.ReportNode(node, buildObjectLiteralNeverMessage())
						return
					}
				}

				// Check array literal assertions
				if opts.ArrayLiteralTypeAssertions == "never" && isArrayLiteral(asExpr.Expression) {
					if !isAnyOrUnknownAssertion(asExpr.Type) {
						ctx.ReportNode(node, buildArrayLiteralNeverMessage())
						return
					}
				}

				if opts.ArrayLiteralTypeAssertions == "allow-as-parameter" && isArrayLiteral(asExpr.Expression) {
					if !isInsideCallExpression(node) && !isAnyOrUnknownAssertion(asExpr.Type) {
						ctx.ReportNode(node, buildArrayLiteralNeverMessage())
						return
					}
				}
			},

			// Handle angle-bracket assertions (<Type>value)
			ast.KindTypeAssertionExpression: func(node *ast.Node) {
				if node.Kind != ast.KindTypeAssertionExpression {
					return
				}

				typeAssertion := node.AsTypeAssertion()
				if typeAssertion == nil {
					return
				}

				// Skip const assertions - always allowed
				if typeAssertion.Type != nil && typeAssertion.Type.Kind == ast.KindTypeOperator {
					typeOp := typeAssertion.Type.AsTypeOperatorNode()
					if typeOp != nil && typeOp.Operator == ast.KindConstKeyword {
						return
					}
				}

				// Check assertion style preference
				if opts.AssertionStyle == "as" {
					replacement := convertToAsAssertion(ctx, typeAssertion)
					ctx.ReportNodeWithFixes(
						node,
						buildPreferAsMessage(),
						rule.RuleFixReplace(ctx.SourceFile, node, replacement),
					)
					return
				}

				if opts.AssertionStyle == "never" {
					// Skip any/unknown assertions
					if !isAnyOrUnknownAssertion(typeAssertion.Type) {
						ctx.ReportNode(node, buildNeverAssertionMessage())
						return
					}
				}

				// Check object literal assertions
				if opts.ObjectLiteralTypeAssertions == "never" && isObjectLiteral(typeAssertion.Expression) {
					if !isAnyOrUnknownAssertion(typeAssertion.Type) {
						ctx.ReportNode(node, buildObjectLiteralNeverMessage())
						return
					}
				}

				if opts.ObjectLiteralTypeAssertions == "allow-as-parameter" && isObjectLiteral(typeAssertion.Expression) {
					if !isInsideCallExpression(node) && !isAnyOrUnknownAssertion(typeAssertion.Type) {
						ctx.ReportNode(node, buildObjectLiteralNeverMessage())
						return
					}
				}

				// Check array literal assertions
				if opts.ArrayLiteralTypeAssertions == "never" && isArrayLiteral(typeAssertion.Expression) {
					if !isAnyOrUnknownAssertion(typeAssertion.Type) {
						ctx.ReportNode(node, buildArrayLiteralNeverMessage())
						return
					}
				}

				if opts.ArrayLiteralTypeAssertions == "allow-as-parameter" && isArrayLiteral(typeAssertion.Expression) {
					if !isInsideCallExpression(node) && !isAnyOrUnknownAssertion(typeAssertion.Type) {
						ctx.ReportNode(node, buildArrayLiteralNeverMessage())
						return
					}
				}
			},
		}
	},
})
