package no_useless_empty_export

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var NoUselessEmptyExportRule = rule.CreateRule(rule.Rule{
	Name: "no-useless-empty-export",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindExportDeclaration: func(node *ast.Node) {
				exportDecl := node.AsExportDeclaration()
				if exportDecl == nil {
					return
				}

				// Check if this is an empty export: export {}
				// An empty export has no export clause (or an empty one) and no module specifier
				if exportDecl.ExportClause != nil {
					// Check if it's a named export clause
					if ast.IsNamedExports(exportDecl.ExportClause) {
						namedExports := exportDecl.ExportClause.AsNamedExports()
						// If there are elements in the named exports, it's not empty
						if namedExports != nil && namedExports.Elements != nil && len(namedExports.Elements.Nodes) > 0 {
							return
						}
					} else {
						// If it's not a named export, it's not an empty export
						return
					}
				}

				// If there's a module specifier (from clause), it's a re-export, not an empty export
				if exportDecl.ModuleSpecifier != nil {
					return
				}

				// Now we have confirmed this is an empty export: export {}
				// Check if the file has other imports or exports that make it a module already
				sourceFile := ctx.SourceFile
				if sourceFile == nil || sourceFile.Statements == nil {
					return
				}

				hasOtherModuleSyntax := false
				for _, stmt := range sourceFile.Statements.Nodes {
					// Skip the current export declaration
					if stmt == node {
						continue
					}

					// Check for import declarations (but not type-only imports)
					if stmt.Kind == ast.KindImportDeclaration {
						importDecl := stmt.AsImportDeclaration()
						if importDecl != nil {
							// Check if it's a type-only import
							isTypeOnly := false
							if importDecl.ImportClause != nil {
								if importDecl.ImportClause.IsTypeOnly {
									isTypeOnly = true
								}
							}
							if !isTypeOnly {
								hasOtherModuleSyntax = true
								break
							}
						}
						continue
					}

					// Check for export declarations
					if stmt.Kind == ast.KindExportDeclaration {
						hasOtherModuleSyntax = true
						break
					}

					// Check for export assignments (export = )
					if stmt.Kind == ast.KindExportAssignment {
						hasOtherModuleSyntax = true
						break
					}

					// Check for import equals declarations (import _ = require('_'))
					if stmt.Kind == ast.KindImportEqualsDeclaration {
						hasOtherModuleSyntax = true
						break
					}

					// Check for declarations with export modifier
					modifiers := getModifiers(stmt)
					if modifiers != nil {
						hasExportModifier := false
						hasDeclareModifier := false
						for _, modifier := range modifiers.Nodes {
							if modifier.Kind == ast.KindExportKeyword {
								hasExportModifier = true
							}
							if modifier.Kind == ast.KindDeclareKeyword {
								hasDeclareModifier = true
							}
						}
						// Only count as module syntax if it has export but not declare
						// or if it's not a type-only export
						if hasExportModifier {
							// Check if this is a type-only export
							isTypeOnlyExport := false
							if stmt.Kind == ast.KindTypeAliasDeclaration {
								isTypeOnlyExport = true
							} else if stmt.Kind == ast.KindInterfaceDeclaration {
								isTypeOnlyExport = true
							}
							// If it's not type-only, or if it's a value export, count it
							if !isTypeOnlyExport && !hasDeclareModifier {
								hasOtherModuleSyntax = true
								break
							}
						}
					}
				}

				// If the file already has other module syntax, the empty export is useless
				if hasOtherModuleSyntax {
					ctx.ReportNodeWithFixes(
						node,
						rule.RuleMessage{
							Id:          "uselessExport",
							Description: "Empty export statement is useless because the file already contains other module syntax.",
						},
						rule.RuleFixRemove(ctx.SourceFile, node),
					)
				}
			},
		}
	},
})

// Helper function to get modifiers from various declaration types
func getModifiers(node *ast.Node) *ast.NodeArray {
	switch node.Kind {
	case ast.KindVariableStatement:
		stmt := node.AsVariableStatement()
		if stmt != nil {
			return stmt.Modifiers()
		}
	case ast.KindFunctionDeclaration:
		decl := node.AsFunctionDeclaration()
		if decl != nil {
			return decl.Modifiers()
		}
	case ast.KindClassDeclaration:
		decl := node.AsClassDeclaration()
		if decl != nil {
			return decl.Modifiers()
		}
	case ast.KindInterfaceDeclaration:
		decl := node.AsInterfaceDeclaration()
		if decl != nil {
			return decl.Modifiers()
		}
	case ast.KindTypeAliasDeclaration:
		decl := node.AsTypeAliasDeclaration()
		if decl != nil {
			return decl.Modifiers()
		}
	case ast.KindEnumDeclaration:
		decl := node.AsEnumDeclaration()
		if decl != nil {
			return decl.Modifiers()
		}
	case ast.KindModuleDeclaration:
		decl := node.AsModuleDeclaration()
		if decl != nil {
			return decl.Modifiers()
		}
	}
	return nil
}
