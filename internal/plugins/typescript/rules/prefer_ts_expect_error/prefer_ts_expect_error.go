package prefer_ts_expect_error

import (
	"regexp"
	"strings"

	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Regular expressions for matching @ts-ignore directives
var (
	// Matches @ts-ignore in any comment
	tsIgnoreRegex = regexp.MustCompile(`@ts-ignore\b`)
)

// PreferTsExpectErrorRule implements the prefer-ts-expect-error rule
// Enforces using @ts-expect-error over @ts-ignore
var PreferTsExpectErrorRule = rule.CreateRule(rule.Rule{
	Name: "prefer-ts-expect-error",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	// Get the full text of the source file
	text := ctx.SourceFile.Text()

	// Process the text to find comments with @ts-ignore
	processComments(ctx, text)

	return rule.RuleListeners{}
}

// processComments scans the source text for comments containing @ts-ignore
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
				checkComment(ctx, commentText, commentStart)
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
				checkComment(ctx, commentText, commentStart)
				pos = commentEnd
			} else {
				pos++
			}
		} else {
			pos++
		}
	}
}

// checkComment checks a single comment for @ts-ignore
func checkComment(ctx rule.RuleContext, commentText string, commentStart int) {
	// Check if the comment contains @ts-ignore
	if !tsIgnoreRegex.MatchString(commentText) {
		return
	}

	// Check if it's actually being used as a directive (not just mentioned in text)
	// A directive should have @ts-ignore at or near the beginning of the comment
	// For single-line: // @ts-ignore or /// @ts-ignore
	// For multi-line: /* @ts-ignore */

	// Strip comment markers to get the content
	content := commentText
	isSingleLine := strings.HasPrefix(content, "//")

	if isSingleLine {
		// Remove leading //
		content = strings.TrimPrefix(content, "//")
		// Also handle ///
		content = strings.TrimPrefix(content, "/")
	} else {
		// Remove leading /* and trailing */
		content = strings.TrimPrefix(content, "/*")
		content = strings.TrimSuffix(content, "*/")
	}

	// Trim leading whitespace and asterisks (for JSDoc-style comments)
	content = strings.TrimSpace(content)

	// For multi-line comments, also check if @ts-ignore is at the start of any line
	if !isSingleLine {
		// Check if @ts-ignore appears at the start (after trimming)
		// OR if it appears on its own line within the comment
		if !strings.HasPrefix(content, "@ts-ignore") {
			// Check if it appears on a separate line
			lines := strings.Split(content, "\n")
			foundDirective := false
			for _, line := range lines {
				trimmedLine := strings.TrimLeft(line, " \t*")
				trimmedLine = strings.TrimSpace(trimmedLine)
				// Check if this line starts with @ts-ignore (or has it after //)
				if strings.HasPrefix(trimmedLine, "@ts-ignore") ||
					(strings.HasPrefix(trimmedLine, "//") && strings.HasPrefix(strings.TrimSpace(strings.TrimPrefix(trimmedLine, "//")), "@ts-ignore")) {
					foundDirective = true
					break
				}
			}
			if !foundDirective {
				// @ts-ignore is mentioned but not as a directive
				return
			}
		}
	} else {
		// For single-line comments, check if @ts-ignore is at or near the start
		if !strings.HasPrefix(content, "@ts-ignore") {
			// @ts-ignore is mentioned but not as a directive
			return
		}
	}

	// Create a fix that replaces @ts-ignore with @ts-expect-error
	newText := strings.ReplaceAll(commentText, "@ts-ignore", "@ts-expect-error")
	fix := &rule.RuleFix{
		Range: core.NewTextRange(commentStart, commentStart+len(commentText)),
		Text:  newText,
	}

	ctx.ReportRangeWithFixes(
		core.NewTextRange(commentStart, commentStart+len(commentText)),
		rule.RuleMessage{
			Id:          "preferExpectErrorComment",
			Description: "Use '@ts-expect-error' instead of '@ts-ignore', as '@ts-expect-error' will error if the following line is error-free.",
		},
		*fix,
	)
}
