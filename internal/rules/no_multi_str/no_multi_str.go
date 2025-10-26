package no_multi_str

import (
	"regexp"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoMultiStrRule implements the no-multi-str rule
// Disallow multiline strings
var NoMultiStrRule = rule.CreateRule(rule.Rule{
	Name: "no-multi-str",
	Run:  run,
})

var linebreakMatcher = regexp.MustCompile(`\r\n|\r|\n|\u2028|\u2029`)

func buildMultilineStringMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "multilineString",
		Description: "Multiline support is limited to browsers supporting ES5 only.",
	}
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	return rule.RuleListeners{
		ast.KindStringLiteral: func(node *ast.Node) {
			if node == nil {
				return
			}

			// Get the raw text of the string literal
			text := node.Text()
			if text == "" {
				return
			}

			// Check if the string contains line breaks (indicating backslash continuation)
			if linebreakMatcher.MatchString(text) {
				// Exclude JSX elements
				parent := node.Parent()
				if parent != nil {
					kindStr := parent.Kind.String()
					if strings.HasPrefix(kindStr, "JSX") {
						return
					}
				}

				ctx.ReportNode(node, buildMultilineStringMessage())
			}
		},
	}
}
