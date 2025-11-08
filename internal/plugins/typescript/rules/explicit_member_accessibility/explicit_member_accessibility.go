package explicit_member_accessibility

import (
	"fmt"
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type AccessibilityLevel string

const (
	AccessibilityExplicit AccessibilityLevel = "explicit"
	AccessibilityNoPublic AccessibilityLevel = "no-public"
	AccessibilityOff      AccessibilityLevel = "off"
)

type MemberOverrides struct {
	Accessors            *AccessibilityLevel `json:"accessors"`
	Constructors         *AccessibilityLevel `json:"constructors"`
	Methods              *AccessibilityLevel `json:"methods"`
	Properties           *AccessibilityLevel `json:"properties"`
	ParameterProperties  *AccessibilityLevel `json:"parameterProperties"`
}

type ExplicitMemberAccessibilityOptions struct {
	Accessibility       AccessibilityLevel `json:"accessibility"`
	IgnoredMethodNames  []string          `json:"ignoredMethodNames"`
	Overrides           MemberOverrides   `json:"overrides"`
}

var ExplicitMemberAccessibilityRule = rule.CreateRule(rule.Rule{
	Name: "explicit-member-accessibility",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := ExplicitMemberAccessibilityOptions{
			Accessibility:      AccessibilityExplicit,
			IgnoredMethodNames: []string{},
			Overrides:          MemberOverrides{},
		}

		// Parse options with dual-format support
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
				if accessibility, ok := optsMap["accessibility"].(string); ok {
					opts.Accessibility = AccessibilityLevel(accessibility)
				}
				if ignoredMethodNames, ok := optsMap["ignoredMethodNames"].([]interface{}); ok {
					for _, name := range ignoredMethodNames {
						if str, ok := name.(string); ok {
							opts.IgnoredMethodNames = append(opts.IgnoredMethodNames, str)
						}
					}
				}
				if overrides, ok := optsMap["overrides"].(map[string]interface{}); ok {
					if accessors, ok := overrides["accessors"].(string); ok {
						level := AccessibilityLevel(accessors)
						opts.Overrides.Accessors = &level
					}
					if constructors, ok := overrides["constructors"].(string); ok {
						level := AccessibilityLevel(constructors)
						opts.Overrides.Constructors = &level
					}
					if methods, ok := overrides["methods"].(string); ok {
						level := AccessibilityLevel(methods)
						opts.Overrides.Methods = &level
					}
					if properties, ok := overrides["properties"].(string); ok {
						level := AccessibilityLevel(properties)
						opts.Overrides.Properties = &level
					}
					if parameterProperties, ok := overrides["parameterProperties"].(string); ok {
						level := AccessibilityLevel(parameterProperties)
						opts.Overrides.ParameterProperties = &level
					}
				}
			}
		}

		getModifiersList := func(node *ast.Node) []*ast.Node {
			switch node.Kind {
			case ast.KindPropertyDeclaration:
				prop := node.AsPropertyDeclaration()
				if prop != nil && prop.Modifiers() != nil {
					return prop.Modifiers().Nodes
				}
			case ast.KindMethodDeclaration:
				method := node.AsMethodDeclaration()
				if method != nil && method.Modifiers() != nil {
					return method.Modifiers().Nodes
				}
			case ast.KindConstructor:
				constructor := node.AsConstructorDeclaration()
				if constructor != nil && constructor.Modifiers() != nil {
					return constructor.Modifiers().Nodes
				}
			case ast.KindGetAccessor:
				accessor := node.AsGetAccessorDeclaration()
				if accessor != nil && accessor.Modifiers() != nil {
					return accessor.Modifiers().Nodes
				}
			case ast.KindSetAccessor:
				accessor := node.AsSetAccessorDeclaration()
				if accessor != nil && accessor.Modifiers() != nil {
					return accessor.Modifiers().Nodes
				}
			case ast.KindParameter:
				param := node.AsParameterDeclaration()
				if param != nil && param.Modifiers() != nil {
					return param.Modifiers().Nodes
				}
			}
			return nil
		}

		hasPublicModifier := func(node *ast.Node) bool {
			modifiers := getModifiersList(node)
			for _, modifier := range modifiers {
				if modifier.Kind == ast.KindPublicKeyword {
					return true
				}
			}
			return false
		}

		hasAccessibilityModifier := func(node *ast.Node) *ast.Node {
			modifiers := getModifiersList(node)
			for _, modifier := range modifiers {
				if modifier.Kind == ast.KindPublicKeyword ||
					modifier.Kind == ast.KindPrivateKeyword ||
					modifier.Kind == ast.KindProtectedKeyword {
					return modifier
				}
			}
			return nil
		}

		getEffectiveAccessibility := func(memberType string) AccessibilityLevel {
			switch memberType {
			case "accessor":
				if opts.Overrides.Accessors != nil {
					return *opts.Overrides.Accessors
				}
			case "constructor":
				if opts.Overrides.Constructors != nil {
					return *opts.Overrides.Constructors
				}
			case "method":
				if opts.Overrides.Methods != nil {
					return *opts.Overrides.Methods
				}
			case "property":
				if opts.Overrides.Properties != nil {
					return *opts.Overrides.Properties
				}
			case "parameterProperty":
				if opts.Overrides.ParameterProperties != nil {
					return *opts.Overrides.ParameterProperties
				}
			}
			return opts.Accessibility
		}

		getMemberName := func(node *ast.Node) string {
			switch node.Kind {
			case ast.KindPropertyDeclaration:
				prop := node.AsPropertyDeclaration()
				if prop != nil && prop.Name() != nil {
					return getMemberNameFromNode(ctx, prop.Name())
				}
			case ast.KindMethodDeclaration:
				method := node.AsMethodDeclaration()
				if method != nil && method.Name() != nil {
					return getMemberNameFromNode(ctx, method.Name())
				}
			case ast.KindGetAccessor:
				accessor := node.AsGetAccessorDeclaration()
				if accessor != nil && accessor.Name() != nil {
					return getMemberNameFromNode(ctx, accessor.Name())
				}
			case ast.KindSetAccessor:
				accessor := node.AsSetAccessorDeclaration()
				if accessor != nil && accessor.Name() != nil {
					return getMemberNameFromNode(ctx, accessor.Name())
				}
			case ast.KindParameter:
				param := node.AsParameterDeclaration()
				if param != nil && param.Name() != nil {
					return getMemberNameFromNode(ctx, param.Name())
				}
			}
			return ""
		}

		getMemberType := func(node *ast.Node) string {
			switch node.Kind {
			case ast.KindPropertyDeclaration:
				return "class property"
			case ast.KindMethodDeclaration:
				return "method definition"
			case ast.KindConstructor:
				return "class constructor"
			case ast.KindGetAccessor:
				return "get property accessor"
			case ast.KindSetAccessor:
				return "set property accessor"
			case ast.KindParameter:
				return "parameter property"
			}
			return "class member"
		}

		checkMember := func(node *ast.Node, memberType string) {
			// Skip private field identifiers (# prefix)
			if node.Kind == ast.KindPropertyDeclaration {
				prop := node.AsPropertyDeclaration()
				if prop != nil && prop.Name() != nil && prop.Name().Kind == ast.KindPrivateIdentifier {
					return
				}
			}
			if node.Kind == ast.KindMethodDeclaration {
				method := node.AsMethodDeclaration()
				if method != nil && method.Name() != nil && method.Name().Kind == ast.KindPrivateIdentifier {
					return
				}
			}

			effectiveAccessibility := getEffectiveAccessibility(memberType)
			if effectiveAccessibility == AccessibilityOff {
				return
			}

			accessibilityModifier := hasAccessibilityModifier(node)
			hasPublic := hasPublicModifier(node)

			if effectiveAccessibility == AccessibilityNoPublic {
				if hasPublic {
					ctx.ReportNodeWithFixes(accessibilityModifier, rule.RuleMessage{
						Id:          "unwantedPublicAccessibility",
						Description: fmt.Sprintf("Public accessibility modifier on %s %s.", getMemberType(node), getMemberName(node)),
					}, rule.RuleFixRemove(ctx.SourceFile, accessibilityModifier))
				}
			} else if effectiveAccessibility == AccessibilityExplicit {
				if accessibilityModifier == nil {
					// Get proper location for reporting
					reportNode := node
					if node.Kind == ast.KindParameter {
						param := node.AsParameterDeclaration()
						if param != nil && param.Name() != nil {
							reportNode = param.Name()
						}
					}

					suggestions := []rule.RuleFix{
						createAccessibilitySuggestion(ctx, node, "public"),
						createAccessibilitySuggestion(ctx, node, "private"),
						createAccessibilitySuggestion(ctx, node, "protected"),
					}

					ctx.ReportNodeWithFixes(reportNode, rule.RuleMessage{
						Id:          "missingAccessibility",
						Description: fmt.Sprintf("Missing accessibility modifier on %s %s.", getMemberType(node), getMemberName(node)),
					}, suggestions...)
				}
			}
		}

		return rule.RuleListeners{
			ast.KindClassDeclaration: func(node *ast.Node) {
				classDecl := node.AsClassDeclaration()
				if classDecl == nil || classDecl.Members == nil {
					return
				}

				for _, member := range classDecl.Members.Nodes {
					switch member.Kind {
					case ast.KindPropertyDeclaration:
						checkMember(member, "property")
					case ast.KindMethodDeclaration:
						methodDecl := member.AsMethodDeclaration()
						if methodDecl != nil {
							memberName := getMemberName(member)
							if slices.Contains(opts.IgnoredMethodNames, memberName) {
								continue
							}
						}
						checkMember(member, "method")
					case ast.KindConstructor:
						checkMember(member, "constructor")
						// Check parameter properties
						constructor := member.AsConstructorDeclaration()
						if constructor != nil && constructor.Parameters != nil {
							for _, param := range constructor.Parameters.Nodes {
								paramDecl := param.AsParameterDeclaration()
								if paramDecl != nil && isParameterProperty(paramDecl) {
									checkMember(param, "parameterProperty")
								}
							}
						}
					case ast.KindGetAccessor:
						checkMember(member, "accessor")
					case ast.KindSetAccessor:
						checkMember(member, "accessor")
					}
				}
			},
		}
	},
})

