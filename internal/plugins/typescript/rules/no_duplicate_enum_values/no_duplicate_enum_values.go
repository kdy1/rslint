package no_duplicate_enum_values

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// evaluateEnumValue attempts to evaluate an enum initializer to a concrete value
// Returns (value, displayValue, isEvaluable)
func evaluateEnumValue(node *ast.Node) (string, string, bool) {
	if node == nil {
		return "", "", false
	}

	switch node.Kind {
	case ast.KindNumericLiteral:
		numLit := node.AsNumericLiteral()
		if numLit != nil {
			// Parse the numeric value to normalize it
			val, err := strconv.ParseFloat(numLit.Text, 64)
			if err == nil {
				// Normalize the value (e.g., 0x10 -> 16, 1e2 -> 100)
				// Special handling for NaN and Infinity
				if math.IsNaN(val) {
					return "NaN", "NaN", false // NaN is never equal to itself
				}
				if math.IsInf(val, 0) {
					return "Infinity", "Infinity", false // Infinity is allowed to duplicate
				}
				// Normalize representation
				normalized := fmt.Sprintf("%v", val)
				return normalized, normalized, true
			}
			return numLit.Text, numLit.Text, true
		}

	case ast.KindStringLiteral:
		strLit := node.AsStringLiteral()
		if strLit != nil {
			// Remove surrounding quotes from the text
			text := strLit.Text
			if len(text) >= 2 && (text[0] == '"' || text[0] == '\'') {
				text = text[1 : len(text)-1]
			}
			return "string:" + text, text, true
		}

	case ast.KindNoSubstitutionTemplateLiteral:
		templateLit := node.AsNoSubstitutionTemplateLiteral()
		if templateLit != nil {
			// Remove backticks from the text
			text := templateLit.Text
			if len(text) >= 2 && text[0] == '`' && text[len(text)-1] == '`' {
				text = text[1 : len(text)-1]
			}
			// Template literals are treated as strings for comparison
			return "string:" + text, text, true
		}

	case ast.KindPrefixUnaryExpression:
		unaryExpr := node.AsPrefixUnaryExpression()
		if unaryExpr != nil {
			operand := unaryExpr.Operand

			// Recursively evaluate the operand
			innerVal, innerDisplay, ok := evaluateEnumValue(operand)
			if !ok {
				return "", "", false
			}

			// Handle different unary operators
			switch unaryExpr.Operator {
			case ast.KindPlusToken:
				// Unary + converts to number
				if strings.HasPrefix(innerVal, "string:") {
					// Convert string to number
					strVal := strings.TrimPrefix(innerVal, "string:")
					numVal, err := strconv.ParseFloat(strVal, 64)
					if err == nil {
						if math.IsNaN(numVal) {
							return "NaN", "NaN", false
						}
						normalized := fmt.Sprintf("%v", numVal)
						return normalized, normalized, true
					}
				} else {
					// Already a number, just remove any leading +
					return innerVal, innerDisplay, true
				}

			case ast.KindMinusToken:
				// Unary - negates the value
				if strings.HasPrefix(innerVal, "string:") {
					// Convert string to number then negate
					strVal := strings.TrimPrefix(innerVal, "string:")
					numVal, err := strconv.ParseFloat(strVal, 64)
					if err == nil {
						negated := -numVal
						if math.IsNaN(negated) {
							return "NaN", "NaN", false
						}
						normalized := fmt.Sprintf("%v", negated)
						return normalized, normalized, true
					}
				} else {
					// Negate numeric value
					numVal, err := strconv.ParseFloat(innerVal, 64)
					if err == nil {
						negated := -numVal
						if math.IsNaN(negated) {
							return "NaN", "NaN", false
						}
						normalized := fmt.Sprintf("%v", negated)
						return normalized, normalized, true
					}
				}
			}
		}

	case ast.KindIdentifier:
		// Check for special identifiers like NaN, Infinity
		identifier := node.AsIdentifier()
		if identifier != nil {
			switch identifier.EscapedText {
			case "NaN":
				return "NaN", "NaN", false // NaN is never equal to itself
			case "Infinity":
				return "Infinity", "Infinity", false // Infinity is allowed to duplicate
			}
		}
	}

	return "", "", false
}

var NoDuplicateEnumValuesRule = rule.CreateRule(rule.Rule{
	Name: "no-duplicate-enum-values",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindEnumDeclaration: func(node *ast.Node) {
				enumDecl := node.AsEnumDeclaration()
				if enumDecl == nil || enumDecl.Members == nil {
					return
				}

				// Track seen values: map[normalizedValue]memberNode
				seenValues := make(map[string]*ast.Node)

				for _, memberNode := range enumDecl.Members.Nodes {
					member := memberNode.AsEnumMember()
					if member == nil || member.Initializer == nil {
						continue
					}

					// Evaluate the initializer
					valueKey, displayValue, isEvaluable := evaluateEnumValue(member.Initializer)
					if !isEvaluable {
						continue
					}

					// Check for duplicate
					if _, exists := seenValues[valueKey]; exists {
						ctx.ReportNode(member.Name(), rule.RuleMessage{
							Id:          "duplicateValue",
							Description: fmt.Sprintf("Duplicate enum member value %s.", displayValue),
						})
					} else {
						seenValues[valueKey] = memberNode
					}
				}
			},
		}
	},
})
