package no_invalid_regexp

import (
	"regexp"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"

	"github.com/web-infra-dev/rslint/internal/rule"
)

// Options mirrors ESLint no-invalid-regexp options
type Options struct {
	AllowConstructorFlags []string `json:"allowConstructorFlags"`
}

func parseOptions(options any) Options {
	opts := Options{
		AllowConstructorFlags: []string{},
	}

	if options == nil {
		return opts
	}

	// Parse options with dual-format support
	var optsMap map[string]interface{}
	var ok bool

	// Handle array format: [{ allowConstructorFlags: ["u", "y"] }]
	if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
		optsMap, ok = optArray[0].(map[string]interface{})
	} else {
		// Handle direct object format: { allowConstructorFlags: ["u", "y"] }
		optsMap, ok = options.(map[string]interface{})
	}

	if ok {
		if v, ok := optsMap["allowConstructorFlags"].([]interface{}); ok {
			for _, flag := range v {
				if s, ok := flag.(string); ok {
					opts.AllowConstructorFlags = append(opts.AllowConstructorFlags, s)
				}
			}
		}
	}

	return opts
}

func buildRegexMessage(message string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "regexMessage",
		Description: message,
	}
}

// Standard ECMAScript RegExp flags
var standardFlags = map[rune]bool{
	'g': true, // global
	'i': true, // ignoreCase
	'm': true, // multiline
	's': true, // dotAll (ES2018)
	'u': true, // unicode (ES2015)
	'y': true, // sticky (ES2015)
	'd': true, // hasIndices (ES2022)
	'v': true, // unicodeSets (ES2024)
}

func validateRegExpFlags(flags string, allowConstructorFlags []string) string {
	if flags == "" {
		return ""
	}

	seen := make(map[rune]bool)
	hasU := false
	hasV := false

	// Create allowed flags map
	allowedFlags := make(map[rune]bool)
	for k, v := range standardFlags {
		allowedFlags[k] = v
	}
	for _, flag := range allowConstructorFlags {
		if len(flag) == 1 {
			allowedFlags[rune(flag[0])] = true
		}
	}

	for _, flag := range flags {
		// Check if flag is allowed
		if !allowedFlags[flag] {
			return "Invalid flags supplied to RegExp constructor '" + string(flag) + "'"
		}

		// Check for duplicates
		if seen[flag] {
			return "Duplicate flags supplied to RegExp constructor '" + string(flag) + "'"
		}
		seen[flag] = true

		if flag == 'u' {
			hasU = true
		}
		if flag == 'v' {
			hasV = true
		}
	}

	// Check for conflicting u and v flags
	if hasU && hasV {
		return "The 'u' and 'v' flags cannot be used together"
	}

	return ""
}

func validateRegExpPattern(pattern string) string {
	// Basic validation for common errors
	if strings.HasSuffix(pattern, "\\") && !strings.HasSuffix(pattern, "\\\\") {
		return "Invalid regular expression: \\ at end of pattern"
	}

	// Check for unmatched brackets
	bracketDepth := 0
	parenDepth := 0
	escaped := false

	for _, ch := range pattern {
		if escaped {
			escaped = false
			continue
		}

		if ch == '\\' {
			escaped = true
			continue
		}

		switch ch {
		case '[':
			bracketDepth++
		case ']':
			bracketDepth--
			if bracketDepth < 0 {
				return "Invalid regular expression: Unmatched ']'"
			}
		case '(':
			parenDepth++
		case ')':
			parenDepth--
			if parenDepth < 0 {
				return "Invalid regular expression: Unmatched ')'"
			}
		}
	}

	if bracketDepth > 0 {
		return "Invalid regular expression: Unterminated character class"
	}
	if parenDepth > 0 {
		return "Invalid regular expression: Unterminated group"
	}

	// Try to compile with Go's regexp package (limited validation)
	_, err := regexp.Compile(pattern)
	if err != nil {
		return "Invalid regular expression: " + err.Error()
	}

	return ""
}

func extractStringLiteral(node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	if node.Kind == ast.KindStringLiteral {
		str := node.AsStringLiteral()
		if str != nil {
			return str.Text, true
		}
	}
	return "", false
}

// NoInvalidRegexpRule implements the no-invalid-regexp rule
var NoInvalidRegexpRule = rule.CreateRule(rule.Rule{
	Name: "no-invalid-regexp",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		checkRegExpCall := func(node *ast.Node, args []*ast.Node) {
			if args == nil || len(args) == 0 {
				return
			}

			// First argument is the pattern
			pattern, ok := extractStringLiteral(args[0])
			if !ok {
				// Non-literal pattern, can't validate
				return
			}

			// Second argument is flags (optional)
			flags := ""
			if len(args) > 1 {
				flags, ok = extractStringLiteral(args[1])
				if !ok {
					// Non-literal flags, can't validate
					return
				}
			}

			// Validate flags first
			if flagError := validateRegExpFlags(flags, opts.AllowConstructorFlags); flagError != "" {
				ctx.ReportNode(node, buildRegexMessage(flagError))
				return
			}

			// Validate pattern
			if patternError := validateRegExpPattern(pattern); patternError != "" {
				ctx.ReportNode(node, buildRegexMessage(patternError))
			}
		}

		return rule.RuleListeners{
			ast.KindNewExpression: func(node *ast.Node) {
				newExpr := node.AsNewExpression()
				if newExpr == nil || newExpr.Expression == nil {
					return
				}

				// Check if this is a RegExp constructor call
				if newExpr.Expression.Kind != ast.KindIdentifier {
					return
				}
				id := newExpr.Expression.AsIdentifier()
				if id == nil || id.Text != "RegExp" {
					return
				}

				var args []*ast.Node
				if newExpr.Arguments != nil {
					args = newExpr.Arguments.Nodes
				}
				checkRegExpCall(node, args)
			},

			ast.KindCallExpression: func(node *ast.Node) {
				callExpr := node.AsCallExpression()
				if callExpr == nil || callExpr.Expression == nil {
					return
				}

				// Check if this is a RegExp() call (without new)
				if callExpr.Expression.Kind != ast.KindIdentifier {
					return
				}
				id := callExpr.Expression.AsIdentifier()
				if id == nil || id.Text != "RegExp" {
					return
				}

				var args []*ast.Node
				if callExpr.Arguments != nil {
					args = callExpr.Arguments.Nodes
				}
				checkRegExpCall(node, args)
			},
		}
	},
})
