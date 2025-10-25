package no_dupe_keys

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// NoDupeKeysRule implements the no-dupe-keys rule
// Disallow duplicate keys in object literals
var NoDupeKeysRule = rule.Rule{
	Name: "no-dupe-keys",
	Run:  run,
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	return rule.RuleListeners{
		ast.KindObjectLiteralExpression: func(node *ast.Node) {
			objLiteral := node.AsObjectLiteralExpression()
			if objLiteral == nil || objLiteral.Properties == nil {
				return
			}

			// Track keys we've seen. Map from normalized key to first property node
			seenKeys := make(map[string]*ast.Node)
			// Track getter and setter separately
			seenGetters := make(map[string]*ast.Node)
			seenSetters := make(map[string]*ast.Node)

			for _, prop := range objLiteral.Properties.Nodes {
				if prop == nil {
					continue
				}

				var keyNode *ast.Node
				var keyName string
				var isGetter, isSetter bool

				switch prop.Kind {
				case ast.KindPropertyAssignment:
					propAssign := prop.AsPropertyAssignment()
					if propAssign == nil || propAssign.Name() == nil {
						continue
					}
					keyNode = propAssign.Name()
					keyName = getKeyName(ctx.SourceFile, keyNode)

				case ast.KindShorthandPropertyAssignment:
					// Shorthand properties don't have duplicate issues
					continue

				case ast.KindSpreadAssignment:
					// Spread assignments don't have keys
					continue

				case ast.KindMethodDeclaration:
					method := prop.AsMethodDeclaration()
					if method == nil || method.Name() == nil {
						continue
					}
					keyNode = method.Name()
					keyName = getKeyName(ctx.SourceFile, keyNode)

				case ast.KindGetAccessor:
					accessor := prop.AsGetAccessorDeclaration()
					if accessor == nil || accessor.Name() == nil {
						continue
					}
					keyNode = accessor.Name()
					keyName = getKeyName(ctx.SourceFile, keyNode)
					isGetter = true

				case ast.KindSetAccessor:
					accessor := prop.AsSetAccessorDeclaration()
					if accessor == nil || accessor.Name() == nil {
						continue
					}
					keyNode = accessor.Name()
					keyName = getKeyName(ctx.SourceFile, keyNode)
					isSetter = true

				default:
					continue
				}

				// Only skip if we truly can't determine the key
				// Empty string "" is a valid key and should be checked
				// We return "" from getKeyName only when we can't statically determine the key
				// (e.g., for computed properties with non-literal expressions)
				if keyName == "" && keyNode.Kind != ast.KindStringLiteral {
					continue
				}

				// Special handling for __proto__
				// __proto__: value is different from ["__proto__"]: value
				// The former sets the prototype, the latter sets a property
				isProtoLiteral := keyName == "__proto__" && keyNode.Kind == ast.KindIdentifier

				if isProtoLiteral {
					// Only check against other __proto__ literal keys
					if firstProp, exists := seenKeys["__proto__"]; exists {
						ctx.ReportNode(keyNode, rule.RuleMessage{
							Id:          "unexpected",
							Description: fmt.Sprintf("Duplicate key '%s'.", keyName),
						})
						_ = firstProp
					} else {
						seenKeys["__proto__"] = keyNode
					}
					continue
				}

				// For getters and setters, they can coexist with same name
				if isGetter {
					// Check against regular properties and other getters
					if firstProp, exists := seenKeys[keyName]; exists {
						ctx.ReportNode(keyNode, rule.RuleMessage{
							Id:          "unexpected",
							Description: fmt.Sprintf("Duplicate key '%s'.", keyName),
						})
						_ = firstProp
					} else if firstGetter, exists := seenGetters[keyName]; exists {
						ctx.ReportNode(keyNode, rule.RuleMessage{
							Id:          "unexpected",
							Description: fmt.Sprintf("Duplicate key '%s'.", keyName),
						})
						_ = firstGetter
					} else {
						seenGetters[keyName] = keyNode
					}
				} else if isSetter {
					// Check against regular properties and other setters
					if firstProp, exists := seenKeys[keyName]; exists {
						ctx.ReportNode(keyNode, rule.RuleMessage{
							Id:          "unexpected",
							Description: fmt.Sprintf("Duplicate key '%s'.", keyName),
						})
						_ = firstProp
					} else if firstSetter, exists := seenSetters[keyName]; exists {
						ctx.ReportNode(keyNode, rule.RuleMessage{
							Id:          "unexpected",
							Description: fmt.Sprintf("Duplicate key '%s'.", keyName),
						})
						_ = firstSetter
					} else {
						seenSetters[keyName] = keyNode
					}
				} else {
					// Regular property - check against everything
					if firstProp, exists := seenKeys[keyName]; exists {
						ctx.ReportNode(keyNode, rule.RuleMessage{
							Id:          "unexpected",
							Description: fmt.Sprintf("Duplicate key '%s'.", keyName),
						})
						_ = firstProp
					} else if firstGetter, exists := seenGetters[keyName]; exists {
						ctx.ReportNode(keyNode, rule.RuleMessage{
							Id:          "unexpected",
							Description: fmt.Sprintf("Duplicate key '%s'.", keyName),
						})
						_ = firstGetter
					} else if firstSetter, exists := seenSetters[keyName]; exists {
						ctx.ReportNode(keyNode, rule.RuleMessage{
							Id:          "unexpected",
							Description: fmt.Sprintf("Duplicate key '%s'.", keyName),
						})
						_ = firstSetter
					} else {
						seenKeys[keyName] = keyNode
					}
				}
			}
		},
	}
}

