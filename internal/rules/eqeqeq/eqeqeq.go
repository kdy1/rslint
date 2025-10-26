package eqeqeq

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// EqeqeqOptions defines the configuration options for this rule
type EqeqeqOptions struct {
	Mode       string // "always", "smart", "allow-null"
	NullOption string // "always", "never", "ignore" (only used when Mode == "always")
}

// parseOptions parses and validates the rule options
func parseOptions(options any) EqeqeqOptions {
	opts := EqeqeqOptions{
		Mode:       "always",
		NullOption: "always",
	}

	if options == nil {
		return opts
	}

	// Handle array format: ["always", { "null": "ignore" }] or ["allow-null"]
	if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
		// First element is the mode
		if modeStr, ok := optArray[0].(string); ok {
			opts.Mode = modeStr
			// Special case: "allow-null" is shorthand for "always" with null: "ignore"
			if modeStr == "allow-null" {
				opts.Mode = "always"
				opts.NullOption = "ignore"
			}
		}
		// Second element might be an options object
		if len(optArray) > 1 {
			if optsMap, ok := optArray[1].(map[string]interface{}); ok {
				if v, ok := optsMap["null"].(string); ok {
					opts.NullOption = v
				}
			}
		}
	} else if optsMap, ok := options.(map[string]interface{}); ok {
		if v, ok := optsMap["mode"].(string); ok {
			opts.Mode = v
		}
		if v, ok := optsMap["null"].(string); ok {
			opts.NullOption = v
		}
	}

	return opts
}

func buildExpectedMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "expected",
		Description: "Expected '===' and instead saw '=='.",
	}
}

func buildExpectedNotMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "expectedNot",
		Description: "Expected '!==' and instead saw '!='.",
	}
}

func buildUnexpectedMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpected",
		Description: "Expected '==' and instead saw '==='.",
	}
}

func buildUnexpectedNotMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedNot",
		Description: "Expected '!=' and instead saw '!=='.",
	}
}

// EqeqeqRule implements the eqeqeq rule
// Require the use of === and !==
var EqeqeqRule = rule.Rule{
	Name: "eqeqeq",
	Run:  run,
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := parseOptions(options)

	return rule.RuleListeners{
		ast.KindBinaryExpression: func(node *ast.Node) {
			binExpr := node.AsBinaryExpression()
			if binExpr == nil {
				return
			}

			op := binExpr.OperatorToken.Kind
			isLooseEquality := op == ast.KindEqualsEqualsToken
			isLooseInequality := op == ast.KindExclamationEqualsToken
			isStrictEquality := op == ast.KindEqualsEqualsEqualsToken
			isStrictInequality := op == ast.KindExclamationEqualsEqualsToken

			if !isLooseEquality && !isLooseInequality && !isStrictEquality && !isStrictInequality {
				return
			}

			left := binExpr.Left
			right := binExpr.Right

			// Check null comparisons
			isNullComparison := isNullLiteral(left) || isNullLiteral(right)

			if opts.Mode == "always" {
				// In "always" mode, enforce strict equality
				if isNullComparison {
					// Handle null option
					if opts.NullOption == "ignore" {
						// Allow both == and === for null
						return
					} else if opts.NullOption == "never" {
						// Require == for null
						if isStrictEquality {
							fix := replaceOperator(ctx, node, "===", "==")
							ctx.ReportNodeWithSuggestions(node, buildUnexpectedMessage(), fix)
						} else if isStrictInequality {
							fix := replaceOperator(ctx, node, "!==", "!=")
							ctx.ReportNodeWithSuggestions(node, buildUnexpectedNotMessage(), fix)
						}
						return
					}
					// else opts.NullOption == "always": enforce === for null (fall through)
				}

				// Enforce strict equality
				if isLooseEquality {
					if canAutoFix(left, right) {
						fix := replaceOperator(ctx, node, "==", "===")
						ctx.ReportNodeWithFixes(node, buildExpectedMessage(), fix)
					} else {
						fix := replaceOperator(ctx, node, "==", "===")
						ctx.ReportNodeWithSuggestions(node, buildExpectedMessage(), fix)
					}
				} else if isLooseInequality {
					if canAutoFix(left, right) {
						fix := replaceOperator(ctx, node, "!=", "!==")
						ctx.ReportNodeWithFixes(node, buildExpectedNotMessage(), fix)
					} else {
						fix := replaceOperator(ctx, node, "!=", "!==")
						ctx.ReportNodeWithSuggestions(node, buildExpectedNotMessage(), fix)
					}
				}
			} else if opts.Mode == "smart" {
				// In "smart" mode, allow == for:
				// 1. typeof comparisons
				// 2. comparing against literals of the same type
				// 3. null comparisons
				if isNullComparison {
					return
				}
				if isTypeOfBinary(left, right) {
					return
				}
				if areLiteralsAndSameType(left, right) {
					return
				}

				// Otherwise, enforce strict equality
				if isLooseEquality {
					fix := replaceOperator(ctx, node, "==", "===")
					ctx.ReportNodeWithFixes(node, buildExpectedMessage(), fix)
				} else if isLooseInequality {
					fix := replaceOperator(ctx, node, "!=", "!==")
					ctx.ReportNodeWithFixes(node, buildExpectedNotMessage(), fix)
				}
			}
		},
	}
}

