package ban_tslint_comment

import (
	"regexp"

	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Regular expressions for matching TSLint directives
var (
	// Matches single-line comments: // tslint:disable, // tslint:enable, // tslint:disable-next-line, // tslint:disable-line
	singleLineTslintRegex = regexp.MustCompile(`^\/\/\s*tslint:(disable|enable|disable-next-line|disable-line)`)

	// Matches multi-line comments: /* tslint:disable */, /* tslint:enable */
	multiLineTslintRegex = regexp.MustCompile(`^\/\*\s*tslint:(disable|enable)`)
)

// BanTslintCommentRule implements the ban-tslint-comment rule
// Bans tslint:<directive> comments
var BanTslintCommentRule = rule.CreateRule(rule.Rule{
	Name: "ban-tslint-comment",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	// Get the full text of the source file
	text := ctx.SourceFile.Text()

	// Process the text to find tslint comments
	processComments(ctx, text)

	return rule.RuleListeners{}
}

// processComments scans the source text for comments and checks for tslint directives
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

				// Find the start of the line
				lineStart := commentStart
				for lineStart > 0 && text[lineStart-1] != '\n' && text[lineStart-1] != '\r' {
					lineStart--
				}

				commentText := text[commentStart:lineEnd]
				checkComment(ctx, text, commentText, commentStart, lineStart, lineEnd, false)
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
				checkComment(ctx, text, commentText, commentStart, commentStart, commentEnd, true)
				pos = commentEnd
			} else {
				pos++
			}
		} else {
			pos++
		}
	}
}

// checkComment checks a single comment for tslint directives
func checkComment(ctx rule.RuleContext, fullText string, commentText string, commentStart int, lineStart int, commentEnd int, isMultiLine bool) {
	var matches []string

	if isMultiLine {
		match := multiLineTslintRegex.FindStringSubmatch(commentText)
		if match != nil {
			matches = match
		}
	} else {
		match := singleLineTslintRegex.FindStringSubmatch(commentText)
		if match != nil {
			matches = match
		}
	}

	if matches == nil {
		return
	}

	// Determine what to remove
	// Check if there's code before the comment on the same line
	lineBeforeComment := fullText[lineStart:commentStart]
	hasCodeBefore := false
	for _, ch := range lineBeforeComment {
		if ch != ' ' && ch != '\t' {
			hasCodeBefore = true
			break
		}
	}

	var fix rule.RuleFix
	if hasCodeBefore {
		// For inline comments, only remove the comment (starting with optional space before comment)
		// Find the last non-whitespace character before the comment
		removeStart := commentStart
		for removeStart > lineStart && (fullText[removeStart-1] == ' ' || fullText[removeStart-1] == '\t') {
			removeStart--
		}
		fix = rule.RuleFixRemoveRange(core.NewTextRange(removeStart, commentEnd))
	} else {
		// For standalone comments, remove the whole line including newline
		removeEnd := commentEnd
		// Include trailing newlines
		if removeEnd < len(fullText) && (fullText[removeEnd] == '\n' || fullText[removeEnd] == '\r') {
			removeEnd++
			// Handle \r\n
			if removeEnd < len(fullText) && fullText[removeEnd-1] == '\r' && fullText[removeEnd] == '\n' {
				removeEnd++
			}
		}
		fix = rule.RuleFixRemoveRange(core.NewTextRange(lineStart, removeEnd))
	}

	// Create the diagnostic with fix
	ctx.ReportRangeWithFixes(
		core.NewTextRange(commentStart, commentEnd),
		rule.RuleMessage{
			Id:          "commentDetected",
			Description: "tslint is deprecated. Please remove this comment.",
		},
		fix,
	)
}
