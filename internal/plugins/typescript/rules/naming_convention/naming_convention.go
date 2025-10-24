package naming_convention

import (
	"regexp"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Format types supported by the naming-convention rule
type Format string

const (
	FormatCamelCase       Format = "camelCase"
	FormatStrictCamelCase Format = "strictCamelCase"
	FormatPascalCase      Format = "PascalCase"
	FormatStrictPascalCase Format = "StrictPascalCase"
	FormatSnakeCase       Format = "snake_case"
	FormatUpperCase       Format = "UPPER_CASE"
)

// Selector types for the naming-convention rule
type Selector string

const (
	SelectorDefault               Selector = "default"
	SelectorVariable              Selector = "variable"
	SelectorFunction              Selector = "function"
	SelectorClass                 Selector = "class"
	SelectorInterface             Selector = "interface"
	SelectorTypeAlias             Selector = "typeAlias"
	SelectorEnum                  Selector = "enum"
	SelectorEnumMember            Selector = "enumMember"
	SelectorTypeParameter         Selector = "typeParameter"
	SelectorParameter             Selector = "parameter"
	SelectorClassProperty         Selector = "classProperty"
	SelectorClassMethod           Selector = "classMethod"
	SelectorObjectLiteralProperty Selector = "objectLiteralProperty"
	SelectorObjectLiteralMethod   Selector = "objectLiteralMethod"
	SelectorVariableLike          Selector = "variableLike"
	SelectorTypeLike              Selector = "typeLike"
	SelectorMemberLike            Selector = "memberLike"
)

// Modifier options for selectors
type Modifier string

const (
	ModifierPrivate       Modifier = "private"
	ModifierProtected     Modifier = "protected"
	ModifierPublic        Modifier = "public"
	ModifierStatic        Modifier = "static"
	ModifierReadonly      Modifier = "readonly"
	ModifierAbstract      Modifier = "abstract"
	ModifierConst         Modifier = "const"
	ModifierAsync         Modifier = "async"
	ModifierGlobal        Modifier = "global"
	ModifierExported      Modifier = "exported"
	ModifierDestructured  Modifier = "destructured"
	ModifierUnused        Modifier = "unused"
	ModifierRequiresQuotes Modifier = "requiresQuotes"
)

// UnderscoreOption specifies how underscores should be handled
type UnderscoreOption string

const (
	UnderscoreAllow           UnderscoreOption = "allow"
	UnderscoreRequire         UnderscoreOption = "require"
	UnderscoreForbid          UnderscoreOption = "forbid"
	UnderscoreAllowDouble     UnderscoreOption = "allowDouble"
	UnderscoreAllowSingleOrDouble UnderscoreOption = "allowSingleOrDouble"
)

// NamingConventionConfig represents a single naming convention configuration
type NamingConventionConfig struct {
	Selector            []Selector         `json:"selector"`
	Modifiers           []Modifier         `json:"modifiers"`
	Types               []string           `json:"types"`
	Format              []Format           `json:"format"`
	Custom              *CustomPattern     `json:"custom"`
	LeadingUnderscore   UnderscoreOption   `json:"leadingUnderscore"`
	TrailingUnderscore  UnderscoreOption   `json:"trailingUnderscore"`
	Prefix              []string           `json:"prefix"`
	Suffix              []string           `json:"suffix"`
	Filter              *FilterPattern     `json:"filter"`
}

// CustomPattern represents a custom regex pattern for matching
type CustomPattern struct {
	Regex string `json:"regex"`
	Match bool   `json:"match"`
}

// FilterPattern represents a filter regex pattern
type FilterPattern struct {
	Regex string `json:"regex"`
	Match bool   `json:"match"`
}

// NamingConventionOptions holds the complete configuration for the rule
type NamingConventionOptions struct {
	Formats []NamingConventionConfig `json:"formats"`
}

var NamingConventionRule = rule.CreateRule(rule.Rule{
	Name: "naming-convention",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// Parse options
		opts := parseNamingConventionOptions(options)

		return rule.RuleListeners{
			ast.KindVariableDeclaration: func(node *ast.Node) {
				validateVariableDeclaration(ctx, node, opts)
			},
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				validateFunctionDeclaration(ctx, node, opts)
			},
			ast.KindClassDeclaration: func(node *ast.Node) {
				validateClassDeclaration(ctx, node, opts)
			},
			ast.KindInterfaceDeclaration: func(node *ast.Node) {
				validateInterfaceDeclaration(ctx, node, opts)
			},
			ast.KindTypeAliasDeclaration: func(node *ast.Node) {
				validateTypeAliasDeclaration(ctx, node, opts)
			},
			ast.KindEnumDeclaration: func(node *ast.Node) {
				validateEnumDeclaration(ctx, node, opts)
			},
			ast.KindTypeParameter: func(node *ast.Node) {
				validateTypeParameter(ctx, node, opts)
			},
			ast.KindParameter: func(node *ast.Node) {
				validateParameter(ctx, node, opts)
			},
			ast.KindPropertyDeclaration: func(node *ast.Node) {
				validatePropertyDeclaration(ctx, node, opts)
			},
			ast.KindMethodDeclaration: func(node *ast.Node) {
				validateMethodDeclaration(ctx, node, opts)
			},
		}
	},
})

