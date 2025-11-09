package no_magic_numbers

import (
	"fmt"
	"strconv"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type NoMagicNumbersOptions struct {
	Ignore                         []interface{} `json:"ignore"`
	IgnoreArrayIndexes             bool          `json:"ignoreArrayIndexes"`
	IgnoreDefaultValues            bool          `json:"ignoreDefaultValues"`
	IgnoreClassFieldInitialValues  bool          `json:"ignoreClassFieldInitialValues"`
	EnforceConst                   bool          `json:"enforceConst"`
	DetectObjects                  bool          `json:"detectObjects"`
	IgnoreEnums                    bool          `json:"ignoreEnums"`
	IgnoreNumericLiteralTypes      bool          `json:"ignoreNumericLiteralTypes"`
	IgnoreReadonlyClassProperties  bool          `json:"ignoreReadonlyClassProperties"`
	IgnoreTypeIndexes              bool          `json:"ignoreTypeIndexes"`
}

var NoMagicNumbersRule = rule.CreateRule(rule.Rule{
	Name: "no-magic-numbers",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := NoMagicNumbersOptions{
			Ignore:                        []interface{}{},
			IgnoreArrayIndexes:            false,
			IgnoreDefaultValues:           false,
			IgnoreClassFieldInitialValues: false,
			EnforceConst:                  false,
			DetectObjects:                 false,
			IgnoreEnums:                   false,
			IgnoreNumericLiteralTypes:     false,
			IgnoreReadonlyClassProperties: false,
			IgnoreTypeIndexes:             false,
		}

		// Parse options with dual-format support (handles both array and object formats)
		if options != nil {
			var optsMap map[string]interface{}
			var ok bool

			// Handle array format: [{ option: value }]
			if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
				optsMap, ok = optArray[0].(map[string]interface{})
			} else {
				// Handle direct object format: { option: value }
				optsMap, ok = options.(map[string]interface{})
			}

			if ok {
				if ignore, ok := optsMap["ignore"].([]interface{}); ok {
					opts.Ignore = ignore
				}
				if ignoreArrayIndexes, ok := optsMap["ignoreArrayIndexes"].(bool); ok {
					opts.IgnoreArrayIndexes = ignoreArrayIndexes
				}
				if ignoreDefaultValues, ok := optsMap["ignoreDefaultValues"].(bool); ok {
					opts.IgnoreDefaultValues = ignoreDefaultValues
				}
				if ignoreClassFieldInitialValues, ok := optsMap["ignoreClassFieldInitialValues"].(bool); ok {
					opts.IgnoreClassFieldInitialValues = ignoreClassFieldInitialValues
				}
				if enforceConst, ok := optsMap["enforceConst"].(bool); ok {
					opts.EnforceConst = enforceConst
				}
				if detectObjects, ok := optsMap["detectObjects"].(bool); ok {
					opts.DetectObjects = detectObjects
				}
				if ignoreEnums, ok := optsMap["ignoreEnums"].(bool); ok {
					opts.IgnoreEnums = ignoreEnums
				}
				if ignoreNumericLiteralTypes, ok := optsMap["ignoreNumericLiteralTypes"].(bool); ok {
					opts.IgnoreNumericLiteralTypes = ignoreNumericLiteralTypes
				}
				if ignoreReadonlyClassProperties, ok := optsMap["ignoreReadonlyClassProperties"].(bool); ok {
					opts.IgnoreReadonlyClassProperties = ignoreReadonlyClassProperties
				}
				if ignoreTypeIndexes, ok := optsMap["ignoreTypeIndexes"].(bool); ok {
					opts.IgnoreTypeIndexes = ignoreTypeIndexes
				}
			}
		}

		// Helper function to get the raw text of a numeric literal
		getRawText := func(node *ast.Node) string {
			nodeRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
			return ctx.SourceFile.Text()[nodeRange.Pos():nodeRange.End()]
		}

		// Helper function to check if a numeric value should be ignored
		shouldIgnore := func(node *ast.Node) bool {
			raw := getRawText(node)

			for _, ignoreValue := range opts.Ignore {
				switch v := ignoreValue.(type) {
				case string:
					if raw == v {
						return true
					}
				case float64:
					// Handle numeric comparison
					if numLit := node.AsNumericLiteral(); numLit != nil {
						if numLit.Text == fmt.Sprintf("%v", v) || raw == fmt.Sprintf("%v", v) {
							return true
						}
						// Also try parsing as float
						if parsed, err := strconv.ParseFloat(numLit.Text, 64); err == nil && parsed == v {
							return true
						}
					}
					// Handle negative numbers
					if node.Kind == ast.KindPrefixUnaryExpression {
						prefixUnary := node.AsPrefixUnaryExpression()
						if prefixUnary != nil && prefixUnary.Operator == ast.KindMinusToken {
							if numLit := prefixUnary.Operand.AsNumericLiteral(); numLit != nil {
								if parsed, err := strconv.ParseFloat(numLit.Text, 64); err == nil && -parsed == v {
									return true
								}
							}
						}
					}
				case int:
					// Handle integer comparison
					floatVal := float64(v)
					if numLit := node.AsNumericLiteral(); numLit != nil {
						if parsed, err := strconv.ParseFloat(numLit.Text, 64); err == nil && parsed == floatVal {
							return true
						}
					}
				}
			}
			return false
		}

		// Helper to check if node is in a readonly class property
		isInReadonlyClassProperty := func(node *ast.Node) bool {
			if !opts.IgnoreReadonlyClassProperties {
				return false
			}

			// Traverse up to find PropertyDeclaration
			parent := node.Parent
			for parent != nil {
				if parent.Kind == ast.KindPropertyDeclaration {
					propDecl := parent.AsPropertyDeclaration()
					if propDecl != nil && propDecl.Modifiers() != nil {
						for _, mod := range propDecl.Modifiers().Nodes {
							if mod.Kind == ast.KindReadonlyKeyword {
								return true
							}
						}
					}
					break
				}
				parent = parent.Parent
			}
			return false
		}

		// Helper to check if node is in an enum
		isInEnum := func(node *ast.Node) bool {
			if !opts.IgnoreEnums {
				return false
			}

			parent := node.Parent
			for parent != nil {
				if parent.Kind == ast.KindEnumMember || parent.Kind == ast.KindEnumDeclaration {
					return true
				}
				parent = parent.Parent
			}
			return false
		}

		// Helper to check if node is a numeric literal type
		isNumericLiteralType := func(node *ast.Node) bool {
			if !opts.IgnoreNumericLiteralTypes {
				return false
			}

			parent := node.Parent
			if parent == nil {
				return false
			}

			// Check if parent is a type context
			switch parent.Kind {
			case ast.KindLiteralType:
				return true
			case ast.KindPrefixUnaryExpression:
				// Negative number in type context
				if parent.Parent != nil && parent.Parent.Kind == ast.KindLiteralType {
					return true
				}
			}
			return false
		}

		// Helper to check if node is in a type index
		isInTypeIndex := func(node *ast.Node) bool {
			if !opts.IgnoreTypeIndexes {
				return false
			}

			parent := node.Parent
			for parent != nil {
				if parent.Kind == ast.KindIndexedAccessType {
					// We cannot directly access IndexType field, so we skip this check for now
					// This is a limitation of the current typescript-go AST API
					return true
				}
				// Also check for mapped type keys like [K in 0 | 1 | 2]
				if parent.Kind == ast.KindMappedType {
					return false // Numbers in mapped type keys are NOT index types
				}
				parent = parent.Parent
			}
			return false
		}

		// Helper to report a magic number
		reportMagicNumber := func(node *ast.Node) {
			raw := getRawText(node)

			// Check if this number should be ignored
			if shouldIgnore(node) {
				return
			}

			// Check TypeScript-specific contexts
			if isInEnum(node) {
				return
			}

			if isInReadonlyClassProperty(node) {
				return
			}

			if isNumericLiteralType(node) {
				return
			}

			if isInTypeIndex(node) {
				return
			}

			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "noMagic",
				Description: fmt.Sprintf("No magic number: %s.", raw),
			})
		}

		return rule.RuleListeners{
			ast.KindNumericLiteral: func(node *ast.Node) {
				numLit := node.AsNumericLiteral()
				if numLit == nil {
					return
				}

				reportMagicNumber(node)
			},
			ast.KindBigIntLiteral: func(node *ast.Node) {
				bigIntLit := node.AsBigIntLiteral()
				if bigIntLit == nil {
					return
				}

				reportMagicNumber(node)
			},
			ast.KindPrefixUnaryExpression: func(node *ast.Node) {
				prefixUnary := node.AsPrefixUnaryExpression()
				if prefixUnary == nil {
					return
				}

				// Handle negative numbers (unary minus)
				if prefixUnary.Operator == ast.KindMinusToken || prefixUnary.Operator == ast.KindPlusToken {
					operand := prefixUnary.Operand
					if operand.Kind == ast.KindNumericLiteral || operand.Kind == ast.KindBigIntLiteral {
						reportMagicNumber(node)
					}
				}
			},
		}
	},
})