// getKeyName extracts and normalizes the key name from a property name node
func getKeyName(sourceFile *ast.SourceFile, keyNode *ast.Node) string {
	if keyNode == nil {
		return ""
	}

	switch keyNode.Kind {
	case ast.KindIdentifier:
		return keyNode.AsIdentifier().Text

	case ast.KindStringLiteral:
		// Get the string value without quotes
		textRange := utils.TrimNodeTextRange(sourceFile, keyNode)
		text := sourceFile.Text()[textRange.Pos():textRange.End()]
		if len(text) >= 2 {
			return text[1 : len(text)-1] // Remove quotes
		}
		return text

	case ast.KindNoSubstitutionTemplateLiteral:
		// Get template string value without backticks
		textRange := utils.TrimNodeTextRange(sourceFile, keyNode)
		text := sourceFile.Text()[textRange.Pos():textRange.End()]
		if len(text) >= 2 {
			return text[1 : len(text)-1] // Remove backticks
		}
		return text

	case ast.KindNumericLiteral:
		// Normalize numeric literals
		textRange := utils.TrimNodeTextRange(sourceFile, keyNode)
		text := sourceFile.Text()[textRange.Pos():textRange.End()]
		return normalizeNumber(text)

	case ast.KindBigIntLiteral:
		// Normalize BigInt literals (e.g., 1n -> "1")
		textRange := utils.TrimNodeTextRange(sourceFile, keyNode)
		text := sourceFile.Text()[textRange.Pos():textRange.End()]
		return normalizeNumber(text)

	case ast.KindComputedPropertyName:
		// For computed property names, we can only check literal values
		computed := keyNode.AsComputedPropertyName()
		if computed != nil && computed.Expression != nil {
			expr := computed.Expression
			// Only process literal values in computed properties
			if expr.Kind == ast.KindStringLiteral ||
				expr.Kind == ast.KindNumericLiteral ||
				expr.Kind == ast.KindBigIntLiteral ||
				expr.Kind == ast.KindNoSubstitutionTemplateLiteral {
				return getKeyName(sourceFile, expr)
			}
		}
		// Can't statically determine the key for non-literal computed properties
		return ""

	default:
		return ""
	}
}

// normalizeNumber converts different numeric representations to the same string
// e.g., 0x1, 0b1, 0o1, 1 all normalize to "1"
func normalizeNumber(numStr string) string {
	numStr = strings.ReplaceAll(numStr, "_", "") // Remove numeric separators

	// Try to parse as int64 first
	if strings.HasPrefix(numStr, "0x") || strings.HasPrefix(numStr, "0X") {
		// Hexadecimal
		if val, err := strconv.ParseInt(numStr[2:], 16, 64); err == nil {
			return strconv.FormatInt(val, 10)
		}
	} else if strings.HasPrefix(numStr, "0b") || strings.HasPrefix(numStr, "0B") {
		// Binary
		if val, err := strconv.ParseInt(numStr[2:], 2, 64); err == nil {
			return strconv.FormatInt(val, 10)
		}
	} else if strings.HasPrefix(numStr, "0o") || strings.HasPrefix(numStr, "0O") {
		// Octal (ES6+ syntax)
		if val, err := strconv.ParseInt(numStr[2:], 8, 64); err == nil {
			return strconv.FormatInt(val, 10)
		}
	} else if len(numStr) > 1 && numStr[0] == '0' && numStr[1] >= '0' && numStr[1] <= '9' {
		// Legacy octal (0123)
		if val, err := strconv.ParseInt(numStr, 8, 64); err == nil {
			return strconv.FormatInt(val, 10)
		}
	} else if strings.HasSuffix(numStr, "n") {
		// BigInt
		if val, err := strconv.ParseInt(numStr[:len(numStr)-1], 10, 64); err == nil {
			return strconv.FormatInt(val, 10)
		}
	}

	// Try parsing as float
	if val, err := strconv.ParseFloat(numStr, 64); err == nil {
		// Check if it's actually an integer
		if val == float64(int64(val)) {
			return strconv.FormatInt(int64(val), 10)
		}
		return strconv.FormatFloat(val, 'f', -1, 64)
	}

	return numStr
}
