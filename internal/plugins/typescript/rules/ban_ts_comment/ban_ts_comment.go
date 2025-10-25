package ban_ts_comment

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// DirectiveConfig defines the configuration for a ts-directive
type DirectiveConfig interface{}

const (
	DirectiveAllow                = false
	DirectiveBan                  = true
	DirectiveAllowWithDescription = "allow-with-description"
)

// BanTsCommentOptions defines the configuration options for this rule
type BanTsCommentOptions struct {
	TsExpectError           DirectiveConfig `json:"ts-expect-error"`
	TsIgnore                DirectiveConfig `json:"ts-ignore"`
	TsNocheck               DirectiveConfig `json:"ts-nocheck"`
	TsCheck                 DirectiveConfig `json:"ts-check"`
	MinimumDescriptionLength int            `json:"minimumDescriptionLength"`
	DescriptionFormat       string          `json:"descriptionFormat"`
}

// parseOptions parses and validates the rule options
func parseOptions(options any) BanTsCommentOptions {
	opts := BanTsCommentOptions{
		TsExpectError:           DirectiveAllowWithDescription,
		TsIgnore:                DirectiveBan,
		TsNocheck:               DirectiveAllowWithDescription,
		TsCheck:                 DirectiveAllow,
		MinimumDescriptionLength: 3,
	}

	if options == nil {
		return opts
	}

	var optsMap map[string]interface{}
	if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
		optsMap, _ = optArray[0].(map[string]interface{})
	} else {
		optsMap, _ = options.(map[string]interface{})
	}

	if optsMap != nil {
		if v, ok := optsMap["ts-expect-error"]; ok {
			opts.TsExpectError = v
		}
		if v, ok := optsMap["ts-ignore"]; ok {
			opts.TsIgnore = v
		}
		if v, ok := optsMap["ts-nocheck"]; ok {
			opts.TsNocheck = v
		}
		if v, ok := optsMap["ts-check"]; ok {
			opts.TsCheck = v
		}
		if v, ok := optsMap["minimumDescriptionLength"].(float64); ok {
			opts.MinimumDescriptionLength = int(v)
		}
		if v, ok := optsMap["descriptionFormat"].(string); ok {
			opts.DescriptionFormat = v
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

func buildDescriptionRequiredMessage(directive string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "tsDirectiveCommentRequiresDescription",
		Description: "Include a description after the \"@" + directive + "\" directive to explain why the @" + directive + " is necessary. The description must be " + "at least 3 characters long.",
	}
}

func buildDescriptionTooShortMessage(directive string, minLength int) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "tsDirectiveCommentDescriptionNotMatchPattern",
		Description: "The description for the \"@" + directive + "\" directive must be at least " + string(rune(minLength+'0')) + " characters long.",
	}
}

func buildDescriptionFormatMessage(directive string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "tsDirectiveCommentDescriptionNotMatchPattern",
		Description: "The description for the \"@" + directive + "\" directive does not match the required format.",
	}
}

func buildReplaceWithExpectErrorMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "replaceTsIgnoreWithTsExpectError",
		Description: "Replace \"@ts-ignore\" with \"@ts-expect-error\".",
	}
}

// countCharacters counts the actual number of characters including emojis
func countCharacters(s string) int {
	return utf8.RuneCountInString(s)
}

// checkDirective checks if a directive violates the rule
func checkDirective(ctx rule.RuleContext, opts BanTsCommentOptions, commentText string, commentRange core.TextRange, directive string, config DirectiveConfig) {
	directivePrefix := "@" + directive

	if !strings.Contains(commentText, directivePrefix) {
		return
	}

	// Find the position of the directive
	directiveIndex := strings.Index(commentText, directivePrefix)
	if directiveIndex == -1 {
		return
	}

	// Extract the part after the directive
	afterDirective := commentText[directiveIndex+len(directivePrefix):]

	// Get description (text after directive, trimmed)
	description := strings.TrimSpace(afterDirective)

	// Remove leading separators like : or -
	description = strings.TrimLeft(description, ":-")
	description = strings.TrimSpace(description)

	// Check if directive is banned
	if config == DirectiveBan || config == true {
		// For ts-ignore, suggest replacing with ts-expect-error
		if directive == "ts-ignore" {
			replacement := strings.Replace(commentText, "@ts-ignore", "@ts-expect-error", 1)
			ctx.ReportWithFixes(commentRange, buildBannedMessage(directive),
				rule.RuleFixReplaceRange(ctx.SourceFile, commentRange, replacement))
		} else {
			ctx.ReportWithFixes(commentRange, buildBannedMessage(directive))
		}
		return
	}

	// Check if description is required
	if config == DirectiveAllowWithDescription || config == "allow-with-description" {
		if description == "" {
			ctx.Report(commentRange, buildDescriptionRequiredMessage(directive))
			return
		}

		// Check minimum length
		descLength := countCharacters(description)
		if descLength < opts.MinimumDescriptionLength {
			ctx.Report(commentRange, buildDescriptionTooShortMessage(directive, opts.MinimumDescriptionLength))
			return
		}

		// Check format if specified
		if opts.DescriptionFormat != "" {
			matched, err := regexp.MatchString(opts.DescriptionFormat, description)
			if err == nil && !matched {
				ctx.Report(commentRange, buildDescriptionFormatMessage(directive))
				return
			}
		}
	}
}

// BanTsCommentRule implements the ban-ts-comment rule
// Disallow @ts-<directive> comments or require descriptions
var BanTsCommentRule = rule.CreateRule(rule.Rule{
	Name: "ban-ts-comment",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := parseOptions(options)

	return rule.RuleListeners{
		ast.KindSourceFile: func(node *ast.Node) {
			sourceFile := node.AsSourceFile()
			if sourceFile == nil {
				return
			}

			text := sourceFile.GetText()
			nodeFactory := ast.NewNodeFactory(ast.NodeFactoryHooks{})

			// Scan all comments in the source file
			pos := 0
			for {
				// Get leading comments at this position
				foundAny := false
				for commentRange := range scanner.GetLeadingCommentRanges(nodeFactory, text, pos) {
					foundAny = true
					commentText := text[commentRange.Pos():commentRange.End()]

					// Check each directive
					checkDirective(ctx, opts, commentText, commentRange, "ts-expect-error", opts.TsExpectError)
					checkDirective(ctx, opts, commentText, commentRange, "ts-ignore", opts.TsIgnore)
					checkDirective(ctx, opts, commentText, commentRange, "ts-nocheck", opts.TsNocheck)
					checkDirective(ctx, opts, commentText, commentRange, "ts-check", opts.TsCheck)

					pos = commentRange.End()
				}

				// Get trailing comments at this position
				for commentRange := range scanner.GetTrailingCommentRanges(nodeFactory, text, pos) {
					foundAny = true
					commentText := text[commentRange.Pos():commentRange.End()]

					// Check each directive
					checkDirective(ctx, opts, commentText, commentRange, "ts-expect-error", opts.TsExpectError)
					checkDirective(ctx, opts, commentText, commentRange, "ts-ignore", opts.TsIgnore)
					checkDirective(ctx, opts, commentText, commentRange, "ts-nocheck", opts.TsNocheck)
					checkDirective(ctx, opts, commentText, commentRange, "ts-check", opts.TsCheck)

					pos = commentRange.End()
				}

				if !foundAny {
					pos++
					if pos >= len(text) {
						break
					}
				}
			}
		},
	}
}
