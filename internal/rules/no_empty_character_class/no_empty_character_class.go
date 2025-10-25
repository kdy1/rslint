package no_empty_character_class

import (
	"regexp"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoEmptyCharacterClassRule implements the no-empty-character-class rule
// Disallow empty character classes in regex
var NoEmptyCharacterClassRule = rule.Rule{
	Name: "no-empty-character-class",
	Run:  run,
}

// Pattern to detect empty character classes in regex patterns
// This matches [] but not [^] (negated empty class is allowed in some contexts)
var emptyCharClassPattern = regexp.MustCompile(`(?:[^\\]|^)\[\]`)

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	return rule.RuleListeners{
		ast.KindRegularExpressionLiteral: func(node *ast.Node) {
			if !ast.IsRegularExpressionLiteral(node) {
				return
			}

			regexLiteral := node.AsRegularExpressionLiteral()
			if regexLiteral == nil {
				return
			}

			text := regexLiteral.Text
			if hasEmptyCharacterClass(text) {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "unexpected",
					Description: "Empty class.",
				})
			}
		},
	}
}

// hasEmptyCharacterClass checks if a regex pattern contains an empty character class
// Returns true if [] is found (but not [^])
func hasEmptyCharacterClass(pattern string) bool {
	// Remove the leading and trailing slashes and flags if present
	if len(pattern) < 2 {
		return false
	}

	// Track if we're inside a character class
	inClass := false
	escaped := false
	i := 0

	for i < len(pattern) {
		ch := pattern[i]

		if escaped {
			escaped = false
			i++
			continue
		}

		if ch == '\\' {
			escaped = true
			i++
			continue
		}

		if ch == '[' && !inClass {
			// Start of character class
			inClass = true
			// Check if next character is ]
			if i+1 < len(pattern) && pattern[i+1] == ']' {
				// Check if it's [^] (negated empty class - allowed)
				// We need to check before the [ for ^
				// Actually [^] means negated empty, which is checked as: [ followed by ^ followed by ]
				// Let me re-check: [^] is actually bracket, caret, bracket
				// So we check if the next char after [ is ]
				return true
			}
			// Check for [^] which is allowed
			if i+2 < len(pattern) && pattern[i+1] == '^' && pattern[i+2] == ']' {
				// Skip past this
				i += 3
				inClass = false
				continue
			}
		} else if ch == ']' && inClass {
			inClass = false
		}

		i++
	}

	return false
}
