// Package naming_convention implements the @typescript-eslint/naming-convention rule.
// This rule enforces naming conventions for identifiers throughout TypeScript code.
// It supports multiple selectors, formats, modifiers, and type-based filtering.
package naming_convention

import (
	"regexp"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Format represents a naming format style
type Format string

const (
	FormatCamelCase       Format = "camelCase"
	FormatStrictCamelCase Format = "strictCamelCase"
	FormatPascalCase      Format = "PascalCase"
	FormatStrictPascalCase Format = "StrictPascalCase"
	FormatSnakeCase       Format = "snake_case"
	FormatUpperCase       Format = "UPPER_CASE"
)

// Selector represents what kind of identifier to match
type Selector string

const (
	SelectorDefault              Selector = "default"
	SelectorVariable             Selector = "variable"
	SelectorFunction             Selector = "function"
	SelectorParameter            Selector = "parameter"
	SelectorProperty             Selector = "property"
	SelectorParameterProperty    Selector = "parameterProperty"
	SelectorMethod               Selector = "method"
	SelectorAccessor             Selector = "accessor"
	SelectorEnumMember           Selector = "enumMember"
	SelectorClassProperty        Selector = "classProperty"
	SelectorObjectLiteralProperty Selector = "objectLiteralProperty"
	SelectorTypeProperty         Selector = "typeProperty"
	SelectorClassMethod          Selector = "classMethod"
	SelectorObjectLiteralMethod  Selector = "objectLiteralMethod"
	SelectorTypeMethod           Selector = "typeMethod"
	SelectorClass                Selector = "class"
	SelectorInterface            Selector = "interface"
	SelectorTypeAlias            Selector = "typeAlias"
	SelectorEnum                 Selector = "enum"
	SelectorTypeParameter        Selector = "typeParameter"
	SelectorImport               Selector = "import"
	SelectorVariableLike         Selector = "variableLike"
	SelectorMemberLike           Selector = "memberLike"
	SelectorTypeLike             Selector = "typeLike"
	SelectorClassicAccessor      Selector = "classicAccessor"
	SelectorAutoAccessor         Selector = "autoAccessor"
)

// Modifier represents additional constraints on identifiers
type Modifier string

const (
	ModifierConst         Modifier = "const"
	ModifierReadonly      Modifier = "readonly"
	ModifierStatic        Modifier = "static"
	ModifierPublic        Modifier = "public"
	ModifierProtected     Modifier = "protected"
	ModifierPrivate       Modifier = "private"
	ModifierPrivateHash   Modifier = "#private"
	ModifierAbstract      Modifier = "abstract"
	ModifierDestructured  Modifier = "destructured"
	ModifierGlobal        Modifier = "global"
	ModifierExported      Modifier = "exported"
	ModifierUnused        Modifier = "unused"
	ModifierRequiresQuotes Modifier = "requiresQuotes"
	ModifierAsync         Modifier = "async"
	ModifierDefault       Modifier = "default"
	ModifierNamespace     Modifier = "namespace"
	ModifierOverride      Modifier = "override"
)

// UnderscoreOption represents how underscores should be handled
type UnderscoreOption string

const (
	UnderscoreAllow             UnderscoreOption = "allow"
	UnderscoreForbid            UnderscoreOption = "forbid"
	UnderscoreRequire           UnderscoreOption = "require"
	UnderscoreRequireDouble     UnderscoreOption = "requireDouble"
	UnderscoreAllowDouble       UnderscoreOption = "allowDouble"
	UnderscoreAllowSingleOrDouble UnderscoreOption = "allowSingleOrDouble"
)

// TypeOption represents primitive type filtering
type TypeOption string

const (
	TypeBoolean  TypeOption = "boolean"
	TypeString   TypeOption = "string"
	TypeNumber   TypeOption = "number"
	TypeFunction TypeOption = "function"
	TypeArray    TypeOption = "array"
)

// CustomMatcher represents a custom regex matcher
type CustomMatcher struct {
	Regex string `json:"regex"`
	Match bool   `json:"match"`
}

// FilterMatcher represents a filter for selectors
type FilterMatcher struct {
	Regex string `json:"regex"`
	Match bool   `json:"match"`
}

// NamingConventionConfig represents a single naming convention configuration
type NamingConventionConfig struct {
	Selector           []Selector         `json:"selector"`
	Format             []Format           `json:"format"`
	Modifiers          []Modifier         `json:"modifiers"`
	Types              []TypeOption       `json:"types"`
	Prefix             []string           `json:"prefix"`
	Suffix             []string           `json:"suffix"`
	LeadingUnderscore  UnderscoreOption   `json:"leadingUnderscore"`
	TrailingUnderscore UnderscoreOption   `json:"trailingUnderscore"`
	Custom             *CustomMatcher     `json:"custom"`
	Filter             *FilterMatcher     `json:"filter"`
}

// parseOptions parses the rule options
func parseOptions(options any) []NamingConventionConfig {
	if options == nil {
		return getDefaultConfig()
	}

	var configs []NamingConventionConfig

	// Handle array of configurations
	if optsArray, ok := options.([]interface{}); ok {
		for _, opt := range optsArray {
			if configMap, ok := opt.(map[string]interface{}); ok {
				config := parseConfig(configMap)
				configs = append(configs, config)
			}
		}
	}

	// If no configs provided, use defaults
	if len(configs) == 0 {
		return getDefaultConfig()
	}

	return configs
}

// parseConfig parses a single configuration object
func parseConfig(configMap map[string]interface{}) NamingConventionConfig {
	config := NamingConventionConfig{
		LeadingUnderscore:  UnderscoreAllow,
		TrailingUnderscore: UnderscoreAllow,
	}

	// Parse selector (can be string or array)
	if selector, ok := configMap["selector"].(string); ok {
		config.Selector = []Selector{Selector(selector)}
	} else if selectorArray, ok := configMap["selector"].([]interface{}); ok {
		for _, s := range selectorArray {
			if str, ok := s.(string); ok {
				config.Selector = append(config.Selector, Selector(str))
			}
		}
	}

	// Parse format
	if format, ok := configMap["format"].([]interface{}); ok {
		for _, f := range format {
			if str, ok := f.(string); ok {
				config.Format = append(config.Format, Format(str))
			} else if f == nil {
				// null format means no format checking
				config.Format = nil
				break
			}
		}
	}

	// Parse modifiers
	if modifiers, ok := configMap["modifiers"].([]interface{}); ok {
		for _, m := range modifiers {
			if str, ok := m.(string); ok {
				config.Modifiers = append(config.Modifiers, Modifier(str))
			}
		}
	}

	// Parse types
	if types, ok := configMap["types"].([]interface{}); ok {
		for _, t := range types {
			if str, ok := t.(string); ok {
				config.Types = append(config.Types, TypeOption(str))
			}
		}
	}

	// Parse prefix
	if prefix, ok := configMap["prefix"].([]interface{}); ok {
		for _, p := range prefix {
			if str, ok := p.(string); ok {
				config.Prefix = append(config.Prefix, str)
			}
		}
	}

	// Parse suffix
	if suffix, ok := configMap["suffix"].([]interface{}); ok {
		for _, s := range suffix {
			if str, ok := s.(string); ok {
				config.Suffix = append(config.Suffix, str)
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

	// Parse custom
	if custom, ok := configMap["custom"].(map[string]interface{}); ok {
		matcher := &CustomMatcher{}
		if regex, ok := custom["regex"].(string); ok {
			matcher.Regex = regex
		}
		if match, ok := custom["match"].(bool); ok {
			matcher.Match = match
		}
		config.Custom = matcher
	}

	// Parse filter
	if filter, ok := configMap["filter"].(map[string]interface{}); ok {
		matcher := &FilterMatcher{}
		if regex, ok := filter["regex"].(string); ok {
			matcher.Regex = regex
		}
		if match, ok := filter["match"].(bool); ok {
			matcher.Match = match
		}
		config.Filter = matcher
	}

	return config
}

// getDefaultConfig returns the default naming convention configuration
func getDefaultConfig() []NamingConventionConfig {
	return []NamingConventionConfig{
		{
			Selector:           []Selector{SelectorDefault},
			Format:             []Format{FormatCamelCase},
			LeadingUnderscore:  UnderscoreAllow,
			TrailingUnderscore: UnderscoreAllow,
		},
		{
			Selector:           []Selector{SelectorImport},
			Format:             []Format{FormatCamelCase, FormatPascalCase},
			LeadingUnderscore:  UnderscoreAllow,
			TrailingUnderscore: UnderscoreAllow,
		},
		{
			Selector:           []Selector{SelectorVariable},
			Format:             []Format{FormatCamelCase, FormatUpperCase},
			LeadingUnderscore:  UnderscoreAllow,
			TrailingUnderscore: UnderscoreAllow,
		},
		{
			Selector:           []Selector{SelectorTypeLike},
			Format:             []Format{FormatPascalCase},
			LeadingUnderscore:  UnderscoreAllow,
			TrailingUnderscore: UnderscoreAllow,
		},
	}
}

// buildDoesNotMatchFormatMessage creates the error message for format violations
func buildDoesNotMatchFormatMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "doesNotMatchFormat",
		Description: "{{type}} name `{{name}}` must match one of the following formats: {{formats}}",
	}
}

// buildDoesNotMatchFormatTrimmedMessage creates the error message for format violations after trimming
func buildDoesNotMatchFormatTrimmedMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "doesNotMatchFormatTrimmed",
		Description: "{{type}} name `{{name}}` must match one of the following formats: {{formats}}",
	}
}