// parseNamingConventionOptions parses the options from the rule configuration
func parseNamingConventionOptions(options any) NamingConventionOptions {
	opts := NamingConventionOptions{
		Formats: []NamingConventionConfig{},
	}

	if options == nil {
		return opts
	}

	// Handle array format: [config1, config2, ...]
	if optArray, isArray := options.([]interface{}); isArray {
		for _, opt := range optArray {
			if configMap, ok := opt.(map[string]interface{}); ok {
				config := parseNamingConventionConfig(configMap)
				opts.Formats = append(opts.Formats, config)
			}
		}
	}

	return opts
}

// parseNamingConventionConfig parses a single naming convention configuration
func parseNamingConventionConfig(configMap map[string]interface{}) NamingConventionConfig {
	config := NamingConventionConfig{
		Selector:           []Selector{},
		Modifiers:          []Modifier{},
		Types:              []string{},
		Format:             []Format{},
		LeadingUnderscore:  UnderscoreAllow,
		TrailingUnderscore: UnderscoreAllow,
		Prefix:             []string{},
		Suffix:             []string{},
	}

	// Parse selector (can be string or array)
	if selectorVal, ok := configMap["selector"]; ok {
		if selectorStr, ok := selectorVal.(string); ok {
			config.Selector = []Selector{Selector(selectorStr)}
		} else if selectorArr, ok := selectorVal.([]interface{}); ok {
			for _, s := range selectorArr {
				if str, ok := s.(string); ok {
					config.Selector = append(config.Selector, Selector(str))
				}
			}
		}
	}

	// Parse modifiers
	if modifiersVal, ok := configMap["modifiers"]; ok {
		if modifiersArr, ok := modifiersVal.([]interface{}); ok {
			for _, m := range modifiersArr {
				if str, ok := m.(string); ok {
					config.Modifiers = append(config.Modifiers, Modifier(str))
				}
			}
		}
	}

	// Parse types
	if typesVal, ok := configMap["types"]; ok {
		if typesArr, ok := typesVal.([]interface{}); ok {
			for _, t := range typesArr {
				if str, ok := t.(string); ok {
					config.Types = append(config.Types, str)
				}
			}
		}
	}

	// Parse format
	if formatVal, ok := configMap["format"]; ok {
		if formatArr, ok := formatVal.([]interface{}); ok {
			for _, f := range formatArr {
				if str, ok := f.(string); ok {
					config.Format = append(config.Format, Format(str))
				}
			}
		}
	}

	// Parse custom pattern
	if customVal, ok := configMap["custom"]; ok {
		if customMap, ok := customVal.(map[string]interface{}); ok {
			config.Custom = &CustomPattern{}
			if regex, ok := customMap["regex"].(string); ok {
				config.Custom.Regex = regex
			}
			if match, ok := customMap["match"].(bool); ok {
				config.Custom.Match = match
			}
		}
	}

	// Parse filter pattern
	if filterVal, ok := configMap["filter"]; ok {
		if filterMap, ok := filterVal.(map[string]interface{}); ok {
			config.Filter = &FilterPattern{}
			if regex, ok := filterMap["regex"].(string); ok {
				config.Filter.Regex = regex
			}
			if match, ok := filterMap["match"].(bool); ok {
				config.Filter.Match = match
			}
		}
	}

	// Parse leading underscore
	if leadingUnderscore, ok := configMap["leadingUnderscore"].(string); ok {
		config.LeadingUnderscore = UnderscoreOption(leadingUnderscore)
	}

	// Parse trailing underscore
	if trailingUnderscore, ok := configMap["trailingUnderscore"].(string); ok {
		config.TrailingUnderscore = UnderscoreOption(trailingUnderscore)
	}

	// Parse prefix
	if prefixVal, ok := configMap["prefix"]; ok {
		if prefixArr, ok := prefixVal.([]interface{}); ok {
			for _, p := range prefixArr {
				if str, ok := p.(string); ok {
					config.Prefix = append(config.Prefix, str)
				}
			}
		}
	}

	// Parse suffix
	if suffixVal, ok := configMap["suffix"]; ok {
		if suffixArr, ok := suffixVal.([]interface{}); ok {
			for _, s := range suffixArr {
				if str, ok := s.(string); ok {
					config.Suffix = append(config.Suffix, str)
				}
			}
		}
	}

	return config
}

