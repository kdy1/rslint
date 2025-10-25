package no_duplicate_imports

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// NoDuplicateImportsOptions defines the configuration options for this rule
type NoDuplicateImportsOptions struct {
	IncludeExports bool `json:"includeExports"`
}

// parseOptions parses and validates the rule options
func parseOptions(options any) NoDuplicateImportsOptions {
	opts := NoDuplicateImportsOptions{
		IncludeExports: false,
	}

	if options == nil {
		return opts
	}

	// Handle both array format [{ option: value }] and object format { option: value }
	var optsMap map[string]interface{}
	if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
		optsMap, _ = optArray[0].(map[string]interface{})
	} else {
		optsMap, _ = options.(map[string]interface{})
	}

	if optsMap != nil {
		if v, ok := optsMap["includeExports"].(bool); ok {
			opts.IncludeExports = v
		}
	}

	return opts
}

// NoDuplicateImportsRule implements the no-duplicate-imports rule
// Disallow duplicate module imports
var NoDuplicateImportsRule = rule.Rule{
	Name: "no-duplicate-imports",
	Run:  run,
}

type importInfo struct {
	module   string
	node     *ast.Node
	isImport bool // true for import, false for export
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := parseOptions(options)

	// Track all imports and exports in the file
	var imports []importInfo

	return rule.RuleListeners{
		ast.KindSourceFile: func(node *ast.Node) {
			// Process at the end of file to check all imports/exports
		},
		ast.KindImportDeclaration: func(node *ast.Node) {
			importDecl := node.AsImportDeclaration()
			if importDecl == nil || importDecl.ModuleSpecifier == nil {
				return
			}

			module := getModuleName(ctx.SourceFile, importDecl.ModuleSpecifier)
			if module == "" {
				return
			}

			// Check for duplicates in existing imports
			for _, imp := range imports {
				if imp.module == module && imp.isImport {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "import",
						Description: fmt.Sprintf("'%s' import is duplicated.", module),
					})
					return
				}
				// If includeExports is true, also check against exports
				if opts.IncludeExports && imp.module == module && !imp.isImport {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "importExport",
						Description: fmt.Sprintf("'%s' import is duplicated as export.", module),
					})
					return
				}
			}

			imports = append(imports, importInfo{
				module:   module,
				node:     node,
				isImport: true,
			})
		},
		ast.KindExportDeclaration: func(node *ast.Node) {
			if !opts.IncludeExports {
				return
			}

			exportDecl := node.AsExportDeclaration()
			if exportDecl == nil || exportDecl.ModuleSpecifier == nil {
				// export { foo } without 'from' clause - not a re-export
				return
			}

			module := getModuleName(ctx.SourceFile, exportDecl.ModuleSpecifier)
			if module == "" {
				return
			}

			// Check for duplicates in existing imports and exports
			for _, imp := range imports {
				if imp.module == module && !imp.isImport {
					// Duplicate export
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "export",
						Description: fmt.Sprintf("'%s' export is duplicated.", module),
					})
					return
				}
				if imp.module == module && imp.isImport {
					// Export duplicating an import
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "importExport",
						Description: fmt.Sprintf("'%s' export is duplicated as import.", module),
					})
					return
				}
			}

			imports = append(imports, importInfo{
				module:   module,
				node:     node,
				isImport: false,
			})
		},
	}
}

// getModuleName extracts the module name from a module specifier
func getModuleName(sourceFile *ast.SourceFile, specifier *ast.Node) string {
	if specifier == nil {
		return ""
	}

	if specifier.Kind == ast.KindStringLiteral {
		textRange := utils.TrimNodeTextRange(sourceFile, specifier)
		text := sourceFile.Text()[textRange.Pos():textRange.End()]
		if len(text) >= 2 {
			return text[1 : len(text)-1] // Remove quotes
		}
	}

	return ""
}
