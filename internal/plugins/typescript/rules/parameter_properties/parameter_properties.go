package parameter_properties

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type ParameterPropertiesOptions struct {
	Allow  []string `json:"allow"`
	Prefer string   `json:"prefer"` // "class-property" or "parameter-property"
}

func buildPreferClassPropertyMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferClassProperty",
		Description: "Property {{parameter}} should be declared as a class property.",
	}
}

func buildPreferParameterPropertyMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferParameterProperty",
		Description: "Property {{parameter}} should be declared as a parameter property.",
	}
}

func parseOptions(options any) ParameterPropertiesOptions {
	opts := ParameterPropertiesOptions{
		Allow:  []string{},
		Prefer: "class-property",
	}
	if options == nil {
		return opts
	}
	// Handle array format: [{ option: value }]
	if arr, ok := options.([]interface{}); ok {
		if len(arr) > 0 {
			if m, ok := arr[0].(map[string]interface{}); ok {
				if v, ok := m["allow"].([]interface{}); ok {
					for _, item := range v {
						if s, ok := item.(string); ok {
							opts.Allow = append(opts.Allow, s)
						}
					}
				}
				if v, ok := m["prefer"].(string); ok {
					opts.Prefer = v
				}
			}
		}
		return opts
	}
	// Handle direct object format
	if m, ok := options.(map[string]interface{}); ok {
		if v, ok := m["allow"].([]interface{}); ok {
			for _, item := range v {
				if s, ok := item.(string); ok {
					opts.Allow = append(opts.Allow, s)
				}
			}
		}
		if v, ok := m["prefer"].(string); ok {
			opts.Prefer = v
		}
	}
	return opts
}

// getModifierString returns a string representation of the modifiers
func getModifierString(param *ast.ParameterDeclaration) string {
	hasPrivate := false
	hasProtected := false
	hasPublic := false
	hasReadonly := false

	if param.Modifiers != nil {
		for _, mod := range param.Modifiers.Elements {
			switch mod.Kind {
			case ast.KindPrivateKeyword:
				hasPrivate = true
			case ast.KindProtectedKeyword:
				hasProtected = true
			case ast.KindPublicKeyword:
				hasPublic = true
			case ast.KindReadonlyKeyword:
				hasReadonly = true
			}
		}
	}

	// Build the modifier string
	result := ""
	if hasPrivate {
		result = "private"
	} else if hasProtected {
		result = "protected"
	} else if hasPublic {
		result = "public"
	} else if hasReadonly {
		return "readonly"
	}

	if hasReadonly && result != "" {
		result = result + " readonly"
	}

	return result
}

// isParameterProperty checks if a parameter is a parameter property
func isParameterProperty(param *ast.ParameterDeclaration) bool {
	if param.Modifiers == nil {
		return false
	}

	for _, mod := range param.Modifiers.Elements {
		if mod.Kind == ast.KindPrivateKeyword ||
			mod.Kind == ast.KindProtectedKeyword ||
			mod.Kind == ast.KindPublicKeyword ||
			mod.Kind == ast.KindReadonlyKeyword {
			return true
		}
	}

	return false
}

// isAllowed checks if the modifier combination is in the allow list
func isAllowed(modifiers string, allowList []string) bool {
	for _, allowed := range allowList {
		if allowed == modifiers {
			return true
		}
	}
	return false
}

// getParameterName extracts the parameter name
func getParameterName(param *ast.ParameterDeclaration) string {
	if param.Name != nil && ast.IsIdentifier(param.Name) {
		ident := param.Name.AsIdentifier()
		if ident != nil {
			return ident.Text
		}
	}
	return ""
}