func getMemberNameFromNode(ctx rule.RuleContext, nameNode *ast.Node) string {
	if nameNode == nil {
		return ""
	}

	switch nameNode.Kind {
	case ast.KindIdentifier:
		id := nameNode.AsIdentifier()
		if id != nil {
			return id.Text
		}
	case ast.KindStringLiteral:
		lit := nameNode.AsStringLiteral()
		if lit != nil {
			// Return the value with quotes for computed names
			return fmt.Sprintf(`"%s"`, lit.Text)
		}
	case ast.KindComputedPropertyName:
		// For computed property names, extract the text
		textRange := utils.TrimNodeTextRange(ctx.SourceFile, nameNode)
		return ctx.SourceFile.Text()[textRange.Pos():textRange.End()]
	case ast.KindPrivateIdentifier:
		// Already handled by caller
		return ""
	default:
		// For other cases, try to get the text
		textRange := utils.TrimNodeTextRange(ctx.SourceFile, nameNode)
		return ctx.SourceFile.Text()[textRange.Pos():textRange.End()]
	}
	return ""
}

func isParameterProperty(param *ast.ParameterDeclaration) bool {
	if param == nil {
		return false
	}

	// Check for accessibility modifiers or readonly
	modifiers := param.Modifiers()
	if modifiers == nil {
		return false
	}

	for _, modifier := range modifiers.Nodes {
		switch modifier.Kind {
		case ast.KindPublicKeyword, ast.KindPrivateKeyword, ast.KindProtectedKeyword, ast.KindReadonlyKeyword:
			return true
		}
	}
	return false
}

