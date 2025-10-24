package prefer_namespace_keyword

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var PreferNamespaceKeywordRule = rule.CreateRule(rule.Rule{
	Name: "prefer-namespace-keyword",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// This rule enforces using 'namespace' keyword over 'module' keyword
		// for defining TypeScript namespaces

		return rule.RuleListeners{
			ast.KindModuleDeclaration: func(node *ast.Node) {
				moduleDecl := node.AsModuleDeclaration()
				if moduleDecl == nil {
					return
				}

				// Check if this is using 'module' keyword instead of 'namespace'
				// TODO: Implement full checking logic
				// 1. Determine if the declaration uses 'module' keyword
				// 2. Check if it's not an ambient module (declare module "foo")
				// 3. Report and suggest using 'namespace' instead

				// Get the module keyword position
				nameRange := utils.TrimNodeTextRange(ctx.SourceFile, moduleDecl.Name)
				moduleText := ctx.SourceFile.Text()[node.Pos():nameRange.Pos()]

				// Simple check for 'module' keyword
				if len(moduleText) > 0 {
					// TODO: Proper implementation needed
					// This is a placeholder that needs AST analysis to determine
					// if this is a 'module' vs 'namespace' declaration
				}
			},
		}
	},
})
