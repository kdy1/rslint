package ban_ts_comment

import (
	"regexp"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// BanTsCommentOptions defines configuration for the ban-ts-comment rule
type BanTsCommentOptions struct {
	TsExpectError          interface{} `json:"ts-expect-error"`
	TsIgnore               interface{} `json:"ts-ignore"`
	TsNocheck              interface{} `json:"ts-nocheck"`
	TsCheck                interface{} `json:"ts-check"`
	MinimumDescriptionLength int       `json:"minimumDescriptionLength"`
}

type directiveConfig struct {
	banned                   bool
	requireDescription       bool
	descriptionFormat        *regexp.Regexp
	minimumDescriptionLength int
}

func parseDirectiveOption(option interface{}, globalMinLength int) directiveConfig {
	if option == nil {
		return directiveConfig{banned: false}
	}

	switch v := option.(type) {
	case bool:
		if v {
			// true means completely banned
			return directiveConfig{banned: true}
		}
		// false means allowed without restriction
		return directiveConfig{banned: false}
	case string:
		if v == "allow-with-description" {
			return directiveConfig{
				banned:                   false,
				requireDescription:       true,
				minimumDescriptionLength: globalMinLength,
			}
		}
		// Default to banned if unrecognized string
		return directiveConfig{banned: true}
	case map[string]interface{}:
		config := directiveConfig{banned: false}

		if descRequired, ok := v["descriptionFormat"]; ok {
			if formatStr, ok := descRequired.(string); ok {
				if re, err := regexp.Compile(formatStr); err == nil {
					config.descriptionFormat = re
					config.requireDescription = true
				}
			}
		}

		return config
	default:
		return directiveConfig{banned: false}
	}
}

func parseOptions(options interface{}) BanTsCommentOptions {
	opts := BanTsCommentOptions{
		TsExpectError:          true, // Default: banned
		TsIgnore:               true, // Default: banned
		TsNocheck:              true, // Default: banned
		TsCheck:                false, // Default: allowed
		MinimumDescriptionLength: 3,
	}

	if options == nil {
		return opts
	}

	switch v := options.(type) {
	case map[string]interface{}:
		if val, ok := v["ts-expect-error"]; ok {
			opts.TsExpectError = val
		}
		if val, ok := v["ts-ignore"]; ok {
			opts.TsIgnore = val
		}
		if val, ok := v["ts-nocheck"]; ok {
			opts.TsNocheck = val
		}
		if val, ok := v["ts-check"]; ok {
			opts.TsCheck = val
		}
		if val, ok := v["minimumDescriptionLength"]; ok {
			if length, ok := val.(float64); ok {
				opts.MinimumDescriptionLength = int(length)
			}
		}
	}

	return opts
}

func buildBannedMessage(directive string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "tsDirectiveComment",
		Description: "Do not use \"@" + directive + "\" because it alters compilation errors.",
	}
}

func buildDescriptionNotMatchFormatMessage(directive string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "tsDirectiveCommentDescriptionNotMatchPattern",
		Description: "The description for the \"@" + directive + "\" directive must match the format.",
	}
}

func buildDescriptionTooShortMessage(directive string, minLength int) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "tsDirectiveCommentRequiresDescription",
		Description: "Include a description after the \"@" + directive + "\" directive to explain why the suppression is necessary.",
	}
}

var BanTsCommentRule = rule.CreateRule(rule.Rule{
	Name: "ban-ts-comment",
	Run: func(ctx rule.RuleContext, options interface{}) rule.RuleListeners {
		opts := parseOptions(options)

		directives := map[string]directiveConfig{
			"ts-expect-error": parseDirectiveOption(opts.TsExpectError, opts.MinimumDescriptionLength),
			"ts-ignore":       parseDirectiveOption(opts.TsIgnore, opts.MinimumDescriptionLength),
			"ts-nocheck":      parseDirectiveOption(opts.TsNocheck, opts.MinimumDescriptionLength),
			"ts-check":        parseDirectiveOption(opts.TsCheck, opts.MinimumDescriptionLength),
		}

		checkComment := func(commentText string, pos int, end int) {
			// Normalize comment text: remove leading // or /* and trailing */
			text := strings.TrimSpace(commentText)
			text = strings.TrimPrefix(text, "//")
			text = strings.TrimPrefix(text, "/*")
			text = strings.TrimSuffix(text, "*/")
			text = strings.TrimSpace(text)

			// Check if it starts with @ (TypeScript directive)
			if !strings.HasPrefix(text, "@") {
				return
			}

			// Extract directive name and description
			parts := strings.SplitN(text[1:], " ", 2) // Skip @ symbol
			directiveName := parts[0]
			description := ""
			if len(parts) > 1 {
				description = strings.TrimSpace(parts[1])
			}

			// Check if this is a directive we care about
			config, exists := directives[directiveName]
			if !exists {
				return
			}

			// If completely banned, report immediately
			if config.banned {
				ctx.ReportRange(core.NewTextRange(pos, end), buildBannedMessage(directiveName))
				return
			}

			// If description is required
			if config.requireDescription {
				// Check minimum length
				if len(description) < config.minimumDescriptionLength {
					ctx.ReportRange(core.NewTextRange(pos, end), buildDescriptionTooShortMessage(directiveName, config.minimumDescriptionLength))
					return
				}

				// Check description format if pattern is specified
				if config.descriptionFormat != nil && !config.descriptionFormat.MatchString(description) {
					ctx.ReportRange(core.NewTextRange(pos, end), buildDescriptionNotMatchFormatMessage(directiveName))
					return
				}
			}
		}

		return rule.RuleListeners{
			ast.KindSourceFile: func(node *ast.Node) {
				sourceFile := ctx.SourceFile
				text := sourceFile.Text()

				// Use ForEachComment to iterate over all comments in the file
				utils.ForEachComment(node, func(comment *ast.CommentRange) {
					// The scanner seems to return incorrect end positions, so calculate the correct end manually
					start := comment.Pos()
					end := start

					// Check if it's a // comment
					if end+1 < len(text) && text[end] == '/' && text[end+1] == '/' {
						// Find the end of the line
						end += 2
						for end < len(text) && text[end] != '\n' && text[end] != '\r' {
							end++
						}
					} else if end+1 < len(text) && text[end] == '/' && text[end+1] == '*' {
						// Find the closing */
						end += 2
						for end+1 < len(text) && !(text[end] == '*' && text[end+1] == '/') {
							end++
						}
						if end+1 < len(text) {
							end += 2
						}
					}

					commentText := text[start:end]
					checkComment(commentText, start, end)
				}, sourceFile)
			},
		}
	},
})
