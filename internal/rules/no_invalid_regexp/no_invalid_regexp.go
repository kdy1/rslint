package no_invalid_regexp

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoInvalidRegexpOptions defines the configuration options for this rule
type NoInvalidRegexpOptions struct {
	AllowConstructorFlags []string `json:"allowConstructorFlags"`
}

// parseOptions parses and validates the rule options
func parseOptions(options any) NoInvalidRegexpOptions {
	opts := NoInvalidRegexpOptions{
		AllowConstructorFlags: []string{},
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
		if v, ok := optsMap["allowConstructorFlags"].([]interface{}); ok {
			for _, flag := range v {
				if flagStr, ok := flag.(string); ok {
					opts.AllowConstructorFlags = append(opts.AllowConstructorFlags, flagStr)
				}
			}
		}
	}

	return opts
}

func buildInvalidRegexpMessage(msg string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "regexMessage",
		Description: msg + ".",
	}
}

// isStringLiteral checks if a node is a string literal
func isStringLiteral(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return node.Kind == ast.KindStringLiteral || node.Kind == ast.KindNoSubstitutionTemplateLiteral
}

// getStringValue extracts the string value from a string literal node
func getStringValue(node *ast.Node, sourceFile *ast.SourceFile) (string, bool) {
	if !isStringLiteral(node) {
		return "", false
	}

	text := sourceFile.Text()
	start := node.Pos()
	end := node.End()

	if start >= end || end > len(text) {
		return "", false
	}

	rawText := text[start:end]

	// Remove quotes from string literals
	if len(rawText) >= 2 {
		quote := rawText[0]
		if (quote == '\'' || quote == '"' || quote == '`') && rawText[len(rawText)-1] == quote {
			return rawText[1 : len(rawText)-1], true
		}
	}

	return rawText, true
}

// validateRegExpFlags checks if the flags are valid
func validateRegExpFlags(flags string, allowedFlags map[rune]bool) string {
	if flags == "" {
		return ""
	}

	seen := make(map[rune]bool)
	var duplicates []string
	var invalid []string
	hasU := false
	hasV := false

	for _, ch := range flags {
		// Check for duplicates
		if seen[ch] {
			if !contains(duplicates, string(ch)) {
				duplicates = append(duplicates, string(ch))
			}
		}
		seen[ch] = true

		// Check if flag is valid
		if !allowedFlags[ch] {
			if !contains(invalid, string(ch)) {
				invalid = append(invalid, string(ch))
			}
		}

		// Track u and v flags
		if ch == 'u' {
			hasU = true
		}
		if ch == 'v' {
			hasV = true
		}
	}

	// Check for conflicting u and v flags
	if hasU && hasV {
		return "Regex 'u' and 'v' flags cannot be used together"
	}

	// Report duplicates
	if len(duplicates) > 0 {
		if len(duplicates) == 1 {
			return fmt.Sprintf("Duplicate flags supplied to RegExp constructor '%s'", duplicates[0])
		}
		return fmt.Sprintf("Duplicate flags supplied to RegExp constructor '%s'", strings.Join(duplicates, "', '"))
	}

	// Report invalid flags
	if len(invalid) > 0 {
		if len(invalid) == 1 {
			return fmt.Sprintf("Invalid flags supplied to RegExp constructor '%s'", invalid[0])
		}
		return fmt.Sprintf("Invalid flags supplied to RegExp constructor '%s'", strings.Join(invalid, ""))
	}

	return ""
}

// contains checks if a string slice contains a value
func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// validateRegExpPattern validates a regex pattern using Go's regexp package
// Returns an error message if invalid, empty string if valid
func validateRegExpPattern(pattern, flags string) string {
	// Convert JavaScript regex pattern to Go regex pattern
	// Note: This is a simplified conversion. JavaScript and Go regex have differences.
	goPattern := convertJSRegexToGo(pattern, flags)

	_, err := regexp.Compile(goPattern)
	if err != nil {
		return formatRegexpError(err, pattern, flags)
	}

	return ""
}