// buildSatisfyCustomMessage creates the error message for custom regex violations
func buildSatisfyCustomMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "satisfyCustom",
		Description: "{{type}} name `{{name}}` must {{satisfyCustom}}",
	}
}

// checkFormat validates that a name matches the specified format
func checkFormat(name string, format Format) bool {
	switch format {
	case FormatCamelCase:
		return matchesCamelCase(name, false)
	case FormatStrictCamelCase:
		return matchesCamelCase(name, true)
	case FormatPascalCase:
		return matchesPascalCase(name, false)
	case FormatStrictPascalCase:
		return matchesPascalCase(name, true)
	case FormatSnakeCase:
		return matchesSnakeCase(name)
	case FormatUpperCase:
		return matchesUpperCase(name)
	}
	return true
}

// matchesCamelCase checks if a name matches camelCase format
func matchesCamelCase(name string, strict bool) bool {
	if len(name) == 0 {
		return true
	}

	// Must start with lowercase letter or underscore
	if !isLowerCase(rune(name[0])) && name[0] != '_' {
		return false
	}

	if strict {
		// Strict mode: no consecutive uppercase letters
		prevUpper := false
		for _, ch := range name {
			if isUpperCase(ch) {
				if prevUpper {
					return false
				}
				prevUpper = true
			} else {
				prevUpper = false
			}
		}
	}

	// Only alphanumeric and underscores allowed
	for _, ch := range name {
		if !isAlphaNumeric(ch) && ch != '_' {
			return false
		}
	}

	return true
}

