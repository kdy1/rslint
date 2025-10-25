package no_misleading_character_class

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// NoMisleadingCharacterClassRule implements the no-misleading-character-class rule
// Disallow characters made with multiple code points in character class syntax
var NoMisleadingCharacterClassRule = rule.Rule{
	Name: "no-misleading-character-class",
	Run:  run,
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	return rule.RuleListeners{
		ast.KindRegularExpressionLiteral: func(node *ast.Node) {
			regexLiteral := node.AsRegularExpressionLiteral()
			if regexLiteral == nil {
				return
			}

			// Get the raw text of the regex literal
			rng := utils.TrimNodeTextRange(ctx.SourceFile, node)
			rawText := ctx.SourceFile.Text()[rng.Pos():rng.End()]

			// Extract pattern and flags from the regex literal
			pattern, flags := parseRegexLiteral(rawText)

			// Skip if using 'v' flag (unicodeSets mode in ES2024)
			if strings.Contains(flags, "v") {
				return
			}

			hasUFlag := strings.Contains(flags, "u")

			// Check for misleading character classes
			checkCharacterClasses(ctx, node, pattern, hasUFlag)
		},
	}
}

// parseRegexLiteral extracts the pattern and flags from a regex literal string
func parseRegexLiteral(text string) (pattern, flags string) {
	// Regex literal format: /pattern/flags
	if !strings.HasPrefix(text, "/") {
		return "", ""
	}

	// Find the last slash (before flags)
	lastSlash := strings.LastIndex(text, "/")
	if lastSlash <= 0 {
		return "", ""
	}

	pattern = text[1:lastSlash]
	if lastSlash+1 < len(text) {
		flags = text[lastSlash+1:]
	}

	return pattern, flags
}

// checkCharacterClasses checks for misleading multi-code-point characters in character classes
func checkCharacterClasses(ctx rule.RuleContext, node *ast.Node, pattern string, hasUFlag bool) {
	// Find all character classes in the pattern
	inClass := false
	escaped := false
	classStart := -1

	for i := 0; i < len(pattern); i++ {
		if escaped {
			escaped = false
			continue
		}

		if pattern[i] == '\\' {
			escaped = true
			continue
		}

		if pattern[i] == '[' && !inClass {
			inClass = true
			classStart = i
			continue
		}

		if pattern[i] == ']' && inClass {
			// Extract the character class content
			classContent := pattern[classStart+1 : i]
			checkClassContent(ctx, node, classContent, hasUFlag)
			inClass = false
			classStart = -1
		}
	}
}

// checkClassContent checks a single character class for misleading characters
func checkClassContent(ctx rule.RuleContext, node *ast.Node, content string, hasUFlag bool) {
	// Skip negated classes for simplicity (they start with ^)
	if strings.HasPrefix(content, "^") {
		content = content[1:]
	}

	// Check for surrogate pairs without u flag
	if !hasUFlag && hasSurrogatePair(content) {
		ctx.ReportNode(node, rule.RuleMessage{
			Id:          "surrogatePairWithoutUFlag",
			Description: "Unexpected surrogate pair in character class. Use 'u' flag.",
		})
		return
	}

	// Check for combining characters
	if hasCombiningCharacter(content) {
		ctx.ReportNode(node, rule.RuleMessage{
			Id:          "combiningClass",
			Description: "Unexpected combined character in character class.",
		})
		return
	}

	// Check for emoji modifiers
	if hasEmojiModifier(content) {
		ctx.ReportNode(node, rule.RuleMessage{
			Id:          "emojiModifier",
			Description: "Unexpected emoji modifier in character class.",
		})
		return
	}

	// Check for regional indicator symbols (flag emojis)
	if hasRegionalIndicator(content) {
		ctx.ReportNode(node, rule.RuleMessage{
			Id:          "regionalIndicatorSymbol",
			Description: "Unexpected regional indicator in character class.",
		})
		return
	}

	// Check for zero-width joiners
	if hasZWJ(content) {
		ctx.ReportNode(node, rule.RuleMessage{
			Id:          "zwj",
			Description: "Unexpected zero-width joiner in character class.",
		})
		return
	}
}

// hasSurrogatePair checks if the content contains surrogate pairs
func hasSurrogatePair(content string) bool {
	for i := 0; i < len(content); {
		r, size := utf8.DecodeRuneInString(content[i:])
		if r == utf8.RuneError {
			i++
			continue
		}
		// Check if this is a surrogate pair (runes > 0xFFFF)
		if r > 0xFFFF {
			return true
		}
		i += size
	}
	return false
}

// hasCombiningCharacter checks for combining marks (accents, diacritics, variation selectors)
func hasCombiningCharacter(content string) bool {
	// Combining marks are in the range U+0300 to U+036F and other ranges
	// Variation selectors: U+FE00 to U+FE0F, U+E0100 to U+E01EF
	for i := 0; i < len(content); {
		r, size := utf8.DecodeRuneInString(content[i:])
		if r == utf8.RuneError {
			i++
			continue
		}
		// Combining Diacritical Marks
		if r >= 0x0300 && r <= 0x036F {
			return true
		}
		// Variation Selectors
		if (r >= 0xFE00 && r <= 0xFE0F) || (r >= 0xE0100 && r <= 0xE01EF) {
			return true
		}
		i += size
	}
	return false
}

// hasEmojiModifier checks for emoji skin tone modifiers
func hasEmojiModifier(content string) bool {
	// Emoji modifiers are in the range U+1F3FB to U+1F3FF
	for i := 0; i < len(content); {
		r, size := utf8.DecodeRuneInString(content[i:])
		if r == utf8.RuneError {
			i++
			continue
		}
		if r >= 0x1F3FB && r <= 0x1F3FF {
			return true
		}
		i += size
	}
	return false
}

// hasRegionalIndicator checks for regional indicator symbols (flag emojis)
func hasRegionalIndicator(content string) bool {
	// Regional indicators are in the range U+1F1E6 to U+1F1FF
	count := 0
	for i := 0; i < len(content); {
		r, size := utf8.DecodeRuneInString(content[i:])
		if r == utf8.RuneError {
			i++
			continue
		}
		if r >= 0x1F1E6 && r <= 0x1F1FF {
			count++
			if count >= 2 {
				return true
			}
		}
		i += size
	}
	return false
}

// hasZWJ checks for zero-width joiners
func hasZWJ(content string) bool {
	// Zero-width joiner is U+200D
	return strings.ContainsRune(content, '\u200D')
}

// Helper: regex to extract character classes (simplified, may need refinement)
var charClassRegex = regexp.MustCompile(`\[([^\]]*)\]`)