func createAccessibilitySuggestion(ctx rule.RuleContext, node *ast.Node, accessibilityType string) rule.RuleFix {
	// Find the insertion position
	insertPos := node.Pos()

	// Get modifiers based on node type
	var modifiers []*ast.Node
	switch node.Kind {
	case ast.KindPropertyDeclaration:
		prop := node.AsPropertyDeclaration()
		if prop != nil && prop.Modifiers() != nil {
			modifiers = prop.Modifiers().Nodes
		}
	case ast.KindMethodDeclaration:
		method := node.AsMethodDeclaration()
		if method != nil && method.Modifiers() != nil {
			modifiers = method.Modifiers().Nodes
		}
	case ast.KindConstructor:
		constructor := node.AsConstructorDeclaration()
		if constructor != nil && constructor.Modifiers() != nil {
			modifiers = constructor.Modifiers().Nodes
		}
	case ast.KindGetAccessor:
		accessor := node.AsGetAccessorDeclaration()
		if accessor != nil && accessor.Modifiers() != nil {
			modifiers = accessor.Modifiers().Nodes
		}
	case ast.KindSetAccessor:
		accessor := node.AsSetAccessorDeclaration()
		if accessor != nil && accessor.Modifiers() != nil {
			modifiers = accessor.Modifiers().Nodes
		}
	case ast.KindParameter:
		param := node.AsParameterDeclaration()
		if param != nil && param.Modifiers() != nil {
			modifiers = param.Modifiers().Nodes
		}
	}

	// Skip past decorators if any
	if len(modifiers) > 0 {
		// Find first non-decorator modifier or use node start
		for _, mod := range modifiers {
			if mod.Kind != ast.KindDecorator {
				insertPos = mod.Pos()
				break
			}
		}
	}

	// For parameter properties, we need to insert after other modifiers like readonly
	if node.Kind == ast.KindParameter {
		param := node.AsParameterDeclaration()
		if param != nil && param.Modifiers() != nil {
			// Find position after decorators but before readonly or name
			foundReadonly := false
			for _, mod := range param.Modifiers().Nodes {
				if mod.Kind == ast.KindReadonlyKeyword {
					foundReadonly = true
					insertPos = mod.Pos()
					break
				}
			}
			if !foundReadonly && param.Name() != nil {
				insertPos = param.Name().Pos()
			}
		}
	}

	return rule.RuleFix{
		Text:  accessibilityType + " ",
		Range: core.NewTextRange(insertPos, insertPos),
	}
}
