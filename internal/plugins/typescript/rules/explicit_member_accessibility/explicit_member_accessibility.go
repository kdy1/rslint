// Package explicit_member_accessibility implements the @typescript-eslint/explicit-member-accessibility rule.
// This rule enforces explicit accessibility modifiers (public, private, protected) on class members,
// improving code clarity and consistency by requiring developers to explicitly declare member visibility.
package explicit_member_accessibility

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type AccessibilityLevel string

const (
	AccessibilityExplicit  AccessibilityLevel = "explicit"
	AccessibilityNoPublic  AccessibilityLevel = "no-public"
	AccessibilityOff       AccessibilityLevel = "off"
)

type MemberOverrides struct {
	Accessibility         *AccessibilityLevel `json:"accessibility,omitempty"`
	ParameterProperties   *AccessibilityLevel `json:"parameterProperties,omitempty"`
	Constructors          *AccessibilityLevel `json:"constructors,omitempty"`
	Methods               *AccessibilityLevel `json:"methods,omitempty"`
	Properties            *AccessibilityLevel `json:"properties,omitempty"`
	Accessors             *AccessibilityLevel `json:"accessors,omitempty"`
}

type ExplicitMemberAccessibilityOptions struct {
	Accessibility        AccessibilityLevel `json:"accessibility"`
	Overrides            MemberOverrides    `json:"overrides"`
	IgnoredMethodNames   []string           `json:"ignoredMethodNames"`
}

func parseOptions(options any) ExplicitMemberAccessibilityOptions {
	opts := ExplicitMemberAccessibilityOptions{
		Accessibility:      AccessibilityExplicit,
		Overrides:          MemberOverrides{},
		IgnoredMethodNames: []string{},
	}

	if options == nil {
		return opts
	}

	var optsMap map[string]interface{}
	if optsArray, ok := options.([]interface{}); ok && len(optsArray) > 0 {
		if m, ok := optsArray[0].(map[string]interface{}); ok {
			optsMap = m
		}
	} else if m, ok := options.(map[string]interface{}); ok {
		optsMap = m
	}

	if optsMap != nil {
		if v, ok := optsMap["accessibility"].(string); ok {
			opts.Accessibility = AccessibilityLevel(v)
		}

		if overrides, ok := optsMap["overrides"].(map[string]interface{}); ok {
			if v, ok := overrides["accessibility"].(string); ok {
				level := AccessibilityLevel(v)
				opts.Overrides.Accessibility = &level
			}
			if v, ok := overrides["parameterProperties"].(string); ok {
				level := AccessibilityLevel(v)
				opts.Overrides.ParameterProperties = &level
			}
			if v, ok := overrides["constructors"].(string); ok {
				level := AccessibilityLevel(v)
				opts.Overrides.Constructors = &level
			}
			if v, ok := overrides["methods"].(string); ok {
				level := AccessibilityLevel(v)
				opts.Overrides.Methods = &level
			}
			if v, ok := overrides["properties"].(string); ok {
				level := AccessibilityLevel(v)
				opts.Overrides.Properties = &level
			}
			if v, ok := overrides["accessors"].(string); ok {
				level := AccessibilityLevel(v)
				opts.Overrides.Accessors = &level
			}
		}

		if ignoredMethodNames, ok := optsMap["ignoredMethodNames"].([]interface{}); ok {
			for _, name := range ignoredMethodNames {
				if str, ok := name.(string); ok {
					opts.IgnoredMethodNames = append(opts.IgnoredMethodNames, str)
				}
			}
		}
	}

	return opts
}

func buildMissingAccessibilityMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingAccessibility",
		Description: "Missing accessibility modifier on class member.",
	}
}

func buildUnwantedPublicAccessibilityMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unwantedPublicAccessibility",
		Description: "Public accessibility modifier on class member.",
	}
}

// Check if a member has an accessibility modifier
func hasAccessibilityModifier(node *ast.Node) bool {
	if node == nil {
		return false
	}

	modifiers := getModifiers(node)
	if modifiers == nil {
		return false
	}

	for _, mod := range modifiers.Nodes {
		if mod.Kind == ast.KindPublicKeyword ||
		   mod.Kind == ast.KindPrivateKeyword ||
		   mod.Kind == ast.KindProtectedKeyword {
			return true
		}
	}
	return false
}

// Check if a member has a public modifier
func hasPublicModifier(node *ast.Node) bool {
	if node == nil {
		return false
	}

	modifiers := getModifiers(node)
	if modifiers == nil {
		return false
	}

	for _, mod := range modifiers.Nodes {
		if mod.Kind == ast.KindPublicKeyword {
			return true
		}
	}
	return false
}

// Get modifiers from a node
func getModifiers(node *ast.Node) *ast.ModifierList {
	switch node.Kind {
	case ast.KindPropertyDeclaration:
		prop := node.AsPropertyDeclaration()
		if prop != nil {
			return prop.Modifiers()
		}
	case ast.KindMethodDeclaration:
		method := node.AsMethodDeclaration()
		if method != nil {
			return method.Modifiers()
		}
	case ast.KindConstructor:
		constructor := node.AsConstructorDeclaration()
		if constructor != nil {
			return constructor.Modifiers()
		}
	case ast.KindGetAccessor:
		accessor := node.AsGetAccessorDeclaration()
		if accessor != nil {
			return accessor.Modifiers()
		}
	case ast.KindSetAccessor:
		accessor := node.AsSetAccessorDeclaration()
		if accessor != nil {
			return accessor.Modifiers()
		}
	case ast.KindParameter:
		param := node.AsParameterDeclaration()
		if param != nil {
			return param.Modifiers()
		}
	}
	return nil
}

