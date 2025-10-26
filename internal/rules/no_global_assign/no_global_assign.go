package no_global_assign

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoGlobalAssignOptions defines the configuration options for this rule
type NoGlobalAssignOptions struct {
	Exceptions []string `json:"exceptions"`
}

// parseOptions parses and validates the rule options
func parseOptions(options any) NoGlobalAssignOptions {
	opts := NoGlobalAssignOptions{
		Exceptions: []string{},
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
		if v, ok := optsMap["exceptions"].([]interface{}); ok {
			for _, exception := range v {
				if str, ok := exception.(string); ok {
					opts.Exceptions = append(opts.Exceptions, str)
				}
			}
		}
	}

	return opts
}

func buildGlobalShouldNotBeModifiedMessage(globalName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "globalShouldNotBeModified",
		Description: "Read-only global '" + globalName + "' should not be modified.",
	}
}

// Common JavaScript built-in global objects that should not be modified
var readOnlyGlobals = map[string]bool{
	// Value properties
	"undefined": true,
	"NaN":       true,
	"Infinity":  true,
	// Function properties
	"eval":               true,
	"isFinite":           true,
	"isNaN":              true,
	"parseFloat":         true,
	"parseInt":           true,
	"decodeURI":          true,
	"decodeURIComponent": true,
	"encodeURI":          true,
	"encodeURIComponent": true,
	// Fundamental objects
	"Object":   true,
	"Function": true,
	"Boolean":  true,
	"Symbol":   true,
	// Numbers and dates
	"Number": true,
	"BigInt": true,
	"Math":   true,
	"Date":   true,
	// Text processing
	"String": true,
	"RegExp": true,
	// Indexed collections
	"Array":             true,
	"Int8Array":         true,
	"Uint8Array":        true,
	"Uint8ClampedArray": true,
	"Int16Array":        true,
	"Uint16Array":       true,
	"Int32Array":        true,
	"Uint32Array":       true,
	"Float32Array":      true,
	"Float64Array":      true,
	"BigInt64Array":     true,
	"BigUint64Array":    true,
	// Keyed collections
	"Map":     true,
	"Set":     true,
	"WeakMap": true,
	"WeakSet": true,
	// Structured data
	"ArrayBuffer":       true,
	"SharedArrayBuffer": true,
	"Atomics":           true,
	"DataView":          true,
	"JSON":              true,
	// Control abstraction
	"Promise":           true,
	"Generator":         true,
	"GeneratorFunction": true,
	"AsyncFunction":     true,
	// Reflection
	"Reflect": true,
	"Proxy":   true,
	// Errors
	"Error":          true,
	"EvalError":      true,
	"RangeError":     true,
	"ReferenceError": true,
	"SyntaxError":    true,
	"TypeError":      true,
	"URIError":       true,
	// Other
	"Intl":        true,
	"WebAssembly": true,
}

// NoGlobalAssignRule implements the no-global-assign rule
// Disallow assignments to native objects or read-only global variables
var NoGlobalAssignRule = rule.CreateRule(rule.Rule{
	Name: "no-global-assign",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		// Create a map of exceptions for quick lookup
		exceptionMap := make(map[string]bool)
		for _, exception := range opts.Exceptions {
			exceptionMap[exception] = true
		}

		// Helper to check if an identifier is a read-only global
		isReadOnlyGlobal := func(name string) bool {
			if exceptionMap[name] {
				return false
			}
			return readOnlyGlobals[name]
		}

		// Helper to check if a node is being assigned to
		checkAssignment := func(node *ast.Node) {
			if node == nil || node.Kind != ast.KindIdentifier {
				return
			}

			name := node.Text()
			if isReadOnlyGlobal(name) {
				ctx.ReportNode(node, buildGlobalShouldNotBeModifiedMessage(name))
			}
		}

		return rule.RuleListeners{
			// Handle binary assignments like: String = "foo"
			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				if binary == nil {
					return
				}

				// Check if it's an assignment operator
				operator := binary.OperatorToken
				if operator == nil {
					return
				}

				// Check for assignment operators (=, +=, -=, etc.)
				if operator.Kind == ast.KindEqualsToken ||
					operator.Kind == ast.KindPlusEqualsToken ||
					operator.Kind == ast.KindMinusEqualsToken ||
					operator.Kind == ast.KindAsteriskEqualsToken ||
					operator.Kind == ast.KindSlashEqualsToken ||
					operator.Kind == ast.KindPercentEqualsToken ||
					operator.Kind == ast.KindAsteriskAsteriskEqualsToken ||
					operator.Kind == ast.KindLessThanLessThanEqualsToken ||
					operator.Kind == ast.KindGreaterThanGreaterThanEqualsToken ||
					operator.Kind == ast.KindGreaterThanGreaterThanGreaterThanEqualsToken ||
					operator.Kind == ast.KindAmpersandEqualsToken ||
					operator.Kind == ast.KindBarEqualsToken ||
					operator.Kind == ast.KindCaretEqualsToken ||
					operator.Kind == ast.KindBarBarEqualsToken ||
					operator.Kind == ast.KindAmpersandAmpersandEqualsToken ||
					operator.Kind == ast.KindQuestionQuestionEqualsToken {
					checkAssignment(binary.Left)
				}
			},
			// Handle postfix increments/decrements like: String++
			ast.KindPostfixUnaryExpression: func(node *ast.Node) {
				postfix := node.AsPostfixUnaryExpression()
				if postfix == nil {
					return
				}

				// Check if it's ++ or --
				if postfix.Operator == ast.KindPlusPlusToken ||
					postfix.Operator == ast.KindMinusMinusToken {
					checkAssignment(postfix.Operand)
				}
			},
			// Handle prefix increments/decrements like: ++String
			ast.KindPrefixUnaryExpression: func(node *ast.Node) {
				prefix := node.AsPrefixUnaryExpression()
				if prefix == nil {
					return
				}

				// Check if it's ++ or --
				if prefix.Operator == ast.KindPlusPlusToken ||
					prefix.Operator == ast.KindMinusMinusToken {
					checkAssignment(prefix.Operand)
				}
			},
		}
	},
})
