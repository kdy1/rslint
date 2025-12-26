package parameter_properties

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type ParameterPropertiesOptions struct {
	Allow  []string `json:"allow,omitempty"`
	Prefer string   `json:"prefer,omitempty"`
}

func buildPreferClassPropertyMessage(parameter string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferClassProperty",
		Description: fmt.Sprintf("Parameter property '%s' should be converted to a class property.", parameter),
	}
}

func buildPreferParameterPropertyMessage(parameter string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferParameterProperty",
		Description: fmt.Sprintf("Class property '%s' should be converted to a parameter property.", parameter),
	}
}

// Get the modifier string from a parameter
func getParameterModifier(param *ast.ParameterDeclaration) string {
	if param == nil {
		return ""
	}

	modifiers := param.Modifiers()
	if modifiers == nil {
		return ""
	}

	var modifierStrs []string
	for _, mod := range modifiers.Nodes {
		switch mod.Kind {
		case ast.KindPublicKeyword:
			modifierStrs = append(modifierStrs, "public")
		case ast.KindPrivateKeyword:
			modifierStrs = append(modifierStrs, "private")
		case ast.KindProtectedKeyword:
			modifierStrs = append(modifierStrs, "protected")
		case ast.KindReadonlyKeyword:
			modifierStrs = append(modifierStrs, "readonly")
		}
	}
	return strings.Join(modifierStrs, " ")
}

// Check if a parameter has any accessibility modifier or readonly
func hasParameterProperty(param *ast.ParameterDeclaration) bool {
	if param == nil {
		return false
	}

	modifiers := param.Modifiers()
	if modifiers == nil {
		return false
	}

	for _, mod := range modifiers.Nodes {
		if mod.Kind == ast.KindPublicKeyword ||
			mod.Kind == ast.KindPrivateKeyword ||
			mod.Kind == ast.KindProtectedKeyword ||
			mod.Kind == ast.KindReadonlyKeyword {
			return true
		}
	}
	return false
}

// Check if a modifier is allowed
func isModifierAllowed(modifier string, allowList []string) bool {
	for _, allowed := range allowList {
		if allowed == modifier {
			return true
		}
	}
	return false
}

// Get parameter name as string
func getParameterName(param *ast.ParameterDeclaration) string {
	if param == nil {
		return ""
	}

	name := param.Name()
	if name == nil {
		return ""
	}

	if ast.IsIdentifier(name) {
		identifier := name.AsIdentifier()
		if identifier != nil {
			return identifier.Text
		}
	}
	return ""
}

