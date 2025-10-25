package no_magic_numbers

import (
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type NoMagicNumbersOptions struct {
	Ignore                          []interface{} `json:"ignore"`
	IgnoreEnums                     bool          `json:"ignoreEnums"`
	IgnoreNumericLiteralTypes       bool          `json:"ignoreNumericLiteralTypes"`
	IgnoreReadonlyClassProperties   bool          `json:"ignoreReadonlyClassProperties"`
	IgnoreTypeIndexes               bool          `json:"ignoreTypeIndexes"`
	DetectObjects                   bool          `json:"detectObjects"`
	EnforceConst                    bool          `json:"enforceConst"`
	IgnoreArrayIndexes              bool          `json:"ignoreArrayIndexes"`
	IgnoreDefaultValues             bool          `json:"ignoreDefaultValues"`
	IgnoreClassFieldInitialValues   bool          `json:"ignoreClassFieldInitialValues"`
}

var NoMagicNumbersRule = rule.CreateRule(rule.Rule{
	Name: "no-magic-numbers",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := NoMagicNumbersOptions{
			Ignore:                        []interface{}{},
			IgnoreEnums:                   false,
			IgnoreNumericLiteralTypes:     false,
			IgnoreReadonlyClassProperties: false,
			IgnoreTypeIndexes:             false,
			DetectObjects:                 false,
			EnforceConst:                  false,
			IgnoreArrayIndexes:            false,
			IgnoreDefaultValues:           false,
			IgnoreClassFieldInitialValues: false,
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
				if detectObjects, ok := optsMap["detectObjects"].(bool); ok {
					opts.DetectObjects = detectObjects
				}
				if enforceConst, ok := optsMap["enforceConst"].(bool); ok {
					opts.EnforceConst = enforceConst
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
			}
		}

		// Check if a number is in the ignore list
		isIgnoredNumber := func(value string) bool {
			for _, ignored := range opts.Ignore {
				switch v := ignored.(type) {
				case float64:
					// Parse the value and compare
					if parsed, err := strconv.ParseFloat(value, 64); err == nil && parsed == v {
						return true
					}
				case string:
					// String comparison for bigints and exact matches
					if v == value {
						return true
					}
				case int:
					if parsed, err := strconv.Atoi(value); err == nil && parsed == v {
						return true
					}
				}
			}
			return false
		}

		// Check if we're in a const variable declaration
		isInConstDeclaration := func(node *ast.Node) bool {
			parent := node.Parent
			for parent != nil {
				if parent.Kind == ast.KindVariableDeclaration {
					varDecl := parent.AsVariableDeclaration()
					if varDecl != nil && varDecl.Parent != nil {
						if varDecl.Parent.Kind == ast.KindVariableDeclarationList {
							varDeclList := varDecl.Parent.AsVariableDeclarationList()
							if varDeclList != nil {
								return (varDeclList.Flags & ast.NodeFlagsConst) != 0
							}
						}
					}
					return false
				}
				parent = parent.Parent
			}
			return false
		}

		// Check if we're in an enum member
		isInEnumMember := func(node *ast.Node) bool {
			parent := node.Parent
			for parent != nil {
				if parent.Kind == ast.KindEnumMember {
					return true
				}
				parent = parent.Parent
			}
			return false
		}

		// Check if we're in a numeric literal type
		isInNumericLiteralType := func(node *ast.Node) bool {
			parent := node.Parent
			// Check if the parent is a literal type node
			if parent != nil && parent.Kind == ast.KindLiteralType {
				// Check if this literal type is used in a type context
				typeParent := parent.Parent
				if typeParent != nil {
					switch typeParent.Kind {
					case ast.KindTypeAliasDeclaration,
						ast.KindInterfaceDeclaration,
						ast.KindPropertySignature,
						ast.KindTypeReference,
						ast.KindUnionType,
						ast.KindIntersectionType,
						ast.KindParameter,
						ast.KindTypeParameter:
						return true
					}
				}
			}
			return false
		}

		// Check if we're in a readonly class property
		isInReadonlyClassProperty := func(node *ast.Node) bool {
			parent := node.Parent
			for parent != nil {
				if parent.Kind == ast.KindPropertyDeclaration {
					propDecl := parent.AsPropertyDeclaration()
					if propDecl != nil {
						// Check if it has readonly modifier
						if ast.HasSyntacticModifier(parent, ast.ModifierFlagsReadonly) {
							return true
						}
					}
					return false
				}
				parent = parent.Parent
			}
			return false
		}

		// Check if we're being used as a type index
		isTypeIndex := func(node *ast.Node) bool {
			parent := node.Parent
			// Check if we're inside a bracket in a type context
			// Walk up to find if we're in an indexed access type
			for parent != nil {
				if parent.Kind == ast.KindIndexedAccessType {
					return true
				}
				// Stop if we hit a non-type node
				if parent.Kind == ast.KindVariableDeclaration ||
					parent.Kind == ast.KindBinaryExpression ||
					parent.Kind == ast.KindCallExpression {
					return false
				}
				parent = parent.Parent
			}
			return false
		}

		// Check if we're being used as an array index
		isArrayIndex := func(node *ast.Node) bool {
			parent := node.Parent
			// Check if parent is an element access expression (e.g., arr[0])
			if parent != nil && parent.Kind == ast.KindElementAccessExpression {
				elemAccess := parent.AsElementAccessExpression()
				if elemAccess != nil && elemAccess.ArgumentExpression == node {
					// Validate that it's a valid array index (0 to 4294967294)
					value := strings.TrimSpace(node.Text())
					if parsed, err := strconv.ParseUint(value, 10, 32); err == nil && parsed <= 4294967294 {
						return true
					}
				}
			}
			return false
		}

		// Check if we're in a default value assignment
		isDefaultValue := func(node *ast.Node) bool {
			parent := node.Parent
			for parent != nil {
				// Check for parameter default value
				if parent.Kind == ast.KindParameter {
					param := parent.AsParameterDeclaration()
					if param != nil && param.Initializer == node {
						return true
					}
				}
				// Check for binding element default value (destructuring)
				if parent.Kind == ast.KindBindingElement {
					bindingElem := parent.AsBindingElement()
					if bindingElem != nil && bindingElem.Initializer == node {
						return true
					}
				}
				parent = parent.Parent
			}
			return false
		}

		// Check if we're in a class field initial value
		isClassFieldInitialValue := func(node *ast.Node) bool {
			parent := node.Parent
			for parent != nil {
				if parent.Kind == ast.KindPropertyDeclaration {
					propDecl := parent.AsPropertyDeclaration()
					if propDecl != nil && propDecl.Initializer == node {
						// Check if this property is in a class
						propParent := parent.Parent
						if propParent != nil {
							switch propParent.Kind {
							case ast.KindClassDeclaration, ast.KindClassExpression:
								return true
							}
						}
					}
					return false
				}
				parent = parent.Parent
			}
			return false
		}

		// Check if we should skip this numeric literal
		shouldIgnore := func(node *ast.Node) bool {
			value := strings.TrimSpace(node.Text())

			// Handle negative numbers - check if parent is a prefix unary expression with minus
			actualValue := value
			if node.Parent != nil && node.Parent.Kind == ast.KindPrefixUnaryExpression {
				prefixUnary := node.Parent.AsPrefixUnaryExpression()
				if prefixUnary != nil && prefixUnary.Operator == ast.KindMinusToken {
					actualValue = "-" + value
				}
			}

			// Check if in ignore list
			if isIgnoredNumber(actualValue) {
				return true
			}

			// Check TypeScript-specific ignores
			if opts.IgnoreEnums && isInEnumMember(node) {
				return true
			}

			if opts.IgnoreNumericLiteralTypes && isInNumericLiteralType(node) {
				return true
			}

			if opts.IgnoreReadonlyClassProperties && isInReadonlyClassProperty(node) {
				return true
			}

			if opts.IgnoreTypeIndexes && isTypeIndex(node) {
				return true
			}

			// Check base ESLint ignores
			if opts.IgnoreArrayIndexes && isArrayIndex(node) {
				return true
			}

			if opts.IgnoreDefaultValues && isDefaultValue(node) {
				return true
			}

			if opts.IgnoreClassFieldInitialValues && isClassFieldInitialValue(node) {
				return true
			}

			// Check enforceConst option
			if opts.EnforceConst {
				// If enforceConst is true, we only ignore numbers in const declarations
				// Numbers in let/var declarations should be reported
				if isInConstDeclaration(node) {
					return true
				}
				// If not in any variable declaration, fall through to other checks
			}

			// Check detectObjects option
			if !opts.DetectObjects {
				// If detectObjects is false, ignore numbers in object properties
				parent := node.Parent
				if parent != nil && parent.Kind == ast.KindPropertyAssignment {
					propAssign := parent.AsPropertyAssignment()
					if propAssign != nil && propAssign.Initializer == node {
						return true
					}
				}
			}

			return false
		}

		return rule.RuleListeners{
			ast.KindNumericLiteral: func(node *ast.Node) {
				// Skip if this number should be ignored
				if shouldIgnore(node) {
					return
				}

				// Get the numeric value for the message
				value := strings.TrimSpace(node.Text())

				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "noMagic",
					Description: "No magic number: " + value + ".",
				})
			},
		}
	},
})
