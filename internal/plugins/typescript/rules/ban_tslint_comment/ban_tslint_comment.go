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

// checkComment checks a single comment for tslint directives
func checkComment(ctx rule.RuleContext, commentText string, commentStart int, isMultiLine bool) {
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

	// Calculate the end position
	commentEnd := commentStart + len(commentText)

	// Create the diagnostic
	ctx.ReportRange(
		core.NewTextRange(commentStart, commentEnd),
		rule.RuleMessage{
			Id:          "commentDetected",
			Description: "tslint is deprecated. Please remove this comment.",
		},
	)
}