// matchesPascalCase checks if a name matches PascalCase format
func matchesPascalCase(name string, strict bool) bool {
	if len(name) == 0 {
		return true
	}

	// Must start with uppercase letter
	if !isUpperCase(rune(name[0])) {
		return false
	}

	if strict {
		// Strict mode: no consecutive uppercase letters
		prevUpper := true // First letter is uppercase
		for i, ch := range name {
			if i == 0 {
				continue
			}
			if isUpperCase(ch) {
				if prevUpper {
					return false
				}
				prevUpper = true
			} else {
				prevUpper = false
			}
		}
	}

	// Only alphanumeric allowed
	for _, ch := range name {
		if !isAlphaNumeric(ch) && ch != '_' {
			return false
		}
	}

	return true
}

// matchesSnakeCase checks if a name matches snake_case format
func matchesSnakeCase(name string) bool {
	if len(name) == 0 {
		return true
	}

	// All lowercase with underscores
	for _, ch := range name {
		if !isLowerCase(ch) && !isDigit(ch) && ch != '_' {
			return false
		}
	}

	return true
}

// matchesUpperCase checks if a name matches UPPER_CASE format
func matchesUpperCase(name string) bool {
	if len(name) == 0 {
		return true
	}

	// All uppercase with underscores
	for _, ch := range name {
		if !isUpperCase(ch) && !isDigit(ch) && ch != '_' {
			return false
		}
	}

	return true
}

// isLowerCase checks if a rune is lowercase
func isLowerCase(ch rune) bool {
	return ch >= 'a' && ch <= 'z'
}

// isUpperCase checks if a rune is uppercase
func isUpperCase(ch rune) bool {
	return ch >= 'A' && ch <= 'Z'
}

// isDigit checks if a rune is a digit
func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

