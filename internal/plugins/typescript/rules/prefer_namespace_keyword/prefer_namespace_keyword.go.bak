package prefer_namespace_keyword

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// PreferNamespaceKeywordRule implements the prefer-namespace-keyword rule
// Require namespace over module keyword
var PreferNamespaceKeywordRule = rule.CreateRule(rule.Rule{
	Name: "prefer-namespace-keyword",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	return rule.RuleListeners{
		ast.KindModuleDeclaration: func(node *ast.Node) {
			moduleDecl := node.AsModuleDeclaration()
			if moduleDecl == nil {
				return
			}

			// Check if this is an external module declaration (declare module 'foo')
			// These are valid and should not be flagged
			if moduleDecl.Name != nil && moduleDecl.Name.Kind == ast.KindStringLiteral {
				// This is an external module declaration like: declare module 'foo' {}
				// This is valid, don't report
				return
			}

			// Check if it uses the 'module' keyword instead of 'namespace'
			// We need to check the source text to see if it starts with 'module'
			moduleRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
			sourceText := ctx.SourceFile.Text()
			moduleText := sourceText[moduleRange.Pos():moduleRange.End()]

			// Skip leading modifiers (declare, export, etc.)
			trimmed := strings.TrimSpace(moduleText)

			// Check if it starts with 'declare module' or just 'module'
			var hasDeclare bool
			var moduleKeywordPos int

			if strings.HasPrefix(trimmed, "declare ") {
				hasDeclare = true
				remaining := strings.TrimPrefix(trimmed, "declare ")
				remaining = strings.TrimSpace(remaining)
				if strings.HasPrefix(remaining, "module ") {
					// This is a custom module (namespace), should use 'namespace' keyword
					// Calculate position of 'module' keyword in original text
					moduleKeywordPos = moduleRange.Pos() + strings.Index(moduleText, "module")
				} else {
					// Already uses 'namespace'
					return
				}
			} else if strings.HasPrefix(trimmed, "module ") {
				// Direct 'module' keyword without 'declare'
				moduleKeywordPos = moduleRange.Pos() + strings.Index(moduleText, "module")
			} else {
				// Already uses 'namespace' or other syntax
				return
			}

			// Build the fix: replace 'module' with 'namespace'
			replacement := strings.Replace(moduleText, "module ", "namespace ", 1)

			// Report with autofix
			ctx.ReportNodeWithFixes(node, rule.RuleMessage{
				Id:          "useNamespace",
				Description: "Use 'namespace' instead of 'module' to declare custom TypeScript modules.",
			}, rule.RuleFix{
				Range: utils.TextRange{
					Pos: moduleRange.Pos(),
					End: moduleRange.End(),
				},
				Text: replacement,
			})
		},
	}
}
