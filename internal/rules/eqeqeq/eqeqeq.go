package eqeqeq

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Options for the eqeqeq rule
type Options struct {
	Mode string `json:"mode"` // "always" (default), "smart", or can be in first position
	Null string `json:"null"` // "always", "never", "ignore" (default)
}

func parseOptions(options any) Options {
	opts := Options{
		Mode: "always",
		Null: "always",
	}

	if options == nil {
		return opts
	}

	// Handle array format: ["smart"] or ["always", { "null": "ignore" }]
	if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
		// First element can be a string mode or an object
		if modeStr, isStr := optArray[0].(string); isStr {
			opts.Mode = modeStr
		}

		// Second element is options object
		if len(optArray) > 1 {
			if optsMap, ok := optArray[1].(map[string]interface{}); ok {
				if nullVal, ok := optsMap["null"].(string); ok {
					opts.Null = nullVal
				}
			}
		} else if optsMap, ok := optArray[0].(map[string]interface{}); ok {
			// First element is an object
			if nullVal, ok := optsMap["null"].(string); ok {
				opts.Null = nullVal
			}
		}
	} else if optsMap, ok := options.(map[string]interface{}); ok {
		// Handle direct object format
		if nullVal, ok := optsMap["null"].(string); ok {
			opts.Null = nullVal
		}
	}

	return opts
}

func buildExpectedMessage(operator string) rule.RuleMessage {
	expected := "==="
	if operator == "!=" {
		expected = "!=="
	}
	return rule.RuleMessage{
		Id:          "unexpected",
		Description: "Expected '" + expected + "' and instead saw '" + operator + "'.",
	}
}

func buildExpectedNullMessage(operator string) rule.RuleMessage {
	expected := "=="
	if operator == "!==" {
		expected = "!="
	}
	return rule.RuleMessage{
		Id:          "unexpectedNull",
		Description: "Expected '" + expected + "' and instead saw '" + operator + "'.",
	}
}

// isNullLiteral checks if a node is a null literal
func isNullLiteral(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if node.Kind == ast.KindNullKeyword {
		return true
	}
	return false
}

// isTypeofExpression checks if a node is a typeof expression
func isTypeofExpression(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if node.Kind == ast.KindTypeOfExpression {
		return true
	}
	return false
}

// isLiteral checks if a node is a literal value
func isLiteral(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindNumericLiteral, ast.KindStringLiteral, ast.KindTrueKeyword,
		ast.KindFalseKeyword, ast.KindNullKeyword, ast.KindRegularExpressionLiteral,
		ast.KindNoSubstitutionTemplateLiteral, ast.KindBigIntLiteral:
		return true
	}
	return false
}

// isNullComparison checks if binary expression is comparing with null
func isNullComparison(left *ast.Node, right *ast.Node) bool {
	return isNullLiteral(left) || isNullLiteral(right)
}

// isSmartAllowed checks if the comparison is allowed in "smart" mode
func isSmartAllowed(left *ast.Node, right *ast.Node) bool {
	// typeof expressions are allowed
	if isTypeofExpression(left) || isTypeofExpression(right) {
		return true
	}

	// Two literals are allowed
	if isLiteral(left) && isLiteral(right) {
		return true
	}

	return false
}

var EqeqeqRule = rule.CreateRule(rule.Rule{
	Name: "eqeqeq",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		listeners := rule.RuleListeners{}

		listeners[ast.KindBinaryExpression] = func(node *ast.Node) {
			binExpr := node.AsBinaryExpression()
			if binExpr == nil || binExpr.OperatorToken == nil {
				return
			}

			operator := binExpr.OperatorToken.Kind
			left := binExpr.Left
			right := binExpr.Right

			// Check if this is a null comparison
			isNull := isNullComparison(left, right)

			// Handle === and !== when null option is "never"
			if opts.Null == "never" {
				if operator == ast.KindEqualsEqualsEqualsToken || operator == ast.KindExclamationEqualsEqualsToken {
					// Only report if comparing with null
					if isNullComparison(left, right) {
						operatorText := "==="
						fixText := "=="
						if operator == ast.KindExclamationEqualsEqualsToken {
							operatorText = "!=="
							fixText = "!="
						}

						ctx.ReportNodeWithFixes(
							node,
							buildExpectedNullMessage(operatorText),
							rule.RuleFixReplace(
								ctx.SourceFile,
								binExpr.OperatorToken,
								fixText,
							),
						)
					}
					return
				}
			}

			// Only care about == and !=
			if operator != ast.KindEqualsEqualsToken && operator != ast.KindExclamationEqualsToken {
				return
			}

			// Handle null option
			if isNull && opts.Null == "ignore" {
				return
			}

			// Handle smart mode
			if opts.Mode == "smart" {
				// In smart mode, null comparisons are allowed
				if isNull {
					return
				}
				// Other smart mode checks
				if isSmartAllowed(left, right) {
					return
				}
			}

			// Report the error and provide auto-fix
			operatorText := "=="
			if operator == ast.KindExclamationEqualsToken {
				operatorText = "!="
			}

			fixText := "==="
			if operator == ast.KindExclamationEqualsToken {
				fixText = "!=="
			}

			ctx.ReportNodeWithFixes(
				node,
				buildExpectedMessage(operatorText),
				rule.RuleFixReplace(
					ctx.SourceFile,
					binExpr.OperatorToken,
					fixText,
				),
			)
		}

		return listeners
	},
})