// isAlphaNumeric checks if a rune is alphanumeric
func isAlphaNumeric(ch rune) bool {
	return isLowerCase(ch) || isUpperCase(ch) || isDigit(ch)
}

// trimUnderscore handles leading/trailing underscore trimming
func trimUnderscore(name string, option UnderscoreOption, leading bool) (string, bool) {
	if len(name) == 0 {
		return name, true
	}

	hasUnderscore := false
	hasDoubleUnderscore := false

	if leading {
		if strings.HasPrefix(name, "__") {
			hasDoubleUnderscore = true
		} else if strings.HasPrefix(name, "_") {
			hasUnderscore = true
		}
	} else {
		if strings.HasSuffix(name, "__") {
			hasDoubleUnderscore = true
		} else if strings.HasSuffix(name, "_") {
			hasUnderscore = true
		}
	}

	switch option {
	case UnderscoreForbid:
		if hasUnderscore || hasDoubleUnderscore {
			return name, false
		}
	case UnderscoreRequire:
		if !hasUnderscore && !hasDoubleUnderscore {
			return name, false
		}
	case UnderscoreRequireDouble:
		if !hasDoubleUnderscore {
			return name, false
		}
	case UnderscoreAllowDouble:
		// Allow single or double
	case UnderscoreAllowSingleOrDouble:
		// Allow single or double
	case UnderscoreAllow:
		// Allow anything
	}

	// Trim the underscores for further validation
	if hasDoubleUnderscore {
		if leading {
			return name[2:], true
		}
		return name[:len(name)-2], true
	} else if hasUnderscore {
		if leading {
			return name[1:], true
		}
		return name[:len(name)-1], true
	}

	return name, true
}

// trimPrefixSuffix trims prefix/suffix from name
func trimPrefixSuffix(name string, prefixes []string, suffixes []string) (string, bool) {
	// Try to match prefix
	hasPrefix := false
	for _, prefix := range prefixes {
		if strings.HasPrefix(name, prefix) {
			name = name[len(prefix):]
			hasPrefix = true
			break
		}
	}

	if len(prefixes) > 0 && !hasPrefix {
		return name, false
	}

	// Try to match suffix
	hasSuffix := false
	for _, suffix := range suffixes {
		if strings.HasSuffix(name, suffix) {
			name = name[:len(name)-len(suffix)]
			hasSuffix = true
			break
		}
	}

	if len(suffixes) > 0 && !hasSuffix {
		return name, false
	}

	return name, true
}

// validateName validates a name against a configuration
func validateName(name string, config NamingConventionConfig) (bool, string) {
	originalName := name

	// 1. Trim leading underscore
	trimmed, valid := trimUnderscore(name, config.LeadingUnderscore, true)
	if !valid {
		return false, originalName
	}
	name = trimmed

	// 2. Trim trailing underscore
	trimmed, valid = trimUnderscore(name, config.TrailingUnderscore, false)
	if !valid {
		return false, originalName
	}
	name = trimmed

	// 3. Trim prefix/suffix
	trimmed, valid = trimPrefixSuffix(name, config.Prefix, config.Suffix)
	if !valid {
		return false, originalName
	}
	name = trimmed

	// 4. Check custom regex
	if config.Custom != nil {
		matched, _ := regexp.MatchString(config.Custom.Regex, originalName)
		if matched != config.Custom.Match {
			return false, originalName
		}
	}

	// 5. Check format (if name is empty after trimming, it passes all formats)
	if len(name) == 0 {
		return true, originalName
	}

	// If format is nil, skip format checking
	if config.Format == nil {
		return true, originalName
	}

	// Check if name matches any of the specified formats
	for _, format := range config.Format {
		if checkFormat(name, format) {
			return true, originalName
		}
	}

	return false, name
}

