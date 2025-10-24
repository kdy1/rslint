package unicode_bom

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
)

const BOM = '\uFEFF'

// Options for unicode-bom rule
type Options struct {
	Mode string `json:"mode"` // "always" or "never"
}

func parseOptions(options any) Options {
	opts := Options{
		Mode: "never", // default
	}

	if options == nil {
		return opts
	}

	// Handle array format: ["always"] or ["never"]
	if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
		if mode, ok := optArray[0].(string); ok {
			opts.Mode = mode
		}
	}

	return opts
}

func buildExpectedBOMMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "expected",
		Description: "Expected Unicode BOM (Byte Order Mark).",
	}
}

func buildUnexpectedBOMMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpected",
		Description: "Unexpected Unicode BOM (Byte Order Mark).",
	}
}

// UnicodeBomRule enforces the presence or absence of a Unicode BOM at the start of files.
var UnicodeBomRule = rule.Rule{
	Name: "unicode-bom",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)
		text := ctx.SourceFile.Text()

		// Check if file starts with BOM (UTF-8 BOM is 0xEF 0xBB 0xBF - 3 bytes)
		hasBOM := len(text) >= 3 && text[0] == 0xEF && text[1] == 0xBB && text[2] == 0xBF

		// Create a synthetic node at position 0 for reporting
		// This is a workaround since ReportRange doesn't support fixes
		syntheticNode := &ast.Node{
			Kind: ast.KindSourceFile,
			Loc:  core.NewTextRange(0, 0),
		}

		// We need to check this immediately, not in a listener, since it's about the whole file
		if opts.Mode == "always" && !hasBOM {
			// Expected BOM but not found - report at position 0
			textRange := core.NewTextRange(0, 0)
			fixes := []rule.RuleFix{
				{
					Text:  "\uFEFF", // Add BOM as Unicode character
					Range: textRange,
				},
			}
			ctx.ReportNodeWithFixes(syntheticNode, buildExpectedBOMMessage(), fixes...)
		} else if opts.Mode == "never" && hasBOM {
			// Found BOM but should not have it - remove it
			bomLength := 3 // UTF-8 BOM is 3 bytes
			textRange := core.NewTextRange(0, bomLength)
			fixes := []rule.RuleFix{
				{
					Text:  "", // Remove BOM
					Range: textRange,
				},
			}
			syntheticNode.Loc = textRange
			ctx.ReportNodeWithFixes(syntheticNode, buildUnexpectedBOMMessage(), fixes...)
		}

		return rule.RuleListeners{} // Return empty listeners since we check immediately
	},
}