func isNullLiteral(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return node.Kind == ast.KindNullKeyword
}

func isTypeOfExpression(node *ast.Node) bool {
	if node == nil {
		return false
	}
	prefix := node.AsPrefixUnaryExpression()
	return prefix != nil && prefix.Operator == ast.KindTypeOfKeyword
}

func isTypeOfBinary(left, right *ast.Node) bool {
	return isTypeOfExpression(left) || isTypeOfExpression(right)
}

func isLiteral(node *ast.Node) bool {
	if node == nil {
		return false
	}
	kind := node.Kind
	return kind == ast.KindStringLiteral ||
		kind == ast.KindNumericLiteral ||
		kind == ast.KindTrueKeyword ||
		kind == ast.KindFalseKeyword ||
		kind == ast.KindNullKeyword ||
		kind == ast.KindBigIntLiteral ||
		kind == ast.KindNoSubstitutionTemplateLiteral
}

func getLiteralType(node *ast.Node) string {
	if node == nil {
		return ""
	}
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return "string"
	case ast.KindNumericLiteral:
		return "number"
	case ast.KindTrueKeyword, ast.KindFalseKeyword:
		return "boolean"
	case ast.KindNullKeyword:
		return "null"
	case ast.KindBigIntLiteral:
		return "bigint"
	default:
		return ""
	}
}

func areLiteralsAndSameType(left, right *ast.Node) bool {
	if !isLiteral(left) || !isLiteral(right) {
		return false
	}
	leftType := getLiteralType(left)
	rightType := getLiteralType(right)
	return leftType != "" && leftType == rightType
}

func canAutoFix(left, right *ast.Node) bool {
	// Auto-fix is safe when:
	// 1. Both sides are typeof expressions
	// 2. Both sides are literals of the same type
	if isTypeOfBinary(left, right) {
		return true
	}
	if areLiteralsAndSameType(left, right) {
		return true
	}
	return false
}

func replaceOperator(ctx rule.RuleContext, node *ast.Node, oldOp, newOp string) rule.RuleFix {
	binExpr := node.AsBinaryExpression()
	if binExpr == nil {
		return rule.RuleFix{}
	}

	// Get the full text and replace the operator
	nodeText := utils.GetNodeText(node)
	leftText := utils.GetNodeText(binExpr.Left)
	rightText := utils.GetNodeText(binExpr.Right)

	// Reconstruct with new operator
	newText := leftText + " " + newOp + " " + rightText

	return rule.RuleFix{
		Message: "Replace with strict equality",
		Edits: []rule.TextEdit{
			rule.RuleFixReplace(ctx.SourceFile, node, newText),
		},
	}
}