// matchesSelector checks if a node matches a selector
func matchesSelector(node *ast.Node, selector Selector, modifiers []Modifier, ctx rule.RuleContext) bool {
	switch selector {
	case SelectorDefault:
		return true
	case SelectorVariable:
		return isVariable(node)
	case SelectorFunction:
		return isFunction(node)
	case SelectorParameter:
		return isParameter(node)
	case SelectorProperty:
		return isProperty(node)
	case SelectorParameterProperty:
		return isParameterProperty(node)
	case SelectorMethod:
		return isMethod(node)
	case SelectorAccessor:
		return isAccessor(node)
	case SelectorEnumMember:
		return node.Kind == ast.KindEnumMember
	case SelectorClassProperty:
		return isClassProperty(node)
	case SelectorObjectLiteralProperty:
		return isObjectLiteralProperty(node)
	case SelectorTypeProperty:
		return isTypeProperty(node)
	case SelectorClassMethod:
		return isClassMethod(node)
	case SelectorObjectLiteralMethod:
		return isObjectLiteralMethod(node)
	case SelectorTypeMethod:
		return isTypeMethod(node)
	case SelectorClass:
		return node.Kind == ast.KindClassDeclaration || node.Kind == ast.KindClassExpression
	case SelectorInterface:
		return node.Kind == ast.KindInterfaceDeclaration
	case SelectorTypeAlias:
		return node.Kind == ast.KindTypeAliasDeclaration
	case SelectorEnum:
		return node.Kind == ast.KindEnumDeclaration
	case SelectorTypeParameter:
		return node.Kind == ast.KindTypeParameter
	case SelectorImport:
		return isImport(node)
	case SelectorVariableLike:
		return isVariableLike(node)
	case SelectorMemberLike:
		return isMemberLike(node)
	case SelectorTypeLike:
		return isTypeLike(node)
	case SelectorClassicAccessor:
		return isClassicAccessor(node)
	case SelectorAutoAccessor:
		return isAutoAccessor(node)
	}
	return false
}

// Helper functions for selector matching
func isVariable(node *ast.Node) bool {
	return node.Kind == ast.KindVariableDeclaration
}

func isFunction(node *ast.Node) bool {
	return node.Kind == ast.KindFunctionDeclaration
}

func isParameter(node *ast.Node) bool {
	return node.Kind == ast.KindParameter
}

func isProperty(node *ast.Node) bool {
	return node.Kind == ast.KindPropertyDeclaration ||
		node.Kind == ast.KindPropertySignature ||
		node.Kind == ast.KindPropertyAssignment
}

func isParameterProperty(node *ast.Node) bool {
	if node.Kind != ast.KindParameter {
		return false
	}
	param := node.AsParameterDeclaration()
	if param == nil {
		return false
	}
	// Check if parameter has modifiers (public, private, protected, readonly)
	return param.Modifiers != nil && len(param.Modifiers.Nodes) > 0
}

func isMethod(node *ast.Node) bool {
	return node.Kind == ast.KindMethodDeclaration ||
		node.Kind == ast.KindMethodSignature
}

func isAccessor(node *ast.Node) bool {
	return node.Kind == ast.KindGetAccessor ||
		node.Kind == ast.KindSetAccessor
}

func isClassProperty(node *ast.Node) bool {
	if node.Kind != ast.KindPropertyDeclaration {
		return false
	}
	// Check if parent is a class
	return utils.IsInClass(node)
}

func isObjectLiteralProperty(node *ast.Node) bool {
	return node.Kind == ast.KindPropertyAssignment ||
		node.Kind == ast.KindShorthandPropertyAssignment
}

func isTypeProperty(node *ast.Node) bool {
	return node.Kind == ast.KindPropertySignature
}

func isClassMethod(node *ast.Node) bool {
	if node.Kind != ast.KindMethodDeclaration {
		return false
	}
	return utils.IsInClass(node)
}

func isObjectLiteralMethod(node *ast.Node) bool {
	return node.Kind == ast.KindMethodDeclaration && !utils.IsInClass(node)
}

func isTypeMethod(node *ast.Node) bool {
	return node.Kind == ast.KindMethodSignature
}

func isImport(node *ast.Node) bool {
	return node.Kind == ast.KindImportClause ||
		node.Kind == ast.KindImportSpecifier ||
		node.Kind == ast.KindNamespaceImport
}

func isVariableLike(node *ast.Node) bool {
	return isVariable(node) || isFunction(node) || isParameter(node)
}

func isMemberLike(node *ast.Node) bool {
	return isProperty(node) || isMethod(node) || isAccessor(node)
}

func isTypeLike(node *ast.Node) bool {
	return node.Kind == ast.KindClassDeclaration ||
		node.Kind == ast.KindInterfaceDeclaration ||
		node.Kind == ast.KindTypeAliasDeclaration ||
		node.Kind == ast.KindEnumDeclaration
}

func isClassicAccessor(node *ast.Node) bool {
	return node.Kind == ast.KindGetAccessor || node.Kind == ast.KindSetAccessor
}