// checkClassPropertyPattern checks if we should prefer parameter properties
func checkClassPropertyPattern(ctx rule.RuleContext, classDecl *ast.ClassDeclaration, opts ParameterPropertiesOptions) {
	if classDecl.Members == nil {
		return
	}

	// Find the constructor
	var constructor *ast.ConstructorDeclaration
	for _, member := range classDecl.Members.Elements {
		if member.Kind == ast.KindConstructor {
			constructor = member.AsConstructorDeclaration()
			break
		}
	}

	if constructor == nil || constructor.Body == nil || constructor.Parameters == nil {
		return
	}

	// For each class member, check if it has a corresponding constructor assignment
	for _, member := range classDecl.Members.Elements {
		if member.Kind != ast.KindPropertyDeclaration {
			continue
		}

		propDecl := member.AsPropertyDeclaration()
		if propDecl == nil || propDecl.Name == nil || !ast.IsIdentifier(propDecl.Name) {
			continue
		}

		propIdent := propDecl.Name.AsIdentifier()
		if propIdent == nil {
			continue
		}
		propName := propIdent.Text

		// Check if property has initializer
		if propDecl.Initializer != nil {
			continue
		}

		// Find matching parameter
		var matchingParam *ast.ParameterDeclaration
		for _, param := range constructor.Parameters.Elements {
			if param.Kind != ast.KindParameter {
				continue
			}
			paramDecl := param.AsParameterDeclaration()
			paramName := getParameterName(paramDecl)
			if paramName == propName {
				matchingParam = paramDecl
				break
			}
		}

		if matchingParam == nil {
			continue
		}

		// Check if there's an assignment in the constructor body
		hasAssignment := false
		isFirstStatement := false
		if constructor.Body.Statements != nil && len(constructor.Body.Statements.Elements) > 0 {
			firstStmt := constructor.Body.Statements.Elements[0]
			if firstStmt.Kind == ast.KindExpressionStatement {
				exprStmt := firstStmt.AsExpressionStatement()
				if exprStmt != nil && exprStmt.Expression != nil &&
					exprStmt.Expression.Kind == ast.KindBinaryExpression {
					binExpr := exprStmt.Expression.AsBinaryExpression()
					if binExpr != nil && binExpr.OperatorToken.Kind == ast.KindEqualsToken {
						// Check if left side is this.propName
						if binExpr.Left != nil && binExpr.Left.Kind == ast.KindPropertyAccessExpression {
							propAccess := binExpr.Left.AsPropertyAccessExpression()
							if propAccess != nil && propAccess.Expression != nil &&
								propAccess.Expression.Kind == ast.KindThisKeyword &&
								propAccess.Name != nil && ast.IsIdentifier(propAccess.Name) {
								nameIdent := propAccess.Name.AsIdentifier()
								if nameIdent != nil && nameIdent.Text == propName {
									// Check if right side is the parameter
									if binExpr.Right != nil && ast.IsIdentifier(binExpr.Right) {
										rightIdent := binExpr.Right.AsIdentifier()
										if rightIdent != nil && rightIdent.Text == propName {
											hasAssignment = true
											isFirstStatement = true
										}
									}
								}
							}
						}
					}
				}
			}
		}

		if !hasAssignment || !isFirstStatement {
			continue
		}

		// Check if types match
		if propDecl.Type == nil || matchingParam.Type == nil {
			continue
		}

		// Get modifier string
		modifierStr := ""
		if propDecl.Modifiers != nil {
			hasPrivate := false
			hasProtected := false
			hasPublic := false
			hasReadonly := false

			for _, mod := range propDecl.Modifiers.Elements {
				switch mod.Kind {
				case ast.KindPrivateKeyword:
					hasPrivate = true
				case ast.KindProtectedKeyword:
					hasProtected = true
				case ast.KindPublicKeyword:
					hasPublic = true
				case ast.KindReadonlyKeyword:
					hasReadonly = true
				}
			}

			if hasPrivate {
				modifierStr = "private"
			} else if hasProtected {
				modifierStr = "protected"
			} else if hasPublic {
				modifierStr = "public"
			}

			if hasReadonly && modifierStr != "" {
				modifierStr = modifierStr + " readonly"
			} else if hasReadonly {
				modifierStr = "readonly"
			}
		}

		// Check if this modifier combination is allowed
		if isAllowed(modifierStr, opts.Allow) {
			continue
		}

		// Report the violation
		ctx.ReportNode(propDecl, buildPreferParameterPropertyMessage(), map[string]string{
			"parameter": propName,
		})
	}
}

var ParameterPropertiesRule = rule.CreateRule(rule.Rule{
	Name: "parameter-properties",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			ast.KindParameter: func(node *ast.Node) {
				// Only check when prefer is "class-property" (default)
				if opts.Prefer != "class-property" {
					return
				}

				param := node.AsParameterDeclaration()
				if param == nil {
					return
				}

				// Skip if not a parameter property
				if !isParameterProperty(param) {
					return
				}

				// Get the modifier string
				modifiers := getModifierString(param)

				// Check if this combination is allowed
				if isAllowed(modifiers, opts.Allow) {
					return
				}

				// Report the violation
				paramName := getParameterName(param)
				if paramName != "" {
					ctx.ReportNode(node, buildPreferClassPropertyMessage(), map[string]string{
						"parameter": paramName,
					})
				}
			},
			ast.KindClassDeclaration: func(node *ast.Node) {
				// Only check when prefer is "parameter-property"
				if opts.Prefer != "parameter-property" {
					return
				}

				classDecl := node.AsClassDeclaration()
				if classDecl == nil {
					return
				}

				checkClassPropertyPattern(ctx, classDecl, opts)
			},
		}
	},
})
