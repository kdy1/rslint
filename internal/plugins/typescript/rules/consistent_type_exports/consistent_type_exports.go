package consistent_type_exports

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type ConsistentTypeExportsOptions struct {
	FixMixedExportsWithInlineTypeSpecifier bool `json:"fixMixedExportsWithInlineTypeSpecifier"`
}

// ConsistentTypeExportsRule enforces consistent type exports
var ConsistentTypeExportsRule = rule.CreateRule(rule.Rule{
	Name: "consistent-type-exports",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := ConsistentTypeExportsOptions{
		FixMixedExportsWithInlineTypeSpecifier: false,
	}

	// Parse options
	if options != nil {
		if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
			if optMap, ok := optArray[0].(map[string]interface{}); ok {
				if fixMixed, ok := optMap["fixMixedExportsWithInlineTypeSpecifier"].(bool); ok {
					opts.FixMixedExportsWithInlineTypeSpecifier = fixMixed
				}
			}
		} else if optMap, ok := options.(map[string]interface{}); ok {
			if fixMixed, ok := optMap["fixMixedExportsWithInlineTypeSpecifier"].(bool); ok {
				opts.FixMixedExportsWithInlineTypeSpecifier = fixMixed
			}
		}
	}

	// Helper to check if a symbol is type-only
	isSymbolTypeBased := func(symbol *ast.Symbol) bool {
		if symbol == nil {
			return false
		}

		// Check if symbol has any value declarations
		// Type-only symbols will only have type-related flags
		flags := symbol.Flags

		// If it has value flags, it's not type-only
		valueFlags := ast.SymbolFlagsValue | ast.SymbolFlagsEnum | ast.SymbolFlagsModule |
			ast.SymbolFlagsFunction | ast.SymbolFlagsClass | ast.SymbolFlagsVariable |
			ast.SymbolFlagsValueModule

		if (flags & valueFlags) != 0 {
			return false
		}

		// If it only has type flags, it's type-only
		typeFlags := ast.SymbolFlagsType | ast.SymbolFlagsInterface | ast.SymbolFlagsTypeAlias |
			ast.SymbolFlagsTypeParameter

		return (flags & typeFlags) != 0
	}

	checkExportDeclaration := func(node *ast.Node) {
		exportDecl := node.AsExportDeclaration()
		if exportDecl == nil {
			return
		}

		// Skip if already marked as type-only
		if exportDecl.IsTypeOnly {
			return
		}

		// Handle export * from 'module'
		if exportDecl.ExportClause == nil && exportDecl.ModuleSpecifier != nil {
			// Check if the entire module exports only types
			moduleSpecifier := exportDecl.ModuleSpecifier
			moduleSymbol := ctx.TypeChecker.GetSymbolAtLocation(moduleSpecifier)

			// For now, we skip checking export * from module
			// This requires more complex analysis of the module's exports
			_ = moduleSymbol
			return
		}

		// Handle named exports: export { x, y, z } or export { x, y, z } from 'module'
		if exportDecl.ExportClause != nil && exportDecl.ExportClause.Kind == ast.KindNamedExports {
			namedExports := exportDecl.ExportClause.AsNamedExports()
			if namedExports == nil || len(namedExports.Elements.Nodes) == 0 {
				return
			}

			var typeSpecifiers []*ast.Node
			var valueSpecifiers []*ast.Node
			var inlineTypeSpecifiers []*ast.Node

			for _, element := range namedExports.Elements.Nodes {
				exportSpecifier := element.AsExportSpecifier()
				if exportSpecifier == nil {
					continue
				}

				// Check if this specifier is already marked as type-only (inline type)
				if exportSpecifier.IsTypeOnly {
					inlineTypeSpecifiers = append(inlineTypeSpecifiers, element)
					continue
				}

				// Get the symbol being exported
				var symbol *ast.Symbol
				// For local exports, we check the property name (what's being exported)
				// For re-exports, we check the name (what's being imported from the module)
				if exportSpecifier.PropertyName != nil {
					symbol = ctx.TypeChecker.GetSymbolAtLocation(exportSpecifier.PropertyName)
				} else {
					symbol = ctx.TypeChecker.GetSymbolAtLocation(exportSpecifier.Name())
				}

				if isSymbolTypeBased(symbol) {
					typeSpecifiers = append(typeSpecifiers, element)
				} else {
					valueSpecifiers = append(valueSpecifiers, element)
				}
			}

			// All specifiers are type-only
			if len(typeSpecifiers) > 0 && len(valueSpecifiers) == 0 && len(inlineTypeSpecifiers) == 0 {
				if len(typeSpecifiers) == 1 {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "typeOverValue",
						Description: "All exports in the declaration are only used as types. Use `export type`.",
					})
				} else {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "typeOverValue",
						Description: "All exports in the declaration are only used as types. Use `export type`.",
					})
				}
				return
			}

			// Mixed: some types, some values
			if len(typeSpecifiers) > 0 && len(valueSpecifiers) > 0 {
				if len(typeSpecifiers) == 1 {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "singleExportIsType",
						Description: "Type export should use `export type`.",
					})
				} else {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "multipleExportsAreTypes",
						Description: "Type exports should use `export type`.",
					})
				}
			}
		}
	}

	return rule.RuleListeners{
		ast.KindExportDeclaration: checkExportDeclaration,
	}
}