func isAutoAccessor(node *ast.Node) bool {
	// Auto-accessors are properties with the 'accessor' keyword
	if node.Kind != ast.KindPropertyDeclaration {
		return false
	}
	prop := node.AsPropertyDeclaration()
	if prop == nil || prop.Modifiers == nil {
		return false
	}
	// Check for accessor modifier (this would need AST support)
	return false // Placeholder until AST supports accessor keyword detection
}

// matchesModifiers checks if a node matches the required modifiers
func matchesModifiers(node *ast.Node, modifiers []Modifier) bool {
	if len(modifiers) == 0 {
		return true
	}

	for _, modifier := range modifiers {
		if !hasModifier(node, modifier) {
			return false
		}
	}
	return true
}

// hasModifier checks if a node has a specific modifier
func hasModifier(node *ast.Node, modifier Modifier) bool {
	switch modifier {
	case ModifierConst:
		return isConst(node)
	case ModifierReadonly:
		return isReadonly(node)
	case ModifierStatic:
		return isStatic(node)
	case ModifierPublic:
		return isPublic(node)
	case ModifierProtected:
		return isProtected(node)
	case ModifierPrivate:
		return isPrivate(node)
	case ModifierPrivateHash:
		return isPrivateHash(node)
	case ModifierAbstract:
		return isAbstract(node)
	case ModifierAsync:
		return isAsync(node)
	case ModifierExported:
		return isExported(node)
	case ModifierOverride:
		return isOverride(node)
	}
	return false
}

// Modifier checking helper functions
func isConst(node *ast.Node) bool {
	if node.Kind == ast.KindVariableDeclaration {
		// Check if parent VariableDeclarationList has const flag
		parent := utils.GetParent(node)
		if parent != nil && parent.Kind == ast.KindVariableDeclarationList {
			list := parent.AsVariableDeclarationList()
			return list != nil && (list.Flags&ast.NodeFlagsConst) != 0
		}
	}
	return false
}

func isReadonly(node *ast.Node) bool {
	return hasModifierFlag(node, ast.ModifierFlagsReadonly)
}

func isStatic(node *ast.Node) bool {
	return hasModifierFlag(node, ast.ModifierFlagsStatic)
}

func isPublic(node *ast.Node) bool {
	return hasModifierFlag(node, ast.ModifierFlagsPublic)
}

func isProtected(node *ast.Node) bool {
	return hasModifierFlag(node, ast.ModifierFlagsProtected)
}

func isPrivate(node *ast.Node) bool {
	return hasModifierFlag(node, ast.ModifierFlagsPrivate)
}

func isPrivateHash(node *ast.Node) bool {
	// Check if name starts with #
	name := getIdentifierName(node)
	return len(name) > 0 && name[0] == '#'
}

func isAbstract(node *ast.Node) bool {
	return hasModifierFlag(node, ast.ModifierFlagsAbstract)
}

func isAsync(node *ast.Node) bool {
	return hasModifierFlag(node, ast.ModifierFlagsAsync)
}

func isExported(node *ast.Node) bool {
	return hasModifierFlag(node, ast.ModifierFlagsExport)
}

func isOverride(node *ast.Node) bool {
	return hasModifierFlag(node, ast.ModifierFlagsOverride)
}

// hasModifierFlag checks if a node has a specific modifier flag
func hasModifierFlag(node *ast.Node, flag ast.ModifierFlags) bool {
	// Get modifiers based on node type
	switch node.Kind {
	case ast.KindPropertyDeclaration:
		prop := node.AsPropertyDeclaration()
		if prop != nil && prop.Modifiers != nil {
			return checkModifiers(prop.Modifiers, flag)
		}
	case ast.KindMethodDeclaration:
		method := node.AsMethodDeclaration()
		if method != nil && method.Modifiers != nil {
			return checkModifiers(method.Modifiers, flag)
		}
	case ast.KindParameter:
		param := node.AsParameterDeclaration()
		if param != nil && param.Modifiers != nil {
			return checkModifiers(param.Modifiers, flag)
		}
	case ast.KindClassDeclaration:
		class := node.AsClassDeclaration()
		if class != nil && class.Modifiers != nil {
			return checkModifiers(class.Modifiers, flag)
		}
	case ast.KindFunctionDeclaration:
		fn := node.AsFunctionDeclaration()
		if fn != nil && fn.Modifiers != nil {
			return checkModifiers(fn.Modifiers, flag)
		}
	}
	return false
}

