package parameter_properties

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// ParameterPropertiesOptions defines the configuration options for this rule
type ParameterPropertiesOptions struct {
	Prefer string   // "class-property" or "parameter-property"
	Allow  []string // allowed modifiers
}

// parseOptions parses and validates the rule options
func parseOptions(options any) ParameterPropertiesOptions {
	opts := ParameterPropertiesOptions{
		Prefer: "class-property",
		Allow:  []string{},
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

	return opts
}

// ParameterPropertiesRule implements the parameter-properties rule
// Require or disallow parameter properties in class constructors
var ParameterPropertiesRule = rule.CreateRule(rule.Rule{
	Name: "parameter-properties",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := parseOptions(options)

	if opts.Prefer == "class-property" {
		return runPreferClassProperty(ctx, opts)
	}
	return runPreferParameterProperty(ctx, opts)
}

// runPreferClassProperty enforces that parameter properties should be class properties
func runPreferClassProperty(ctx rule.RuleContext, opts ParameterPropertiesOptions) rule.RuleListeners {
	return rule.RuleListeners{
		ast.KindParameter: func(node *ast.Node) {
			param := node.AsParameterDeclaration()
			if param == nil {
				return
			}

			// Check if this is a parameter property
			modifiers := getParameterPropertyModifiers(param)
			if modifiers == "" {
				return
			}

			// Check if this modifier combination is allowed
			if isModifierAllowed(modifiers, opts.Allow) {
				return
			}

			// Report the violation
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "preferClassProperty",
				Description: "Parameter property '" + getParameterName(param) + "' should be a class property.",
			})
		},
	}
}

// runPreferParameterProperty enforces that class properties should be parameter properties
func runPreferParameterProperty(ctx rule.RuleContext, opts ParameterPropertiesOptions) rule.RuleListeners {
	type classInfo struct {
		properties  map[string]*ast.Node
		constructor *ast.Node
		assignments map[string]*ast.Node
		params      map[string]*ast.Node
	}

	var stack []*classInfo

	return rule.RuleListeners{
		ast.KindClassDeclaration: func(node *ast.Node) {
			// Push new class context
			stack = append(stack, &classInfo{
				properties:  make(map[string]*ast.Node),
				assignments: make(map[string]*ast.Node),
				params:      make(map[string]*ast.Node),
			})
		},

		rule.ListenerOnExit(ast.KindClassDeclaration): func(node *ast.Node) {
			if len(stack) == 0 {
				return
			}

			// Pop and analyze class
			info := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			// Check each property to see if it should be a parameter property
			for name, propNode := range info.properties {
				// Must have matching parameter and assignment
				_, hasParam := info.params[name]
				_, hasAssignment := info.assignments[name]

				if !hasParam || !hasAssignment {
					continue
				}

				prop := propNode.AsPropertyDeclaration()
				if prop == nil {
					continue
				}

				// Get modifiers
				modifiers := getPropertyModifiers(prop)
				if modifiers == "" {
					modifiers = "public"
				}

				// Check if allowed
				if isModifierAllowed(modifiers, opts.Allow) {
					continue
				}

				// Report violation
				ctx.ReportNode(propNode, rule.RuleMessage{
					Id:          "preferParameterProperty",
					Description: "Class property '" + name + "' should be a parameter property.",
				})
			}
		},

		ast.KindPropertyDeclaration: func(node *ast.Node) {
			if len(stack) == 0 {
				return
			}

			prop := node.AsPropertyDeclaration()
			if prop == nil {
				return
			}

			// Skip static properties
			if ast.GetCombinedModifierFlags(node)&ast.ModifierFlagsStatic != 0 {
				return
			}

			// Skip properties with initializers
			if prop.Initializer != nil {
				return
			}

			name := getPropertyName(prop)
			if name != "" {
				info := stack[len(stack)-1]
				info.properties[name] = node
			}
		},

		ast.KindConstructor: func(node *ast.Node) {
			if len(stack) == 0 {
				return
			}

			info := stack[len(stack)-1]
			info.constructor = node

			constructor := node.AsConstructorDeclaration()
			if constructor == nil || constructor.Parameters == nil {
				return
			}

			// Collect parameters
			for _, paramNode := range constructor.Parameters.Nodes {
				if paramNode.Kind != ast.KindParameter {
					continue
				}

				param := paramNode.AsParameterDeclaration()
				if param == nil {
					continue
				}

				name := getParameterName(param)
				if name != "" {
					info.params[name] = paramNode
				}
			}

			// Analyze constructor body for assignments
			if constructor.Body == nil || constructor.Body.Statements == nil {
				return
			}

			for _, stmtNode := range constructor.Body.Statements() {
				if stmtNode.Kind != ast.KindExpressionStatement {
					continue
				}

				exprStmt := stmtNode.AsExpressionStatement()
				if exprStmt == nil || exprStmt.Expression == nil {
					continue
				}

				// Check for assignment: this.prop = param
				if exprStmt.Expression.Kind != ast.KindBinaryExpression {
					continue
				}

				binExpr := exprStmt.Expression.AsBinaryExpression()
				if binExpr == nil || binExpr.OperatorToken.Kind != ast.KindEqualsToken {
					continue
				}

				// Left side should be this.prop
				if binExpr.Left == nil || binExpr.Left.Kind != ast.KindPropertyAccessExpression {
					continue
				}

				propAccess := binExpr.Left.AsPropertyAccessExpression()
				if propAccess == nil || propAccess.Expression == nil {
					continue
				}

				if propAccess.Expression.Kind != ast.KindThisKeyword {
					continue
				}

				propName := ""
				if propAccess.Name() != nil && propAccess.Name().Kind == ast.KindIdentifier {
					propName = propAccess.Name().AsIdentifier().Text
				}

				if propName != "" {
					info.assignments[propName] = stmtNode
				}
			}
		},
	}
}