// Validation functions for different node types

func validateVariableDeclaration(ctx rule.RuleContext, node *ast.Node, opts NamingConventionOptions) {
	// TODO: Implement variable naming validation
	// 1. Extract variable name from declaration
	// 2. Determine modifiers (const, exported, global, etc.)
	// 3. Check against matching configurations
	// 4. Report violations with specific error messages
}

func validateFunctionDeclaration(ctx rule.RuleContext, node *ast.Node, opts NamingConventionOptions) {
	// TODO: Implement function naming validation
}

func validateClassDeclaration(ctx rule.RuleContext, node *ast.Node, opts NamingConventionOptions) {
	// TODO: Implement class naming validation
}

func validateInterfaceDeclaration(ctx rule.RuleContext, node *ast.Node, opts NamingConventionOptions) {
	// TODO: Implement interface naming validation
}

func validateTypeAliasDeclaration(ctx rule.RuleContext, node *ast.Node, opts NamingConventionOptions) {
	// TODO: Implement type alias naming validation
}

func validateEnumDeclaration(ctx rule.RuleContext, node *ast.Node, opts NamingConventionOptions) {
	// TODO: Implement enum naming validation
	// Also need to check enum members
}

func validateTypeParameter(ctx rule.RuleContext, node *ast.Node, opts NamingConventionOptions) {
	// TODO: Implement type parameter naming validation
}

func validateParameter(ctx rule.RuleContext, node *ast.Node, opts NamingConventionOptions) {
	// TODO: Implement parameter naming validation
}

func validatePropertyDeclaration(ctx rule.RuleContext, node *ast.Node, opts NamingConventionOptions) {
	// TODO: Implement property naming validation
	// Check for modifiers: private, protected, public, static, readonly, abstract
}

func validateMethodDeclaration(ctx rule.RuleContext, node *ast.Node, opts NamingConventionOptions) {
	// TODO: Implement method naming validation
	// Check for modifiers: private, protected, public, static, abstract, async
}

// Helper functions for format checking

func checkFormat(name string, format Format) bool {
	switch format {
	case FormatCamelCase:
		return isCamelCase(name)
	case FormatStrictCamelCase:
		return isStrictCamelCase(name)
	case FormatPascalCase:
		return isPascalCase(name)
	case FormatStrictPascalCase:
		return isStrictPascalCase(name)
	case FormatSnakeCase:
		return isSnakeCase(name)
	case FormatUpperCase:
		return isUpperCase(name)
	default:
		return false
	}
}

func isCamelCase(name string) bool {
	if len(name) == 0 {
		return false
	}
	// Must start with lowercase letter
	if name[0] < 'a' || name[0] > 'z' {
		return false
	}
	// No underscores allowed (except leading/trailing which are handled separately)
	return !strings.Contains(name, "_")
}

func isStrictCamelCase(name string) bool {
	if !isCamelCase(name) {
		return false
	}
	// Check for consecutive uppercase letters
	for i := 0; i < len(name)-1; i++ {
		if name[i] >= 'A' && name[i] <= 'Z' && name[i+1] >= 'A' && name[i+1] <= 'Z' {
			return false
		}
	}
	return true
}