// checkModifiers checks if modifier list contains the specified flag
func checkModifiers(modifiers *ast.NodeArray, flag ast.ModifierFlags) bool {
	if modifiers == nil || modifiers.Nodes == nil {
		return false
	}
	for _, mod := range modifiers.Nodes {
		if mod.Kind == getModifierKind(flag) {
			return true
		}
	}
	return false
}

// getModifierKind converts ModifierFlags to SyntaxKind
func getModifierKind(flag ast.ModifierFlags) ast.SyntaxKind {
	switch flag {
	case ast.ModifierFlagsPublic:
		return ast.KindPublicKeyword
	case ast.ModifierFlagsPrivate:
		return ast.KindPrivateKeyword
	case ast.ModifierFlagsProtected:
		return ast.KindProtectedKeyword
	case ast.ModifierFlagsStatic:
		return ast.KindStaticKeyword
	case ast.ModifierFlagsReadonly:
		return ast.KindReadonlyKeyword
	case ast.ModifierFlagsAbstract:
		return ast.KindAbstractKeyword
	case ast.ModifierFlagsAsync:
		return ast.KindAsyncKeyword
	case ast.ModifierFlagsExport:
		return ast.KindExportKeyword
	case ast.ModifierFlagsOverride:
		return ast.KindOverrideKeyword
	}
	return ast.KindUnknown
}

// getIdentifierName extracts the name from a node
func getIdentifierName(node *ast.Node) string {
	if node == nil {
		return ""
	}

	switch node.Kind {
	case ast.KindIdentifier:
		ident := node.AsIdentifier()
		if ident != nil {
			return ident.Text
		}
	case ast.KindVariableDeclaration:
		varDecl := node.AsVariableDeclaration()
		if varDecl != nil && varDecl.Name != nil {
			return getIdentifierName(varDecl.Name)
		}
	case ast.KindParameter:
		param := node.AsParameterDeclaration()
		if param != nil && param.Name != nil {
			return getIdentifierName(param.Name)
		}
	case ast.KindPropertyDeclaration, ast.KindPropertySignature:
		prop := node.AsPropertyDeclaration()
		if prop != nil && prop.Name != nil {
			return getIdentifierName(prop.Name)
		}
	case ast.KindMethodDeclaration, ast.KindMethodSignature:
		method := node.AsMethodDeclaration()
		if method != nil && method.Name != nil {
			return getIdentifierName(method.Name)
		}
	case ast.KindFunctionDeclaration:
		fn := node.AsFunctionDeclaration()
		if fn != nil && fn.Name != nil {
			return getIdentifierName(fn.Name)
		}
	case ast.KindClassDeclaration:
		class := node.AsClassDeclaration()
		if class != nil && class.Name != nil {
			return getIdentifierName(class.Name)
		}
	case ast.KindInterfaceDeclaration:
		iface := node.AsInterfaceDeclaration()
		if iface != nil && iface.Name != nil {
			return getIdentifierName(iface.Name)
		}
	case ast.KindTypeAliasDeclaration:
		typeAlias := node.AsTypeAliasDeclaration()
		if typeAlias != nil && typeAlias.Name != nil {
			return getIdentifierName(typeAlias.Name)
		}
	case ast.KindEnumDeclaration:
		enumDecl := node.AsEnumDeclaration()
		if enumDecl != nil && enumDecl.Name != nil {
			return getIdentifierName(enumDecl.Name)
		}
	case ast.KindEnumMember:
		member := node.AsEnumMember()
		if member != nil && member.Name != nil {
			return getIdentifierName(member.Name)
		}
	case ast.KindTypeParameter:
		typeParam := node.AsTypeParameterDeclaration()
		if typeParam != nil && typeParam.Name != nil {
			return getIdentifierName(typeParam.Name)
		}
	}

	return ""
}

