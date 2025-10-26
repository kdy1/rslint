package no_dupe_keys

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoDupeKeysRule implements the no-dupe-keys rule
// Disallow duplicate keys in object literals
var NoDupeKeysRule = rule.Rule{
	Name: "no-dupe-keys",
	Run:  run,
}

// getPropertyKeyName extracts the normalized key name from a property
func getPropertyKeyName(prop *ast.Node) (string, bool) {
	if prop == nil {
		return "", false
	}

	name := prop.Name()
	if name == nil {
		return "", false
	}

	// Handle computed property names - we can't statically analyze them
	// unless they're literals
	if prop.Kind == ast.KindPropertyAssignment ||
	   prop.Kind == ast.KindShorthandPropertyAssignment ||
	   prop.Kind == ast.KindMethodDeclaration ||
	   prop.Kind == ast.KindGetAccessor ||
	   prop.Kind == ast.KindSetAccessor {

		// Check for computed property
		// In TypeScript AST, computed properties are wrapped in ComputedPropertyName
		if name.Kind == ast.KindComputedPropertyName {
			// Get the expression inside the computed property
			expr := name.Expression()
			if expr == nil {
				return "", false // Can't determine the key statically
			}

			// Only handle literal computed properties
			switch expr.Kind {
			case ast.KindStringLiteral:
				text := expr.Text()
				if len(text) >= 2 {
					return text[1 : len(text)-1], true // Remove quotes
				}
				return "", false
			case ast.KindNumericLiteral:
				// Normalize numeric literals
				numText := expr.Text()
				return normalizeNumericKey(numText), true
			case ast.KindNoSubstitutionTemplateLiteral:
				text := expr.Text()
				if len(text) >= 2 {
					return text[1 : len(text)-1], true // Remove backticks
				}
				return "", false
			default:
				// Can't statically determine the key
				return "", false
			}
		}
	}

	// Handle regular identifiers
	if name.Kind == ast.KindIdentifier {
		return name.Text(), true
	}

	// Handle string literals
	if name.Kind == ast.KindStringLiteral {
		text := name.Text()
		if len(text) >= 2 {
			return text[1 : len(text)-1], true // Remove quotes
		}
		return "", false
	}

	// Handle numeric literals
	if name.Kind == ast.KindNumericLiteral {
		numText := name.Text()
		return normalizeNumericKey(numText), true
	}

	return "", false
}

// normalizeNumericKey normalizes different numeric representations to their decimal form
func normalizeNumericKey(numText string) string {
	// Remove underscores from numeric separators
	numText = strings.ReplaceAll(numText, "_", "")

	var num int64

	// Handle different numeric bases
	if strings.HasPrefix(numText, "0x") || strings.HasPrefix(numText, "0X") {
		// Hexadecimal
		num, _ = strconv.ParseInt(numText[2:], 16, 64)
	} else if strings.HasPrefix(numText, "0o") || strings.HasPrefix(numText, "0O") {
		// Octal (ES6+)
		num, _ = strconv.ParseInt(numText[2:], 8, 64)
	} else if strings.HasPrefix(numText, "0b") || strings.HasPrefix(numText, "0B") {
		// Binary
		num, _ = strconv.ParseInt(numText[2:], 2, 64)
	} else if strings.HasPrefix(numText, "0") && len(numText) > 1 && numText[1] >= '0' && numText[1] <= '9' {
		// Legacy octal (0123)
		num, _ = strconv.ParseInt(numText, 8, 64)
	} else if strings.HasSuffix(numText, "n") {
		// BigInt - remove 'n' suffix and parse
		num, _ = strconv.ParseInt(numText[:len(numText)-1], 10, 64)
	} else {
		// Decimal
		num, _ = strconv.ParseInt(numText, 10, 64)
	}

	return fmt.Sprintf("%d", num)
}

// checkObjectLiteral checks for duplicate keys in an object literal
func checkObjectLiteral(ctx rule.RuleContext, node *ast.Node) {
	if node == nil {
		return
	}

	props := node.Properties()
	if props == nil || len(props) == 0 {
		return
	}

	// Track keys we've seen
	// Special handling for __proto__: only literal __proto__ sets the prototype
	seenKeys := make(map[string]*ast.Node)
	seenProtoLiteral := false

	for _, prop := range props {
		if prop == nil {
			continue
		}

		// Skip spread elements
		if prop.Kind == ast.KindSpreadAssignment {
			continue
		}

		keyName, ok := getPropertyKeyName(prop)
		if !ok {
			// Can't determine the key (e.g., computed property with variable)
			continue
		}

		// Special handling for __proto__
		// Literal __proto__ (not computed) sets the prototype
		// Computed ['__proto__'] creates a regular property
		if keyName == "__proto__" {
			name := prop.Name()
			isComputed := name != nil && name.Kind == ast.KindComputedPropertyName

			if !isComputed {
				// This is a literal __proto__
				if seenProtoLiteral {
					// Duplicate literal __proto__
					ctx.ReportNode(prop, rule.RuleMessage{
						Id:          "unexpected",
						Description: "Duplicate key '__proto__'.",
						Data: map[string]interface{}{
							"name": "__proto__",
						},
					})
				}
				seenProtoLiteral = true
				continue
			}
			// Computed __proto__ is treated as a regular key, fall through
		}

		// Check for getter/setter pairs - these are allowed
		isGetter := prop.Kind == ast.KindGetAccessor
		isSetter := prop.Kind == ast.KindSetAccessor

		if firstProp, exists := seenKeys[keyName]; exists {
			// Check if it's a getter/setter pair
			firstIsGetter := firstProp.Kind == ast.KindGetAccessor
			firstIsSetter := firstProp.Kind == ast.KindSetAccessor

			if (isGetter && firstIsSetter) || (isSetter && firstIsGetter) {
				// This is a valid getter/setter pair
				continue
			}

			// Duplicate key found
			ctx.ReportNode(prop, rule.RuleMessage{
				Id:          "unexpected",
				Description: "Duplicate key '" + keyName + "'.",
				Data: map[string]interface{}{
					"name": keyName,
				},
			})
		} else {
			seenKeys[keyName] = prop
		}
	}
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	return rule.RuleListeners{
		ast.KindObjectLiteralExpression: func(node *ast.Node) {
			checkObjectLiteral(ctx, node)
		},
	}
}
