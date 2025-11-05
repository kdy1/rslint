package naming_convention

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NamingConventionRule enforces naming conventions for various identifiers
var NamingConventionRule = rule.CreateRule(rule.Rule{
	Name: "naming-convention",
	Run:  run,
})

type SelectorConfig struct {
	Selector           interface{} // string or []string
	Format             []string
	LeadingUnderscore  string
	TrailingUnderscore string
	Prefix             []string
	Suffix             []string
	Filter             *FilterConfig
}

type FilterConfig struct {
	Regex string
	Match bool
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	// Parse options
	var configs []SelectorConfig
	if optSlice, ok := options.([]interface{}); ok && len(optSlice) > 0 {
		if configMap, ok := optSlice[0].(map[string]interface{}); ok {
			// Single config object
			configs = append(configs, parseConfig(configMap))
		} else if configSlice, ok := optSlice[0].([]interface{}); ok {
			// Array of config objects
			for _, cfg := range configSlice {
				if cfgMap, ok := cfg.(map[string]interface{}); ok {
					configs = append(configs, parseConfig(cfgMap))
				}
			}
		}
	}

	// Helper function to check if a name matches the format
	checkName := func(name string, config SelectorConfig, node *ast.Node) {
		if name == "" {
			return
		}

		// Apply filter if present
		if config.Filter != nil && config.Filter.Regex != "" {
			matched, _ := regexp.MatchString(config.Filter.Regex, name)
			if config.Filter.Match && !matched {
				return
			}
			if !config.Filter.Match && matched {
				return
			}
		}

		originalName := name

		// Check leading underscore
		leadingUnderscores := 0
		for len(name) > 0 && name[0] == '_' {
			leadingUnderscores++
			name = name[1:]
		}

		if config.LeadingUnderscore == "forbid" && leadingUnderscores > 0 {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "unexpectedUnderscore",
				Description: fmt.Sprintf("Unexpected leading underscore in name %q.", originalName),
			})
			return
		}

		if config.LeadingUnderscore == "require" && leadingUnderscores != 1 {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "missingUnderscore",
				Description: fmt.Sprintf("Name %q must have one leading underscore.", originalName),
			})
			return
		}

		if config.LeadingUnderscore == "requireDouble" && leadingUnderscores != 2 {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "missingUnderscore",
				Description: fmt.Sprintf("Name %q must have two leading underscores.", originalName),
			})
			return
		}

		// Check trailing underscore
		trailingUnderscores := 0
		for len(name) > 0 && name[len(name)-1] == '_' {
			trailingUnderscores++
			name = name[:len(name)-1]
		}

		if config.TrailingUnderscore == "forbid" && trailingUnderscores > 0 {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "unexpectedUnderscore",
				Description: fmt.Sprintf("Unexpected trailing underscore in name %q.", originalName),
			})
			return
		}

		if config.TrailingUnderscore == "require" && trailingUnderscores != 1 {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "missingUnderscore",
				Description: fmt.Sprintf("Name %q must have one trailing underscore.", originalName),
			})
			return
		}

		if config.TrailingUnderscore == "requireDouble" && trailingUnderscores != 2 {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "missingUnderscore",
				Description: fmt.Sprintf("Name %q must have two trailing underscores.", originalName),
			})
			return
		}

		// Check prefix
		if len(config.Prefix) > 0 {
			hasPrefix := false
			for _, prefix := range config.Prefix {
				if strings.HasPrefix(name, prefix) {
					name = name[len(prefix):]
					hasPrefix = true
					break
				}
			}
			if !hasPrefix {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "missingAffix",
					Description: fmt.Sprintf("Name %q must have prefix %v.", originalName, config.Prefix),
				})
				return
			}
		}

		// Check suffix
		if len(config.Suffix) > 0 {
			hasSuffix := false
			for _, suffix := range config.Suffix {
				if strings.HasSuffix(name, suffix) {
					name = name[:len(name)-len(suffix)]
					hasSuffix = true
					break
				}
			}
			if !hasSuffix {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "missingAffix",
					Description: fmt.Sprintf("Name %q must have suffix %v.", originalName, config.Suffix),
				})
				return
			}
		}

		// Check format
		if len(config.Format) > 0 {
			matched := false
			for _, format := range config.Format {
				if matchesFormat(name, format) {
					matched = true
					break
				}
			}
			if !matched {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "doesNotMatchFormat",
					Description: fmt.Sprintf("Name %q does not match format %v.", originalName, config.Format),
				})
			}
		}
	}

	// Helper function to check enum declarations
	checkEnumDeclaration := func(node *ast.Node) {
		enumDecl := node.AsEnumDeclaration()
		if enumDecl == nil || enumDecl.Name() == nil {
			return
		}

		name := getNodeText(enumDecl.Name())

		// Find applicable configs
		for _, config := range configs {
			if matchesSelector(config.Selector, "enum") {
				checkName(name, config, enumDecl.Name())
			}
		}
	}

	return rule.RuleListeners{
		ast.KindEnumDeclaration: checkEnumDeclaration,
	}
}

