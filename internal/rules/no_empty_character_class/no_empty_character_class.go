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
// Returns true if [] is found (but not [^] which is allowed in most contexts)
func hasEmptyCharacterClass(pattern string) bool {
	if len(pattern) < 2 {
		return false
	}

	i := 0
	for i < len(pattern) {
		ch := pattern[i]

		// Handle escape sequences
		if ch == '\\' {
			// Skip the escaped character
			i += 2
			continue
		}

		// Check for start of character class
		if ch == '[' {
			i++ // Move past '['

			// Check for negation
			isNegated := false
			if i < len(pattern) && pattern[i] == '^' {
				isNegated = true
				i++
			}

			// Check if immediately followed by ]
			if i < len(pattern) && pattern[i] == ']' {
				// [^] is allowed in standard regex (matches any character except newline)
				// [] is empty and should be reported
				if !isNegated {
					return true
				}
				i++ // Skip the ]
				continue
			}

			// For v-flag (ES2024), check for nested character classes
			// which can also be empty like [[]] or [a&&[]]
			// This is a simplified check - we look for [[ or && patterns
			if i < len(pattern) {
				// Scan through the character class
				classDepth := 1
				for i < len(pattern) && classDepth > 0 {
					if pattern[i] == '\\' {
						i += 2
						continue
					}
					if pattern[i] == '[' {
						// Nested character class (v-flag feature)
						// Check if it's immediately empty [[]]
						if i+1 < len(pattern) && pattern[i+1] == ']' {
							return true
						}
						classDepth++
					} else if pattern[i] == ']' {
						classDepth--
						if classDepth == 0 {
							break
						}
						// Check if previous characters form an empty nested class
						// This is complex, so for now we do a simple check
					}
					// Check for && or -- operators with empty classes (v-flag)
					if i+1 < len(pattern) && (pattern[i] == '&' && pattern[i+1] == '&' ||
						pattern[i] == '-' && pattern[i+1] == '-') {
						// Check if followed by []
						j := i + 2
						for j < len(pattern) && pattern[j] == ' ' {
							j++
						}
						if j+1 < len(pattern) && pattern[j] == '[' && pattern[j+1] == ']' {
							return true
						}
					}
					i++
				}
			}
		}

		i++
	}

	return false
}