var ParameterPropertiesRule = rule.CreateRule(rule.Rule{
	Name: "parameter-properties",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := ParameterPropertiesOptions{
			Allow:  []string{},
			Prefer: "class-property",
		}

		// Parse options
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
				if allowVal, ok := optsMap["allow"].([]interface{}); ok {
					opts.Allow = make([]string, 0, len(allowVal))
					for _, item := range allowVal {
						if str, ok := item.(string); ok {
							opts.Allow = append(opts.Allow, str)
						}
					}
				}
				if preferVal, ok := optsMap["prefer"].(string); ok {
					opts.Prefer = preferVal
				}
			}
		}

		// Handle prefer: "class-property" (default)
		checkParameterProperty := func(node *ast.Node) {
			if node.Kind != ast.KindConstructor {
				return
			}

			constructor := node.AsConstructorDeclaration()
			if constructor == nil || constructor.Parameters == nil {
				return
			}

			for _, param := range constructor.Parameters.Nodes {
				paramDecl := param.AsParameterDeclaration()
				if paramDecl == nil {
					continue
				}

				// Skip rest parameters
				if paramDecl.DotDotDotToken != nil && paramDecl.DotDotDotToken.Kind == ast.KindDotDotDotToken {
					continue
				}

				// Skip destructuring parameters
				if !ast.IsIdentifier(paramDecl.Name()) {
					continue
				}

				if hasParameterProperty(paramDecl) {
					modifier := getParameterModifier(paramDecl)
					paramName := getParameterName(paramDecl)

					// Check if this modifier is allowed
					if !isModifierAllowed(modifier, opts.Allow) {
						ctx.ReportNode(param, buildPreferClassPropertyMessage(paramName))
					}
				}
			}
		}

		// Handle prefer: "parameter-property"
		checkClassProperty := func(classNode *ast.Node) {
			if classNode.Kind != ast.KindClassDeclaration && classNode.Kind != ast.KindClassExpression {
				return
			}

			classDecl := classNode.AsClassDeclaration()
			if classDecl == nil {
				return
			}

			// Find constructor
			var constructor *ast.ConstructorDeclaration
			if classDecl.Members != nil {
				for _, member := range classDecl.Members.Nodes {
					if member.Kind == ast.KindConstructor {
						constructor = member.AsConstructorDeclaration()
						break
					}
				}
			}

			if constructor == nil || constructor.Body == nil {
				return
			}

			// Find properties that are assigned in constructor
			if classDecl.Members == nil {
				return
			}

			for _, member := range classDecl.Members.Nodes {
				if member.Kind != ast.KindPropertyDeclaration {
					continue
				}

				propDecl := member.AsPropertyDeclaration()
				if propDecl == nil {
					continue
				}

				propName := propDecl.Name()
				if !ast.IsIdentifier(propName) {
					continue
				}

				propNameIdent := propName.AsIdentifier()
				if propNameIdent == nil {
					continue
				}
				propNameText := propNameIdent.Text

				// Check if there's a constructor parameter with the same name
				var matchingParam *ast.ParameterDeclaration
				if constructor.Parameters != nil {
					for _, param := range constructor.Parameters.Nodes {
						paramDecl := param.AsParameterDeclaration()
						if paramDecl != nil {
							paramName := paramDecl.Name()
							if ast.IsIdentifier(paramName) {
								paramNameIdent := paramName.AsIdentifier()
								if paramNameIdent != nil && paramNameIdent.Text == propNameText {
									matchingParam = paramDecl
									break
								}
							}
						}
					}
				}

				if matchingParam == nil {
					continue
				}

				// Check if the property is assigned in constructor body with this.propName = paramName
				body := constructor.Body
				if body == nil {
					continue
				}

				statements := body.Statements()
				if statements == nil {
					continue
				}

				hasSimpleAssignment := false
				for _, stmt := range statements {
					if stmt.Kind != ast.KindExpressionStatement {
						continue
					}

					exprStmt := stmt.AsExpressionStatement()
					if exprStmt == nil || exprStmt.Expression.Kind != ast.KindBinaryExpression {
						continue
					}

					binExpr := exprStmt.Expression.AsBinaryExpression()
					if binExpr == nil || binExpr.OperatorToken.Kind != ast.KindEqualsToken {
						continue
					}

					// Check if left side is this.propName
					if binExpr.Left.Kind != ast.KindPropertyAccessExpression {
						continue
					}

					propAccess := binExpr.Left.AsPropertyAccessExpression()
					if propAccess == nil || propAccess.Expression.Kind != ast.KindThisKeyword {
						continue
					}

					leftPropNameNode := propAccess.Name()
					if !ast.IsIdentifier(leftPropNameNode) {
						continue
					}

					leftPropIdent := leftPropNameNode.AsIdentifier()
					if leftPropIdent == nil || leftPropIdent.Text != propNameText {
						continue
					}

					// Check if right side is the parameter
					if !ast.IsIdentifier(binExpr.Right) {
						continue
					}

					rightIdent := binExpr.Right.AsIdentifier()
					if rightIdent == nil || rightIdent.Text != propNameText {
						continue
					}

					hasSimpleAssignment = true
					break
				}

				if !hasSimpleAssignment {
					continue
				}

				// Get the modifier from the property
				var modifier string
				propModifiers := propDecl.Modifiers()
				if propModifiers != nil {
					var modifierStrs []string
					for _, mod := range propModifiers.Nodes {
						switch mod.Kind {
						case ast.KindPublicKeyword:
							modifierStrs = append(modifierStrs, "public")
						case ast.KindPrivateKeyword:
							modifierStrs = append(modifierStrs, "private")
						case ast.KindProtectedKeyword:
							modifierStrs = append(modifierStrs, "protected")
						case ast.KindReadonlyKeyword:
							modifierStrs = append(modifierStrs, "readonly")
						}
					}
					modifier = strings.Join(modifierStrs, " ")
				}

				// Check if this modifier is allowed
				if isModifierAllowed(modifier, opts.Allow) {
					continue
				}

				ctx.ReportNode(member, buildPreferParameterPropertyMessage(propNameText))
			}
		}

		if opts.Prefer == "parameter-property" {
			return rule.RuleListeners{
				ast.KindClassDeclaration: checkClassProperty,
				ast.KindClassExpression:  checkClassProperty,
			}
		}

		// Default: prefer class-property
		return rule.RuleListeners{
			ast.KindConstructor: checkParameterProperty,
		}
	},
})