// Check if parameter is a parameter property
func isParameterProperty(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindParameter {
		return false
	}

	param := node.AsParameterDeclaration()
	if param == nil || param.Modifiers() == nil {
		return false
	}

	// Parameter properties have public, private, protected, or readonly modifiers
	for _, mod := range param.Modifiers().Nodes {
		if mod.Kind == ast.KindPublicKeyword ||
		   mod.Kind == ast.KindPrivateKeyword ||
		   mod.Kind == ast.KindProtectedKeyword ||
		   mod.Kind == ast.KindReadonlyKeyword {
			return true
		}
	}
	return false
}

// Get the effective accessibility level for a member type
func getEffectiveAccessibility(memberType string, opts ExplicitMemberAccessibilityOptions) AccessibilityLevel {
	switch memberType {
	case "parameterProperty":
		if opts.Overrides.ParameterProperties != nil {
			return *opts.Overrides.ParameterProperties
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
	case "accessor":
		if opts.Overrides.Accessors != nil {
			return *opts.Overrides.Accessors
		}
	}

	if opts.Overrides.Accessibility != nil {
		return *opts.Overrides.Accessibility
	}

	return opts.Accessibility
}

// Get member name for error reporting
func getMemberName(ctx rule.RuleContext, node *ast.Node) string {
	switch node.Kind {
	case ast.KindPropertyDeclaration:
		prop := node.AsPropertyDeclaration()
		if prop != nil && prop.Name() != nil {
			name, _ := utils.GetNameFromMember(ctx.SourceFile, prop.Name())
			return name
		}
	case ast.KindMethodDeclaration:
		method := node.AsMethodDeclaration()
		if method != nil && method.Name() != nil {
			name, _ := utils.GetNameFromMember(ctx.SourceFile, method.Name())
			return name
		}
	case ast.KindConstructor:
		return "constructor"
	case ast.KindGetAccessor:
		accessor := node.AsGetAccessorDeclaration()
		if accessor != nil && accessor.Name() != nil {
			name, _ := utils.GetNameFromMember(ctx.SourceFile, accessor.Name())
			return name
		}
	case ast.KindSetAccessor:
		accessor := node.AsSetAccessorDeclaration()
		if accessor != nil && accessor.Name() != nil {
			name, _ := utils.GetNameFromMember(ctx.SourceFile, accessor.Name())
			return name
		}
	case ast.KindParameter:
		param := node.AsParameterDeclaration()
		if param != nil && param.Name() != nil && param.Name().Kind == ast.KindIdentifier {
			ident := param.Name().AsIdentifier()
			if ident != nil {
				return ident.Text
			}
		}
	}
	return ""
}

// Check if method name should be ignored
func isIgnoredMethodName(ctx rule.RuleContext, node *ast.Node, ignoredMethodNames []string) bool {
	if len(ignoredMethodNames) == 0 {
		return false
	}

	name := getMemberName(ctx, node)
	if name == "" {
		return false
	}

	for _, ignored := range ignoredMethodNames {
		if name == ignored {
			return true
		}
	}
	return false
}

// Check if member should be checked based on accessibility level
func shouldCheckMember(node *ast.Node, accessibilityLevel AccessibilityLevel) (bool, string) {
	if accessibilityLevel == AccessibilityOff {
		return false, ""
	}

	hasModifier := hasAccessibilityModifier(node)
	hasPublic := hasPublicModifier(node)

	if accessibilityLevel == AccessibilityExplicit {
		// Must have an accessibility modifier
		if !hasModifier {
			return true, "missingAccessibility"
		}
	} else if accessibilityLevel == AccessibilityNoPublic {
		// Cannot have public modifier
		if hasPublic {
			return true, "unwantedPublicAccessibility"
		}
	}

	return false, ""
}

var ExplicitMemberAccessibilityRule = rule.CreateRule(rule.Rule{
	Name: "explicit-member-accessibility",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		checkMember := func(node *ast.Node, memberType string) {
			effectiveAccessibility := getEffectiveAccessibility(memberType, opts)

			shouldReport, messageType := shouldCheckMember(node, effectiveAccessibility)
			if !shouldReport {
				return
			}

			// For methods, check if the name is in the ignored list
			if memberType == "method" && isIgnoredMethodName(ctx, node, opts.IgnoredMethodNames) {
				return
			}

			// Report the error
			var message rule.RuleMessage
			if messageType == "unwantedPublicAccessibility" {
				message = buildUnwantedPublicAccessibilityMessage()
			} else {
				message = buildMissingAccessibilityMessage()
			}
			ctx.ReportNode(node, message)
		}

		checkParameterProperty := func(node *ast.Node) {
			if !isParameterProperty(node) {
				return
			}
			checkMember(node, "parameterProperty")
		}

		return rule.RuleListeners{
			ast.KindPropertyDeclaration: func(node *ast.Node) {
				checkMember(node, "property")
			},
			ast.KindMethodDeclaration: func(node *ast.Node) {
				checkMember(node, "method")
			},
			ast.KindConstructor: func(node *ast.Node) {
				checkMember(node, "constructor")

				// Also check parameter properties
				constructor := node.AsConstructorDeclaration()
				if constructor != nil && constructor.Parameters != nil && constructor.Parameters.Nodes != nil {
					for _, param := range constructor.Parameters.Nodes {
						checkParameterProperty(param)
					}
				}
			},
			ast.KindGetAccessor: func(node *ast.Node) {
				checkMember(node, "accessor")
			},
			ast.KindSetAccessor: func(node *ast.Node) {
				checkMember(node, "accessor")
			},
		}
	},
})