// getParameterPropertyModifiers returns the modifiers string for a parameter property
func getParameterPropertyModifiers(param *ast.ParameterDeclaration) string {
	if param.Modifiers() == nil {
		return ""
	}

	var modifiers []string
	hasReadonly := false
	hasAccess := ""

	for _, mod := range param.Modifiers().Nodes {
		switch mod.Kind {
		case ast.KindReadonlyKeyword:
			hasReadonly = true
		case ast.KindPrivateKeyword:
			hasAccess = "private"
		case ast.KindProtectedKeyword:
			hasAccess = "protected"
		case ast.KindPublicKeyword:
			hasAccess = "public"
		}
	}

	if hasAccess != "" {
		modifiers = append(modifiers, hasAccess)
	}
	if hasReadonly {
		modifiers = append(modifiers, "readonly")
	}

	if len(modifiers) == 0 {
		return ""
	}

	result := modifiers[0]
	for i := 1; i < len(modifiers); i++ {
		result += " " + modifiers[i]
	}
	return result
}

// getPropertyModifiers returns the modifiers string for a property
func getPropertyModifiers(prop *ast.PropertyDeclaration) string {
	flags := ast.GetCombinedModifierFlags(&prop.Node)
	var modifiers []string

	hasAccess := ""
	if flags&ast.ModifierFlagsPrivate != 0 {
		hasAccess = "private"
	} else if flags&ast.ModifierFlagsProtected != 0 {
		hasAccess = "protected"
	} else if flags&ast.ModifierFlagsPublic != 0 {
		hasAccess = "public"
	}

	if hasAccess != "" {
		modifiers = append(modifiers, hasAccess)
	}

	if flags&ast.ModifierFlagsReadonly != 0 {
		modifiers = append(modifiers, "readonly")
	}

	if len(modifiers) == 0 {
		return ""
	}

	result := modifiers[0]
	for i := 1; i < len(modifiers); i++ {
		result += " " + modifiers[i]
	}
	return result
}

// isModifierAllowed checks if a modifier combination is in the allow list
func isModifierAllowed(modifier string, allow []string) bool {
	for _, allowed := range allow {
		if modifier == allowed {
			return true
		}
	}
	return false
}

// getParameterName returns the name of a parameter
func getParameterName(param *ast.ParameterDeclaration) string {
	if param.Name() == nil {
		return ""
	}

	if param.Name().Kind == ast.KindIdentifier {
		ident := param.Name().AsIdentifier()
		if ident != nil {
			return ident.Text
		}
	}

	return ""
}

// getPropertyName returns the name of a property
func getPropertyName(prop *ast.PropertyDeclaration) string {
	if prop.Name() == nil {
		return ""
	}

	if prop.Name().Kind == ast.KindIdentifier {
		ident := prop.Name().AsIdentifier()
		if ident != nil {
			return ident.Text
		}
	}

	return ""
}