// Helper function to parse config from map
func parseConfig(configMap map[string]interface{}) SelectorConfig {
	config := SelectorConfig{}

	if selector, ok := configMap["selector"]; ok {
		config.Selector = selector
	}

	if format, ok := configMap["format"].([]interface{}); ok {
		for _, f := range format {
			if fStr, ok := f.(string); ok {
				config.Format = append(config.Format, fStr)
			}
		}
	}

	if leadingUnderscore, ok := configMap["leadingUnderscore"].(string); ok {
		config.LeadingUnderscore = leadingUnderscore
	}

	if trailingUnderscore, ok := configMap["trailingUnderscore"].(string); ok {
		config.TrailingUnderscore = trailingUnderscore
	}

	if prefix, ok := configMap["prefix"].([]interface{}); ok {
		for _, p := range prefix {
			if pStr, ok := p.(string); ok {
				config.Prefix = append(config.Prefix, pStr)
			}
		}
	}

	if suffix, ok := configMap["suffix"].([]interface{}); ok {
		for _, s := range suffix {
			if sStr, ok := s.(string); ok {
				config.Suffix = append(config.Suffix, sStr)
			}
		}
	}

	if filter, ok := configMap["filter"].(map[string]interface{}); ok {
		config.Filter = &FilterConfig{}
		if regex, ok := filter["regex"].(string); ok {
			config.Filter.Regex = regex
		}
		if match, ok := filter["match"].(bool); ok {
			config.Filter.Match = match
		}
	}

	return config
}

// Helper function to check if selector matches
func matchesSelector(selector interface{}, target string) bool {
	if selectorStr, ok := selector.(string); ok {
		return selectorStr == target
	}
	if selectorSlice, ok := selector.([]string); ok {
		for _, s := range selectorSlice {
			if s == target {
				return true
			}
		}
	}
	if selectorSlice, ok := selector.([]interface{}); ok {
		for _, s := range selectorSlice {
			if sStr, ok := s.(string); ok && sStr == target {
				return true
			}
		}
	}
	return false
}

// Helper function to check if name matches format
func matchesFormat(name string, format string) bool {
	switch format {
	case "camelCase":
		return isCamelCase(name)
	case "strictCamelCase":
		return isStrictCamelCase(name)
	case "PascalCase":
		return isPascalCase(name)
	case "StrictPascalCase":
		return isStrictPascalCase(name)
	case "snake_case":
		return isSnakeCase(name)
	case "UPPER_CASE":
		return isUpperCase(name)
	default:
		return true
	}
}

// Format validation functions
func isCamelCase(name string) bool {
	if len(name) == 0 {
		return false
	}
	// Must start with lowercase letter
	if !unicode.IsLower(rune(name[0])) {
		return false
	}
	// camelCase allows mixed case (like camelCaseUNSTRICT) but not underscores in non-uppercase parts
	// Reject names with underscore not followed by uppercase
	for i := 0; i < len(name); i++ {
		if name[i] == '_' {
			// Check if we're in an all-uppercase sequence (like UNSTRICT in camelCaseUNSTRICT)
			// This is allowed, otherwise underscore is not allowed
			if i > 0 && !unicode.IsUpper(rune(name[i-1])) {
				return false
			}
			if i+1 < len(name) && !unicode.IsUpper(rune(name[i+1])) {
				return false
			}
		}
	}
	return true
}

func isStrictCamelCase(name string) bool {
	if len(name) == 0 {
		return false
	}
	// Must start with lowercase letter
	if !unicode.IsLower(rune(name[0])) {
		return false
	}
	// Must not have consecutive uppercase letters (except at boundaries)
	// Must not contain underscores
	if strings.Contains(name, "_") {
		return false
	}
	return !startsWithUpperCase(name) && !hasConsecutiveUppercase(name)
}

func isPascalCase(name string) bool {
	if len(name) == 0 {
		return false
	}
	// Must start with uppercase letter
	if !unicode.IsUpper(rune(name[0])) {
		return false
	}
	// Allow mixed case and some uppercase sequences
	return true
}

func isStrictPascalCase(name string) bool {
	if len(name) == 0 {
		return false
	}
	// Must start with uppercase letter
	if !unicode.IsUpper(rune(name[0])) {
		return false
	}
	// Must not have consecutive uppercase letters (except acronyms like I18n)
	// Must not contain underscores
	if strings.Contains(name, "_") {
		return false
	}
	return !hasConsecutiveUppercase(name)
}

func isSnakeCase(name string) bool {
	if len(name) == 0 {
		return false
	}
	// Must be all lowercase with underscores
	// Must not have consecutive underscores
	// Must not start or end with underscore (those are handled separately)
	for i, ch := range name {
		if !unicode.IsLower(ch) && !unicode.IsDigit(ch) && ch != '_' {
			return false
		}
		if ch == '_' && i > 0 && name[i-1] == '_' {
			return false
		}
	}
	return true
}

func isUpperCase(name string) bool {
	if len(name) == 0 {
		return false
	}
	// Must be all uppercase with underscores
	for _, ch := range name {
		if !unicode.IsUpper(ch) && !unicode.IsDigit(ch) && ch != '_' {
			return false
		}
	}
	return true
}

func startsWithUpperCase(name string) bool {
	if len(name) == 0 {
		return false
	}
	return unicode.IsUpper(rune(name[0]))
}

func hasConsecutiveUppercase(name string) bool {
	prevWasUpper := false
	for _, ch := range name {
		if unicode.IsUpper(ch) {
			if prevWasUpper {
				return true
			}
			prevWasUpper = true
		} else {
			prevWasUpper = false
		}
	}
	return false
}

// Helper function to get node text
func getNodeText(node *ast.Node) string {
	if node == nil {
		return ""
	}

	switch node.Kind {
	case ast.KindIdentifier:
		ident := node.AsIdentifier()
		if ident != nil {
			return ident.Text
		}
	}

	return ""
}
