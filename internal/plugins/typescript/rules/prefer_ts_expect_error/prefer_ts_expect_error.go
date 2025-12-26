package prefer_ts_expect_error

import (
	"regexp"
	"strings"

	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Regular expressions for matching @ts-ignore directives
var (
	// Matches single-line comments: // @ts-ignore or /// @ts-ignore
	singleLineTsIgnoreRegex = regexp.MustCompile(`^(\/\/\/?\s*)@ts-ignore\b`)

	// Matches multi-line comments: /* @ts-ignore */
	multiLineTsIgnoreRegex = regexp.MustCompile(`^(\/\*[\s*]*)@ts-ignore\b`)
)

// PreferTsExpectErrorRule implements the prefer-ts-expect-error rule
// Recommends using @ts-expect-error over @ts-ignore
var PreferTsExpectErrorRule = rule.CreateRule(rule.Rule{
	Name: "prefer-ts-expect-error",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	// Get the full text of the source file
	text := ctx.SourceFile.Text()

	// Process the text to find comments
	processComments(ctx, text)

	return rule.RuleListeners{}
}

// processComments scans the source text for comments and checks for @ts-ignore directives
func processComments(ctx rule.RuleContext, text string) {
	pos := 0
	length := len(text)

	for pos < length {
		// Skip to next potential comment
		if pos+1 < length {
			if text[pos] == '/' && text[pos+1] == '/' {
				// Single-line comment
				commentStart := pos
				pos += 2
				lineEnd := pos
				for lineEnd < length && text[lineEnd] != '\n' && text[lineEnd] != '\r' {
					lineEnd++
				}
				commentText := text[commentStart:lineEnd]
				checkComment(ctx, commentText, commentStart, false)
				pos = lineEnd
			} else if text[pos] == '/' && text[pos+1] == '*' {
				// Multi-line comment
				commentStart := pos
				pos += 2
				commentEnd := pos
				for commentEnd+1 < length {
					if text[commentEnd] == '*' && text[commentEnd+1] == '/' {
						commentEnd += 2
						break
					}
					commentEnd++
				}
				commentText := text[commentStart:commentEnd]
				checkComment(ctx, commentText, commentStart, true)
				pos = commentEnd
			} else {
				pos++
			}
		} else {
			pos++
		}
	}
}

// checkComment checks a single comment for @ts-ignore directives
func checkComment(ctx rule.RuleContext, commentText string, commentStart int, isMultiLine bool) {
	var match []string

	if isMultiLine {
		match = multiLineTsIgnoreRegex.FindStringSubmatch(commentText)
	} else {
		match = singleLineTsIgnoreRegex.FindStringSubmatch(commentText)
	}

	if match == nil {
		return
	}

	// For multi-line comments, check if there's meaningful content after @ts-ignore on subsequent lines
	// If there is, this is not a directive comment (it's just a comment that mentions @ts-ignore)
	if isMultiLine {
		// Extract the part after @ts-ignore
		idx := strings.Index(commentText, "@ts-ignore")
		if idx == -1 {
			return
		}

		afterDirective := commentText[idx+len("@ts-ignore"):]

		// Remove the trailing */
		withoutClosing := strings.TrimSuffix(afterDirective, "*/")

		// Find the first newline after the directive
		firstNewline := strings.Index(withoutClosing, "\n")
		if firstNewline != -1 {
			// Get content after the first newline
			afterFirstLine := withoutClosing[firstNewline+1:]

			// Check if there's any meaningful content after the directive line
			// (excluding whitespace and asterisks)
			lines := strings.Split(afterFirstLine, "\n")
			for _, line := range lines {
				trimmed := strings.TrimLeft(line, " \t*")
				trimmed = strings.TrimSpace(trimmed)
				// If this line has content, it's not a directive comment
				if len(trimmed) > 0 {
					return
				}
			}
		}
	}

	// Generate the fixed text by replacing @ts-ignore with @ts-expect-error
	fixedText := strings.Replace(commentText, "@ts-ignore", "@ts-expect-error", 1)

	// Report the issue with auto-fix
	ctx.ReportRangeWithFix(
		core.NewTextRange(commentStart, commentStart+len(commentText)),
		rule.RuleMessage{
			Id:          "preferExpectErrorComment",
			Description: "Use '@ts-expect-error' instead of '@ts-ignore', as '@ts-expect-error' will error if the following line is error-free.",
		},
		[]rule.Fix{
			{
				Range:   core.NewTextRange(commentStart, commentStart+len(commentText)),
				NewText: fixedText,
			},
		},
	)
}
