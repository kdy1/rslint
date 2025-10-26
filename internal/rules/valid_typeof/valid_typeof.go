package valid_typeof

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// ValidTypeofOptions defines the configuration options for this rule
type ValidTypeofOptions struct {
	RequireStringLiterals bool `json:"requireStringLiterals"`
}

// parseOptions parses and validates the rule options
func parseOptions(options any) ValidTypeofOptions {
	opts := ValidTypeofOptions{
		RequireStringLiterals: false,
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
		if v, ok := optsMap["requireStringLiterals"].(bool); ok {
			opts.RequireStringLiterals = v
		}
	}

	return opts
}

// Valid typeof values according to the JavaScript specification
var validTypes = map[string]bool{
	"symbol":    true,
	"undefined": true,
	"object":    true,
	"boolean":   true,
	"number":    true,
	"string":    true,
	"function":  true,
	"bigint":    true,
}

func buildInvalidValueMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "invalidValue",
		Description: "Invalid typeof comparison value.",
	}
}

func buildNotStringMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "notString",
		Description: "Typeof comparisons should be to string literals.",
	}
}

// ValidTypeofRule implements the valid-typeof rule
// Enforce comparing `typeof` expressions against valid strings
var ValidTypeofRule = rule.Rule{
	Name: "valid-typeof",
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

			// Check if this is an equality comparison
			op := binExpr.OperatorToken.Kind
			if op != ast.KindEqualsEqualsToken &&
				op != ast.KindEqualsEqualsEqualsToken &&
				op != ast.KindExclamationEqualsToken &&
				op != ast.KindExclamationEqualsEqualsToken {
				return
			}

			// Check if either side is a typeof expression
			var typeofNode *ast.Node
			var otherNode *ast.Node

			if isTypeofExpression(binExpr.Left) {
				typeofNode = binExpr.Left
				otherNode = binExpr.Right
			} else if isTypeofExpression(binExpr.Right) {
				typeofNode = binExpr.Right
				otherNode = binExpr.Left
			} else {
				return
			}

			// Check the other side of the comparison
			if isStringLiteral(otherNode) || isTemplateLiteral(otherNode) {
				// Get the string value
				value := getStringValue(otherNode)
				if value != "" && !validTypes[value] {
					ctx.ReportNode(otherNode, buildInvalidValueMessage())
				}
			} else if isUndefinedIdentifier(otherNode) {
				// Special case: typeof x === undefined
				if opts.RequireStringLiterals {
					// Suggest converting to string literal
					ctx.ReportNodeWithSuggestions(otherNode, buildNotStringMessage(),
						rule.RuleFix{
							Message: "Replace with string literal \"undefined\"",
							Edits: []rule.TextEdit{
								rule.RuleFixReplace(ctx.SourceFile, otherNode, `"undefined"`),
							},
						})
				}
			} else if !isTypeofExpression(otherNode) {
				// Not a typeof, string literal, or undefined identifier
				if opts.RequireStringLiterals {
					ctx.ReportNode(otherNode, buildNotStringMessage())
				}
			}
		},
	}
}

func isTypeofExpression(node *ast.Node) bool {
	if node == nil {
		return false
	}
	prefix := node.AsPrefixUnaryExpression()
	return prefix != nil && prefix.Operator == ast.KindTypeOfKeyword
}

func isStringLiteral(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return node.Kind == ast.KindStringLiteral
}

func isTemplateLiteral(node *ast.Node) bool {
	if node == nil {
		return false
	}
	// NoSubstitutionTemplateLiteral is a template literal without ${...}
	if node.Kind == ast.KindNoSubstitutionTemplateLiteral {
		return true
	}
	// TemplateExpression has substitutions, so it's not a valid static string
	return false
}

func isUndefinedIdentifier(node *ast.Node) bool {
	if node == nil {
		return false
	}
	ident := node.AsIdentifier()
	if ident == nil {
		return false
	}
	return utils.GetNodeText(node) == "undefined"
}

func getStringValue(node *ast.Node) string {
	if node == nil {
		return ""
	}

	if node.Kind == ast.KindStringLiteral {
		text := utils.GetNodeText(node)
		// Remove quotes
		if len(text) >= 2 {
			return text[1 : len(text)-1]
		}
		return text
	}

	if node.Kind == ast.KindNoSubstitutionTemplateLiteral {
		text := utils.GetNodeText(node)
		// Remove backticks
		if len(text) >= 2 {
			return text[1 : len(text)-1]
		}
		return text
	}

	return ""
}
