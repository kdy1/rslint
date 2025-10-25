package parameter_properties

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type ParameterPropertiesOptions struct {
	Prefer string   `json:"prefer"`
	Allow  []string `json:"allow"`
}

// Helper function to check if a modifier combination is in the allow list
func isAllowed(modifiers []string, allowList []string) bool {
	// Create a normalized version of the modifier string
	modStr := strings.Join(modifiers, " ")

	for _, allowed := range allowList {
		if modStr == allowed {
			return true
		}
	}
	return false
}

// Helper function to extract modifiers from a parameter
func getParameterModifiers(param *ast.Node) []string {
	var modifiers []string

	flags := ast.GetCombinedModifierFlags(param)

	// Add accessibility modifiers
	if flags&ast.ModifierFlagsPublic != 0 {
		modifiers = append(modifiers, "public")
	} else if flags&ast.ModifierFlagsPrivate != 0 {
		modifiers = append(modifiers, "private")
	} else if flags&ast.ModifierFlagsProtected != 0 {
		modifiers = append(modifiers, "protected")
	}

	// Add readonly modifier
	if flags&ast.ModifierFlagsReadonly != 0 {
		modifiers = append(modifiers, "readonly")
	}

	return modifiers
}

// Helper function to check if a parameter is a parameter property
func isParameterProperty(param *ast.Node) bool {
	return ast.HasSyntacticModifier(param, ast.ModifierFlagsParameterPropertyModifier)
}

// Helper function to get parameter name as string
func getParameterName(param *ast.Node) string {
	paramDecl := param.AsParameterDeclaration()
	if paramDecl == nil || paramDecl.Name() == nil {
		return ""
	}

	name := paramDecl.Name()
	if name.Kind == ast.KindIdentifier {
		return name.AsIdentifier().Text
	}

	return ""
}

// Helper function to check if a property is assigned from a constructor parameter with the same name
func isPropertyAssignedFromParameter(ctx rule.RuleContext, property *ast.Node, constructor *ast.Node) bool {
	if constructor == nil {
		return false
	}

	ctor := constructor.AsConstructorDeclaration()
	if ctor == nil || ctor.Body == nil {
		return false
	}

	propertyDecl := property.AsPropertyDeclaration()
	if propertyDecl == nil {
		return false
	}

	propertyName := ""
	if propertyDecl.Name().Kind == ast.KindIdentifier {
		propertyName = propertyDecl.Name().AsIdentifier().Text
	} else {
		return false // Don't handle computed property names
	}

	// Check if there's a parameter with the same name
	hasMatchingParameter := false
	if ctor.Parameters != nil {
		for _, param := range ctor.Parameters.Nodes {
			paramName := getParameterName(param)
			if paramName == propertyName {
				hasMatchingParameter = true
				break
			}
		}
	}

	if !hasMatchingParameter {
		return false
	}

	// Check if the property is assigned from the parameter in the constructor body
	// Look for: this.propertyName = propertyName
	body := ctor.Body.AsBlock()
	if body == nil || body.Statements == nil {
		return false
	}

	for _, stmt := range body.Statements.Nodes {
		if stmt.Kind != ast.KindExpressionStatement {
			continue
		}

		exprStmt := stmt.AsExpressionStatement()
		if exprStmt == nil || exprStmt.Expression == nil {
			continue
		}

		if exprStmt.Expression.Kind != ast.KindBinaryExpression {
			continue
		}

		binary := exprStmt.Expression.AsBinaryExpression()
		if binary == nil || binary.OperatorToken.Kind != ast.KindEqualsToken {
			continue
		}

		// Check if left side is this.propertyName
		left := binary.Left
		if !ast.IsPropertyAccessExpression(left) {
			continue
		}

		propAccess := left.AsPropertyAccessExpression()
		if propAccess == nil || propAccess.Expression.Kind != ast.KindThisKeyword {
			continue
		}

		if propAccess.Name().Kind != ast.KindIdentifier {
			continue
		}

		leftPropName := propAccess.Name().AsIdentifier().Text
		if leftPropName != propertyName {
			continue
		}

		// Check if right side is the parameter name
		right := binary.Right
		if right.Kind != ast.KindIdentifier {
			continue
		}

		rightName := right.AsIdentifier().Text
		if rightName == propertyName {
			return true
		}
	}

	return false
}

