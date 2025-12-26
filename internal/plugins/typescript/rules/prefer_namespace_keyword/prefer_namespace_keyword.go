package prefer_namespace_keyword

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// build the message for prefer-namespace-keyword rule
func buildUseNamespaceMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useNamespace",
		Description: "Use 'namespace' instead of 'module' to declare custom TypeScript modules.",
	}
}

// rule instance
// Enforces using the namespace keyword instead of the module keyword for TypeScript custom modules
var PreferNamespaceKeywordRule = rule.CreateRule(rule.Rule{
	Name: "prefer-namespace-keyword",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindModuleDeclaration: func(node *ast.Node) {
				moduleDecl := node.AsModuleDeclaration()
				if moduleDecl == nil {
					return
				}

				// Only care about module declarations that use the 'module' keyword
				if moduleDecl.Keyword != ast.KindModuleKeyword {
					return
				}

				// Get the module name
				moduleName := moduleDecl.Name()
				if moduleName == nil {
					return
				}

				// Ignore module declarations with string literal names (like declare module 'foo')
				// These are ambient module declarations for external APIs
				if moduleName.Kind == ast.KindStringLiteral {
					return
				}

				// Get the text range of the 'module' keyword
				// The keyword starts at the node position (accounting for any 'declare' modifier)
				nodeStart := utils.TrimNodeTextRange(ctx.SourceFile, node).Pos()

				// If the node has a declare modifier, skip past it
				if utils.IncludesModifier(node, ast.KindDeclareKeyword) {
					// Find the position after 'declare '
					sourceText := ctx.SourceFile.Text()
					// Start from nodeStart and find "module" keyword
					for i := nodeStart; i < len(sourceText)-6; i++ {
						if sourceText[i:i+6] == "module" {
							nodeStart = i
							break
						}
					}
				}

				// The fix replaces 'module' (6 characters) with 'namespace'
				keywordEnd := nodeStart + 6

				// Create a fix that replaces 'module' with 'namespace'
				fix := rule.RuleFix{
					Text:  "namespace",
					Range: core.NewTextRange(nodeStart, keywordEnd),
				}

				ctx.ReportNodeWithFixes(node, buildUseNamespaceMessage(), fix)
			},
		}
	},
})
