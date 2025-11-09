package no_import_type_side_effects

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoImportTypeSideEffectsRule enforces the use of top-level type-only imports
// when an import only has specifiers with inline type qualifiers
var NoImportTypeSideEffectsRule = rule.CreateRule(rule.Rule{
	Name: "no-import-type-side-effects",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	checkImportDeclaration := func(node *ast.Node) {
		importDecl := node.AsImportDeclaration()
		if importDecl == nil {
			return
		}

		importClauseNode := importDecl.ImportClause
		if importClauseNode == nil {
			// Side-effect only import: import 'mod';
			return
		}

		importClause := importClauseNode.AsImportClause()
		if importClause == nil {
			return
		}

		// Skip if entire import is already type-only: import type { A } from 'mod';
		if importClause.IsTypeOnly {
			return
		}

		// Check if there's a default import (e.g., import T from 'mod')
		if importClause.Name() != nil {
			// Has default import, not applicable
			return
		}

		// Check named bindings
		namedBindings := importClause.NamedBindings
		if namedBindings == nil {
			return
		}

		// Skip namespace imports: import * as T from 'mod'
		if namedBindings.Kind == ast.KindNamespaceImport {
			return
		}

		// Handle named imports: import { type A, type B } from 'mod'
		if namedBindings.Kind != ast.KindNamedImports {
			return
		}

		namedImports := namedBindings.AsNamedImports()
		if namedImports == nil || len(namedImports.Elements.Nodes) == 0 {
			return
		}

		// Check if all import specifiers are type-only
		allTypeOnly := true
		hasAnySpecifiers := false

		for _, element := range namedImports.Elements.Nodes {
			importSpecifier := element.AsImportSpecifier()
			if importSpecifier == nil {
				continue
			}

			hasAnySpecifiers = true

			// Check if this specifier has inline type qualifier
			if !importSpecifier.IsTypeOnly {
				allTypeOnly = false
				break
			}
		}

		// If all specifiers are type-only, report the issue and provide a fix
		if hasAnySpecifiers && allTypeOnly {
			// Build the fix: convert to top-level type import
			fix := buildFix(ctx, node, importDecl, namedImports)

			ctx.ReportNodeWithFixes(node, rule.RuleMessage{
				Id:          "useTopLevelQualifier",
				Description: "Use top-level type qualifier instead of inline type qualifiers.",
			}, fix)
		}
	}

	return rule.RuleListeners{
		ast.KindImportDeclaration: checkImportDeclaration,
	}
}

// buildFix creates a fix that converts inline type imports to top-level type import
func buildFix(ctx rule.RuleContext, node *ast.Node, importDecl *ast.ImportDeclaration, namedImports *ast.NamedImports) rule.RuleFix {
	sourceText := ctx.SourceFile.Text()

	// Get the start and end of the import declaration
	importStart := node.Pos()
	importEnd := node.End()

	// Build the new import statement
	// We need to extract the module specifier
	moduleSpecifier := importDecl.ModuleSpecifier
	moduleText := strings.TrimSpace(sourceText[moduleSpecifier.Pos():moduleSpecifier.End()])

	// Build the specifiers without the inline "type" keyword
	var specifierTexts []string
	for _, element := range namedImports.Elements.Nodes {
		importSpecifier := element.AsImportSpecifier()
		if importSpecifier == nil {
			continue
		}

		// Get the specifier text and remove the "type " prefix
		specifierText := getSpecifierText(ctx, importSpecifier)
		specifierTexts = append(specifierTexts, specifierText)
	}

	// Build the new import statement
	newImport := "import type { "
	for i, specifier := range specifierTexts {
		if i > 0 {
			newImport += ", "
		}
		newImport += specifier
	}
	newImport += " } from " + moduleText + ";"

	return rule.RuleFixReplaceRange(
		core.NewTextRange(importStart, importEnd),
		newImport,
	)
}

// getSpecifierText extracts the specifier text without the "type " keyword
func getSpecifierText(ctx rule.RuleContext, specifier *ast.ImportSpecifier) string {
	sourceText := ctx.SourceFile.Text()

	// If there's a property name (alias), use "PropertyName as Name"
	// Otherwise, just use the name
	if specifier.PropertyName != nil {
		// Extract both identifiers directly and trim whitespace
		propertyName := strings.TrimSpace(sourceText[specifier.PropertyName.Pos():specifier.PropertyName.End()])
		importedName := strings.TrimSpace(sourceText[specifier.Name().Pos():specifier.Name().End()])
		return propertyName + " as " + importedName
	}

	// For non-aliased imports, just get the name
	name := strings.TrimSpace(sourceText[specifier.Name().Pos():specifier.Name().End()])
	return name
}