var ParameterPropertiesRule = rule.CreateRule(rule.Rule{
	Name: "parameter-properties",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := ParameterPropertiesOptions{
			Prefer: "class-property",
			Allow:  []string{},
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
				if prefer, ok := optsMap["prefer"].(string); ok {
					opts.Prefer = prefer
				}
				if allow, ok := optsMap["allow"].([]interface{}); ok {
					for _, item := range allow {
						if str, ok := item.(string); ok {
							opts.Allow = append(opts.Allow, str)
						}
					}
				}
			}
		}

		if opts.Prefer == "class-property" {
			// Report parameter properties that are not in the allow list
			return rule.RuleListeners{
				ast.KindConstructor: func(node *ast.Node) {
					constructor := node.AsConstructorDeclaration()
					if constructor == nil || constructor.Parameters == nil {
						return
					}

					for _, param := range constructor.Parameters.Nodes {
						if !isParameterProperty(param) {
							continue
						}

						modifiers := getParameterModifiers(param)
						if len(modifiers) == 0 {
							continue
						}

						// Check if this modifier combination is allowed
						if isAllowed(modifiers, opts.Allow) {
							continue
						}

						// Report the parameter
						paramDecl := param.AsParameterDeclaration()
						if paramDecl == nil || paramDecl.Name() == nil {
							continue
						}

						ctx.ReportNode(paramDecl.Name(), rule.RuleMessage{
							Id:          "preferClassProperty",
							Description: "Use class properties instead of parameter properties.",
						})
					}
				},
			}
		} else if opts.Prefer == "parameter-property" {
			// Report class properties that could be parameter properties
			// We need to track the class and its constructor
			type classInfo struct {
				constructor *ast.Node
				properties  []*ast.Node
			}

			var classStack []*classInfo

			return rule.RuleListeners{
				ast.KindClassDeclaration: func(node *ast.Node) {
					classStack = append(classStack, &classInfo{
						constructor: nil,
						properties:  []*ast.Node{},
					})
				},
				rule.ListenerOnExit(ast.KindClassDeclaration): func(node *ast.Node) {
					if len(classStack) > 0 {
						info := classStack[len(classStack)-1]
						classStack = classStack[:len(classStack)-1]

						// Check each property
						for _, property := range info.properties {
							if isPropertyAssignedFromParameter(ctx, property, info.constructor) {
								propertyDecl := property.AsPropertyDeclaration()
								if propertyDecl != nil && propertyDecl.Name() != nil {
									ctx.ReportNode(propertyDecl.Name(), rule.RuleMessage{
										Id:          "preferParameterProperty",
										Description: "Use parameter properties instead of class properties.",
									})
								}
							}
						}
					}
				},
				ast.KindClassExpression: func(node *ast.Node) {
					classStack = append(classStack, &classInfo{
						constructor: nil,
						properties:  []*ast.Node{},
					})
				},
				rule.ListenerOnExit(ast.KindClassExpression): func(node *ast.Node) {
					if len(classStack) > 0 {
						info := classStack[len(classStack)-1]
						classStack = classStack[:len(classStack)-1]

						// Check each property
						for _, property := range info.properties {
							if isPropertyAssignedFromParameter(ctx, property, info.constructor) {
								propertyDecl := property.AsPropertyDeclaration()
								if propertyDecl != nil && propertyDecl.Name() != nil {
									ctx.ReportNode(propertyDecl.Name(), rule.RuleMessage{
										Id:          "preferParameterProperty",
										Description: "Use parameter properties instead of class properties.",
									})
								}
							}
						}
					}
				},
				ast.KindConstructor: func(node *ast.Node) {
					if len(classStack) > 0 {
						classStack[len(classStack)-1].constructor = node
					}
				},
				ast.KindPropertyDeclaration: func(node *ast.Node) {
					// Only track properties that could be parameter properties
					// (i.e., have accessibility modifiers or are readonly)
					propertyDecl := node.AsPropertyDeclaration()
					if propertyDecl == nil {
						return
					}

					// Check if property has visibility modifiers or readonly
					flags := ast.GetCombinedModifierFlags(node)
					hasModifiers := (flags&ast.ModifierFlagsAccessibilityModifier != 0) ||
						(flags&ast.ModifierFlagsReadonly != 0)

					if !hasModifiers {
						return
					}

					// Don't track properties with initializers
					if propertyDecl.Initializer != nil {
						return
					}

					if len(classStack) > 0 {
						classStack[len(classStack)-1].properties = append(classStack[len(classStack)-1].properties, node)
					}
				},
			}
		}

		return rule.RuleListeners{}
	},
})
