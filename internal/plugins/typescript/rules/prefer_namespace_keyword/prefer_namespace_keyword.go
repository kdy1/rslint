package prefer_namespace_keyword

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// PreferNamespaceKeywordRule implements the prefer-namespace-keyword rule
// Require using `namespace` keyword over `module` keyword to declare custom TypeScript modules
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

			// Skip if this is already a namespace or a string literal module (e.g., declare module "foo")
			if moduleDecl.Keyword == ast.KindNamespaceKeyword {
				return
			}

			// Check if it's a module with string literal name (ambient module)
			if moduleDecl.Name != nil && moduleDecl.Name().Kind == ast.KindStringLiteral {
				return
			}

			// This is a module declaration that should use namespace keyword
			// Get the position of the module keyword
			moduleKeywordStart := node.Pos()
			moduleKeywordEnd := moduleKeywordStart + 6 // length of "module"

			// Check if preceded by 'declare' modifier
			if utils.IncludesModifier(moduleDecl.Modifiers(), ast.KindDeclareKeyword) {
				// Need to skip past 'declare ' to find 'module'
				sourceText := ctx.SourceFile.Text()
				for i := moduleKeywordStart; i < moduleKeywordStart+50 && i < len(sourceText)-6; i++ {
					if sourceText[i:i+6] == "module" {
						moduleKeywordStart = i
						moduleKeywordEnd = i + 6
						break
					}
				}
			}

			ctx.ReportNodeWithFixes(node, buildUseNamespaceMessage(),
				rule.RuleFixReplaceRange(
					core.NewTextRange(moduleKeywordStart, moduleKeywordEnd),
					"namespace",
				),
			)
		},
	}
}

func buildUseNamespaceMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useNamespace",
		Description: "Use 'namespace' instead of 'module' to declare custom TypeScript modules.",
	}
}