// convertJSRegexToGo converts JavaScript regex patterns to Go patterns
func convertJSRegexToGo(pattern, flags string) string {
	// For basic validation, we'll use the pattern mostly as-is
	// JavaScript features not in Go (lookbehind, named groups, etc.) need special handling

	// Add flags prefix for case-insensitive mode
	prefix := ""
	if strings.Contains(flags, "i") {
		prefix = "(?i)"
	}
	if strings.Contains(flags, "m") {
		prefix += "(?m)"
	}
	if strings.Contains(flags, "s") {
		prefix += "(?s)"
	}

	return prefix + pattern
}

// formatRegexpError formats regexp compilation errors into readable messages
func formatRegexpError(err error, pattern, flags string) string {
	errMsg := err.Error()

	// Map common Go regexp errors to ESLint-style messages
	if strings.Contains(errMsg, "missing closing ]") {
		return "Unterminated character class"
	}
	if strings.Contains(errMsg, "missing closing )") {
		return "Unmatched ')'"
	}
	if strings.Contains(errMsg, "unexpected )") {
		return "Unmatched ')'"
	}
	if strings.Contains(errMsg, "trailing backslash") {
		return "\\ at end of pattern"
	}
	if strings.Contains(errMsg, "nothing to repeat") {
		return "Nothing to repeat"
	}
	if strings.Contains(errMsg, "invalid escape") {
		return "Invalid escape"
	}
	if strings.Contains(errMsg, "invalid character class range") {
		return "Range out of order in character class"
	}

	// Return a generic error message with the pattern for unhandled cases
	return fmt.Sprintf("Invalid regular expression: %s", errMsg)
}

// validateSpecialJSFeatures checks for JavaScript-specific regex features
// Returns error message if found invalid patterns, empty string if valid
func validateSpecialJSFeatures(pattern, flags string) string {
	hasU := strings.Contains(flags, "u")
	hasV := strings.Contains(flags, "v")

	// Check for invalid escapes in unicode mode
	if hasU {
		// In unicode mode, \a is invalid (but valid in non-unicode mode)
		// Simple check for common invalid escapes
		i := 0
		for i < len(pattern) {
			if pattern[i] == '\\' && i+1 < len(pattern) {
				next := pattern[i+1]
				// Invalid single-letter escapes in unicode mode
				if next >= 'a' && next <= 'z' && !isValidEscape(next) {
					return "Invalid escape"
				}
				// Check for incomplete unicode escapes
				if next == 'u' {
					if i+2 < len(pattern) && pattern[i+2] == '{' {
						// \u{...} format - need closing }
						closeIdx := strings.Index(pattern[i+3:], "}")
						if closeIdx == -1 {
							return "Invalid Unicode escape sequence"
						}
					}
				}
				i += 2
			} else {
				i++
			}
		}
	}

	// ES2024 v-flag character class set operations
	if hasV {
		// With v flag, character class set operations are allowed: [A--B], [A&&B]
		// This is newer syntax that should be valid
	}

	// Check for trailing backslash
	if len(pattern) > 0 && pattern[len(pattern)-1] == '\\' {
		return "\\ at end of pattern"
	}

	// Check for unmatched brackets and parens
	if err := checkBalancedBrackets(pattern); err != "" {
		return err
	}

	return ""
}

// isValidEscape checks if a character is a valid escape sequence
func isValidEscape(ch byte) bool {
	// Valid escape characters in JavaScript regex
	validEscapes := "bBdDfnrsStvwW0123456789\\/.^$*+?()[]{}|"
	return strings.ContainsRune(validEscapes, rune(ch))
}

