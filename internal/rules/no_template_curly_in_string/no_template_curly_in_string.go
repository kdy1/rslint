package no_template_curly_in_string

import (
	"regexp"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// NoTemplateCurlyInStringRule implements the no-template-curly-in-string rule
// Disallow template literal placeholder syntax in regular strings
var NoTemplateCurlyInStringRule = rule.Rule{
	Name: "no-template-curly-in-string",
	Run:  run,
}

var templateLiteralPattern = regexp.MustCompile(`\$\{[^}]+\}`)

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	return rule.RuleListeners{
		ast.KindStringLiteral: func(node *ast.Node) {
			stringLit := node.AsStringLiteral()
			if stringLit == nil {
				return
			}

			// Get the string content without quotes
			rng := utils.TrimNodeTextRange(ctx.SourceFile, node)
			text := ctx.SourceFile.Text()[rng.Pos():rng.End()]

			// Remove surrounding quotes
			if len(text) < 2 {
				return
			}
			stringContent := text[1 : len(text)-1]

			// Check if the string contains template literal syntax
			if templateLiteralPattern.MatchString(stringContent) {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "unexpectedTemplateExpression",
					Description: "Unexpected template string expression.",
				})
			}
		},
	}
}