func isPascalCase(name string) bool {
	if len(name) == 0 {
		return false
	}
	// Must start with uppercase letter
	if name[0] < 'A' || name[0] > 'Z' {
		return false
	}
	// No underscores allowed
	return !strings.Contains(name, "_")
}

func isStrictPascalCase(name string) bool {
	if !isPascalCase(name) {
		return false
	}
	// Check for consecutive uppercase letters
	for i := 0; i < len(name)-1; i++ {
		if name[i] >= 'A' && name[i] <= 'Z' && name[i+1] >= 'A' && name[i+1] <= 'Z' {
			return false
		}
	}
	return true
}

func isSnakeCase(name string) bool {
	if len(name) == 0 {
		return false
	}
	// Must be all lowercase with optional underscores
	pattern := regexp.MustCompile(`^[a-z][a-z0-9]*(_[a-z0-9]+)*$`)
	return pattern.MatchString(name)
}

func isUpperCase(name string) bool {
	if len(name) == 0 {
		return false
	}
	// Must be all uppercase with optional underscores
	pattern := regexp.MustCompile(`^[A-Z][A-Z0-9]*(_[A-Z0-9]+)*$`)
	return pattern.MatchString(name)
}

func checkLeadingUnderscore(name string, option UnderscoreOption) (bool, string) {
	hasLeading := strings.HasPrefix(name, "_")
	hasDouble := strings.HasPrefix(name, "__")

	switch option {
	case UnderscoreRequire:
		if !hasLeading {
			return false, "must have a leading underscore"
		}
		return true, strings.TrimLeft(name, "_")
	case UnderscoreForbid:
		if hasLeading {
			return false, "must not have a leading underscore"
		}
		return true, name
	case UnderscoreAllowDouble:
		if hasLeading && !hasDouble {
			return false, "must have either no leading underscore or double leading underscore"
		}
		return true, strings.TrimLeft(name, "_")
	case UnderscoreAllowSingleOrDouble:
		return true, strings.TrimLeft(name, "_")
	case UnderscoreAllow:
		return true, strings.TrimLeft(name, "_")
	default:
		return true, name
	}
}

func checkTrailingUnderscore(name string, option UnderscoreOption) (bool, string) {
	hasTrailing := strings.HasSuffix(name, "_")
	hasDouble := strings.HasSuffix(name, "__")

	switch option {
	case UnderscoreRequire:
		if !hasTrailing {
			return false, "must have a trailing underscore"
		}
		return true, strings.TrimRight(name, "_")
	case UnderscoreForbid:
		if hasTrailing {
			return false, "must not have a trailing underscore"
		}
		return true, name
	case UnderscoreAllowDouble:
		if hasTrailing && !hasDouble {
			return false, "must have either no trailing underscore or double trailing underscore"
		}
		return true, strings.TrimRight(name, "_")
	case UnderscoreAllowSingleOrDouble:
		return true, strings.TrimRight(name, "_")
	case UnderscoreAllow:
		return true, strings.TrimRight(name, "_")
	default:
		return true, name
	}
}

func checkPrefixSuffix(name string, prefixes []string, suffixes []string) (bool, string, string) {
	// Check if name has one of the required prefixes
	hasPrefix := len(prefixes) == 0
	matchedPrefix := ""
	for _, prefix := range prefixes {
		if strings.HasPrefix(name, prefix) {
			hasPrefix = true
			matchedPrefix = prefix
			break
		}
	}

	// Check if name has one of the required suffixes
	hasSuffix := len(suffixes) == 0
	matchedSuffix := ""
	for _, suffix := range suffixes {
		if strings.HasSuffix(name, suffix) {
			hasSuffix = true
			matchedSuffix = suffix
			break
		}
	}

	if !hasPrefix || !hasSuffix {
		return false, matchedPrefix, matchedSuffix
	}

	// Trim prefix and suffix for format checking
	trimmedName := name
	if matchedPrefix != "" {
		trimmedName = strings.TrimPrefix(trimmedName, matchedPrefix)
	}
	if matchedSuffix != "" {
		trimmedName = strings.TrimSuffix(trimmedName, matchedSuffix)
	}

	return true, matchedPrefix, matchedSuffix
}