// checkBalancedBrackets checks for balanced brackets and parentheses
func checkBalancedBrackets(pattern string) string {
	charStack := 0
	parenStack := 0
	i := 0
	inCharClass := false

	for i < len(pattern) {
		ch := pattern[i]

		// Skip escaped characters
		if ch == '\\' && i+1 < len(pattern) {
			i += 2
			continue
		}

		if ch == '[' {
			if !inCharClass {
				inCharClass = true
				charStack++
			}
		} else if ch == ']' {
			if inCharClass {
				charStack--
				if charStack == 0 {
					inCharClass = false
				}
			}
		} else if ch == '(' && !inCharClass {
			parenStack++
		} else if ch == ')' && !inCharClass {
			parenStack--
			if parenStack < 0 {
				return "Unmatched ')'"
			}
		}

		i++
	}

	if charStack > 0 {
		return "Unterminated character class"
	}
	if parenStack > 0 {
		return "Unmatched '('"
	}

	return ""
}

// NoInvalidRegexpRule implements the no-invalid-regexp rule
// Disallow invalid regular expression strings in `RegExp` constructors
var NoInvalidRegexpRule = rule.CreateRule(rule.Rule{
	Name: "no-invalid-regexp",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := parseOptions(options)

	// Build allowed flags map
	// Standard flags: d, g, i, m, s, u, v, y
	allowedFlags := map[rune]bool{
		'd': true,
		'g': true,
		'i': true,
		'm': true,
		's': true,
		'u': true,
		'v': true,
		'y': true,
	}

	// Add custom allowed flags from options
	for _, flag := range opts.AllowConstructorFlags {
		for _, ch := range flag {
			if unicode.IsLetter(ch) || unicode.IsDigit(ch) {
				allowedFlags[ch] = true
			}
		}
	}

	checkRegExpCall := func(node *ast.Node) {
		if node == nil {
			return
		}

		var callee *ast.Node
		var args []*ast.Node

		// Handle both CallExpression and NewExpression
		if node.Kind == ast.KindCallExpression {
			callExpr := node.AsCallExpression()
			if callExpr == nil {
				return
			}
			callee = callExpr.Expression
			args = callExpr.Arguments()
		} else if node.Kind == ast.KindNewExpression {
			newExpr := node.AsNewExpression()
			if newExpr == nil {
				return
			}
			callee = newExpr.Expression
			args = newExpr.Arguments()
		} else {
			return
		}

		// Check if it's a RegExp call
		if callee == nil || callee.Kind != ast.KindIdentifier {
			return
		}

		if callee.Text() != "RegExp" {
			return
		}

		// Get pattern argument (first argument)
		var pattern string
		var patternNode *ast.Node
		if len(args) > 0 && isStringLiteral(args[0]) {
			var ok bool
			pattern, ok = getStringValue(args[0], ctx.SourceFile)
			if ok {
				patternNode = args[0]
			}
		}

		// Get flags argument (second argument)
		var flags string
		var flagsNode *ast.Node
		if len(args) > 1 && isStringLiteral(args[1]) {
			var ok bool
			flags, ok = getStringValue(args[1], ctx.SourceFile)
			if ok {
				flagsNode = args[1]
			}
		}

		// Validate flags first
		if flagsNode != nil {
			if flagsError := validateRegExpFlags(flags, allowedFlags); flagsError != "" {
				ctx.ReportNode(flagsNode, buildInvalidRegexpMessage(flagsError))
				return
			}
		}

		// Validate pattern if we have a static pattern
		if patternNode != nil {
			// Check for JavaScript-specific features first
			if err := validateSpecialJSFeatures(pattern, flags); err != "" {
				ctx.ReportNode(patternNode, buildInvalidRegexpMessage(err))
				return
			}

			// Then validate with Go's regexp
			if err := validateRegExpPattern(pattern, flags); err != "" {
				ctx.ReportNode(patternNode, buildInvalidRegexpMessage(err))
				return
			}
		}
	}

	return rule.RuleListeners{
		ast.KindCallExpression: checkRegExpCall,
		ast.KindNewExpression:  checkRegExpCall,
	}
}
