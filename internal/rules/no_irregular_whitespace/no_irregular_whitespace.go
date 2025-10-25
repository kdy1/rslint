package no_irregular_whitespace

import (
	"github.com/microsoft/typescript-go/shim/ast"

	"github.com/web-infra-dev/rslint/internal/rule"
)

// Options mirrors ESLint no-irregular-whitespace options
type Options struct {
	SkipComments  bool `json:"skipComments"`
	SkipStrings   bool `json:"skipStrings"`
	SkipTemplates bool `json:"skipTemplates"`
	SkipRegExps   bool `json:"skipRegExps"`
	SkipJSXText   bool `json:"skipJSXText"`
}

func parseOptions(options any) Options {
	opts := Options{
		SkipComments:  false,
		SkipStrings:   true, // default: skip strings
		SkipTemplates: false,
		SkipRegExps:   false,
		SkipJSXText:   false,
	}

	if options == nil {
		return opts
	}

	// Parse options with dual-format support
	var optsMap map[string]interface{}
	var ok bool

	// Handle array format: [{ skipComments: true }]
	if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
		optsMap, ok = optArray[0].(map[string]interface{})
	} else {
		// Handle direct object format: { skipComments: true }
		optsMap, ok = options.(map[string]interface{})
	}

	if ok {
		if v, ok := optsMap["skipComments"].(bool); ok {
			opts.SkipComments = v
		}
		if v, ok := optsMap["skipStrings"].(bool); ok {
			opts.SkipStrings = v
		}
		if v, ok := optsMap["skipTemplates"].(bool); ok {
			opts.SkipTemplates = v
		}
		if v, ok := optsMap["skipRegExps"].(bool); ok {
			opts.SkipRegExps = v
		}
		if v, ok := optsMap["skipJSXText"].(bool); ok {
			opts.SkipJSXText = v
		}
	}

	return opts
}

func buildNoIrregularWhitespaceMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noIrregularWhitespace",
		Description: "Irregular whitespace not allowed.",
	}
}

// Irregular whitespace characters
var irregularWhitespace = map[rune]bool{
	'\u000B': true, // LINE TABULATION
	'\u000C': true, // FORM FEED
	'\u0085': true, // NEXT LINE
	'\u00A0': true, // NO-BREAK SPACE
	'\u180E': true, // MONGOLIAN VOWEL SEPARATOR
	'\uFEFF': true, // ZERO WIDTH NO-BREAK SPACE (BOM)
	'\u2000': true, // EN QUAD
	'\u2001': true, // EM QUAD
	'\u2002': true, // EN SPACE
	'\u2003': true, // EM SPACE
	'\u2004': true, // THREE-PER-EM SPACE
	'\u2005': true, // FOUR-PER-EM SPACE
	'\u2006': true, // SIX-PER-EM SPACE
	'\u2007': true, // FIGURE SPACE
	'\u2008': true, // PUNCTUATION SPACE
	'\u2009': true, // THIN SPACE
	'\u200A': true, // HAIR SPACE
	'\u200B': true, // ZERO WIDTH SPACE
	'\u202F': true, // NARROW NO-BREAK SPACE
	'\u205F': true, // MEDIUM MATHEMATICAL SPACE
	'\u2028': true, // LINE SEPARATOR
	'\u2029': true, // PARAGRAPH SEPARATOR
	'\u3000': true, // IDEOGRAPHIC SPACE
}

func hasIrregularWhitespace(text string) bool {
	for _, ch := range text {
		if irregularWhitespace[ch] {
			return true
		}
	}
	return false
}

// NoIrregularWhitespaceRule implements the no-irregular-whitespace rule
var NoIrregularWhitespaceRule = rule.CreateRule(rule.Rule{
	Name: "no-irregular-whitespace",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		checkText := func(node *ast.Node, text string) {
			if text == "" {
				return
			}

			if hasIrregularWhitespace(text) {
				ctx.ReportNode(node, buildNoIrregularWhitespaceMessage())
			}
		}

		return rule.RuleListeners{
			// Check identifiers
			ast.KindIdentifier: func(node *ast.Node) {
				id := node.AsIdentifier()
				if id != nil {
					checkText(node, id.Text)
				}
			},

			// Check string literals (unless skipStrings is true)
			ast.KindStringLiteral: func(node *ast.Node) {
				if opts.SkipStrings {
					return
				}
				str := node.AsStringLiteral()
				if str != nil {
					checkText(node, str.Text)
				}
			},

			// Check no substitution templates (unless skipTemplates is true)
			ast.KindNoSubstitutionTemplateLiteral: func(node *ast.Node) {
				if opts.SkipTemplates {
					return
				}
				tpl := node.AsNoSubstitutionTemplateLiteral()
				if tpl != nil {
					checkText(node, tpl.Text)
				}
			},

			// Check regular expressions (unless skipRegExps is true)
			ast.KindRegularExpressionLiteral: func(node *ast.Node) {
				if opts.SkipRegExps {
					return
				}
				re := node.AsRegularExpressionLiteral()
				if re != nil {
					checkText(node, re.Text)
				}
			},
		}
	},
})