// getIdentifierType returns a human-readable type description
func getIdentifierType(node *ast.Node) string {
	switch node.Kind {
	case ast.KindVariableDeclaration:
		return "Variable"
	case ast.KindParameter:
		if isParameterProperty(node) {
			return "Parameter Property"
		}
		return "Parameter"
	case ast.KindFunctionDeclaration:
		return "Function"
	case ast.KindPropertyDeclaration:
		if isClassProperty(node) {
			return "Class Property"
		}
		return "Property"
	case ast.KindMethodDeclaration:
		if isClassMethod(node) {
			return "Class Method"
		}
		return "Object Literal Method"
	case ast.KindGetAccessor:
		return "Accessor"
	case ast.KindSetAccessor:
		return "Accessor"
	case ast.KindClassDeclaration:
		return "Class"
	case ast.KindInterfaceDeclaration:
		return "Interface"
	case ast.KindTypeAliasDeclaration:
		return "Type alias"
	case ast.KindEnumDeclaration:
		return "Enum"
	case ast.KindEnumMember:
		return "Enum Member"
	case ast.KindTypeParameter:
		return "Type Parameter"
	case ast.KindPropertySignature:
		return "Type Property"
	case ast.KindMethodSignature:
		return "Type Method"
	}
	return "Identifier"
}

// matchesFilter checks if a name matches the filter criteria
func matchesFilter(name string, filter *FilterMatcher) bool {
	if filter == nil {
		return true
	}

	matched, _ := regexp.MatchString(filter.Regex, name)
	return matched == filter.Match
}

// findMatchingConfig finds the most specific matching configuration
func findMatchingConfig(node *ast.Node, configs []NamingConventionConfig, ctx rule.RuleContext) *NamingConventionConfig {
	name := getIdentifierName(node)
	if name == "" {
		return nil
	}

	// Iterate through configs to find the first matching one
	// Configs should be pre-sorted from most specific to least specific
	for i := range configs {
		config := &configs[i]

		// Check filter first
		if !matchesFilter(name, config.Filter) {
			continue
		}

		// Check if any selector matches
		selectorMatches := false
		for _, selector := range config.Selector {
			if matchesSelector(node, selector, config.Modifiers, ctx) {
				selectorMatches = true
				break
			}
		}

		if !selectorMatches {
			continue
		}

		// Check modifiers
		if !matchesModifiers(node, config.Modifiers) {
			continue
		}

		// TODO: Check types (requires type information from type checker)

		return config
	}

	return nil
}

// NamingConventionRule is the exported rule
var NamingConventionRule = rule.CreateRule(rule.Rule{
	Name: "naming-convention",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		configs := parseOptions(options)

		checkIdentifier := func(node *ast.Node) {
			name := getIdentifierName(node)
			if name == "" {
				return
			}

			// Find matching configuration
			config := findMatchingConfig(node, configs, ctx)
			if config == nil {
				return
			}

			// Validate the name
			valid, trimmedName := validateName(name, *config)
			if !valid {
				identType := getIdentifierType(node)

				// Build format list for error message
				formatList := ""
				for i, format := range config.Format {
					if i > 0 {
						formatList += ", "
					}
					formatList += string(format)
				}

				data := map[string]interface{}{
					"type":    identType,
					"name":    name,
					"formats": formatList,
				}

				// Use appropriate error message
				if trimmedName != name {
					ctx.ReportNode(node, buildDoesNotMatchFormatTrimmedMessage(), data)
				} else {
					ctx.ReportNode(node, buildDoesNotMatchFormatMessage(), data)
				}
			}
		}

		return rule.RuleListeners{
			ast.KindVariableDeclaration:    checkIdentifier,
			ast.KindParameter:              checkIdentifier,
			ast.KindFunctionDeclaration:    checkIdentifier,
			ast.KindPropertyDeclaration:    checkIdentifier,
			ast.KindPropertySignature:      checkIdentifier,
			ast.KindMethodDeclaration:      checkIdentifier,
			ast.KindMethodSignature:        checkIdentifier,
			ast.KindGetAccessor:            checkIdentifier,
			ast.KindSetAccessor:            checkIdentifier,
			ast.KindClassDeclaration:       checkIdentifier,
			ast.KindInterfaceDeclaration:   checkIdentifier,
			ast.KindTypeAliasDeclaration:   checkIdentifier,
			ast.KindEnumDeclaration:        checkIdentifier,
			ast.KindEnumMember:             checkIdentifier,
			ast.KindTypeParameter:          checkIdentifier,
			ast.KindPropertyAssignment:     checkIdentifier,
			ast.KindShorthandPropertyAssignment: checkIdentifier,
			ast.KindImportClause:           checkIdentifier,
			ast.KindImportSpecifier:        checkIdentifier,
			ast.KindNamespaceImport:        checkIdentifier,
		}
	},
})
