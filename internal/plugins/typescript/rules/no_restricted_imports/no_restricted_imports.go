package no_restricted_imports

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type RestrictedPath struct {
	Name              string   `json:"name"`
	ImportNames       []string `json:"importNames"`
	Message           string   `json:"message"`
	AllowTypeImports  bool     `json:"allowTypeImports"`
}

type RestrictedPattern struct {
	Group             string   `json:"group"`
	ImportNames       []string `json:"importNames"`
	Message           string   `json:"message"`
	CaseSensitive     bool     `json:"caseSensitive"`
	AllowTypeImports  bool     `json:"allowTypeImports"`
	regex             *regexp.Regexp
}

type NoRestrictedImportsOptions struct {
	Paths    []RestrictedPath    `json:"paths"`
	Patterns []RestrictedPattern `json:"patterns"`
}

// getImportSource extracts the module specifier from an import/export declaration
func getImportSource(node *ast.Node) string {
	if node == nil {
		return ""
	}

	switch node.Kind {
	case ast.KindImportDeclaration:
		importDecl := node.AsImportDeclaration()
		if importDecl.ModuleSpecifier != nil && importDecl.ModuleSpecifier.Kind == ast.KindStringLiteral {
			return importDecl.ModuleSpecifier.AsStringLiteral().Text
		}
	case ast.KindExportDeclaration:
		exportDecl := node.AsExportDeclaration()
		if exportDecl.ModuleSpecifier != nil && exportDecl.ModuleSpecifier.Kind == ast.KindStringLiteral {
			return exportDecl.ModuleSpecifier.AsStringLiteral().Text
		}
	case ast.KindImportEqualsDeclaration:
		importEqDecl := node.AsImportEqualsDeclaration()
		if importEqDecl.ModuleReference != nil && importEqDecl.ModuleReference.Kind == ast.KindExternalModuleReference {
			extModRef := importEqDecl.ModuleReference.AsExternalModuleReference()
			if extModRef.Expression != nil && extModRef.Expression.Kind == ast.KindStringLiteral {
				return extModRef.Expression.AsStringLiteral().Text
			}
		}
	}
	return ""
}

// isTypeOnlyImport checks if an import is type-only
func isTypeOnlyImport(node *ast.Node) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindImportDeclaration:
		importDecl := node.AsImportDeclaration()
		// Check for top-level type-only import: import type { ... } from '...'
		if importDecl.ImportClause != nil {
			importClause := importDecl.ImportClause.AsImportClause()
			if importClause.IsTypeOnly {
				return true
			}
		}
	case ast.KindExportDeclaration:
		exportDecl := node.AsExportDeclaration()
		// Check for type-only export: export type { ... } from '...'
		if exportDecl.IsTypeOnly {
			return true
		}
	}
	return false
}

// getNamedImports extracts ONLY the named imports (not default or namespace imports)
// This is used when checking importNames restrictions
func getNamedImports(node *ast.Node) []string {
	names := []string{}

	switch node.Kind {
	case ast.KindImportDeclaration:
		importDecl := node.AsImportDeclaration()
		if importDecl.ImportClause != nil {
			importClause := importDecl.ImportClause.AsImportClause()

			// Only process named imports: import { a, b } from '...'
			if importClause.NamedBindings != nil && importClause.NamedBindings.Kind == ast.KindNamedImports {
				namedImports := importClause.NamedBindings.AsNamedImports()
				if namedImports.Elements != nil {
					for _, element := range namedImports.Elements.Nodes {
						if element.Kind == ast.KindImportSpecifier {
							importSpec := element.AsImportSpecifier()
							// Get the imported name (not the local alias)
							if importSpec.PropertyName != nil && ast.IsIdentifier(importSpec.PropertyName) {
								names = append(names, importSpec.PropertyName.AsIdentifier().Text)
							} else if importSpec.Name() != nil && ast.IsIdentifier(importSpec.Name()) {
								names = append(names, importSpec.Name().AsIdentifier().Text)
							}
						}
					}
				}
			}
		}
	case ast.KindExportDeclaration:
		exportDecl := node.AsExportDeclaration()
		if exportDecl.ExportClause != nil && exportDecl.ExportClause.Kind == ast.KindNamedExports {
			namedExports := exportDecl.ExportClause.AsNamedExports()
			if namedExports.Elements != nil {
				for _, element := range namedExports.Elements.Nodes {
					if element.Kind == ast.KindExportSpecifier {
						exportSpec := element.AsExportSpecifier()
						// Get the exported name (from the source module)
						if exportSpec.PropertyName != nil && ast.IsIdentifier(exportSpec.PropertyName) {
							names = append(names, exportSpec.PropertyName.AsIdentifier().Text)
						} else if exportSpec.Name() != nil && ast.IsIdentifier(exportSpec.Name()) {
							names = append(names, exportSpec.Name().AsIdentifier().Text)
						}
					}
				}
			}
		}
	}

	return names
}

// getImportedNames extracts the list of ALL imported names from an import/export declaration
func getImportedNames(node *ast.Node) []string {
	names := []string{}

	switch node.Kind {
	case ast.KindImportDeclaration:
		importDecl := node.AsImportDeclaration()
		if importDecl.ImportClause != nil {
			importClause := importDecl.ImportClause.AsImportClause()

			// Default import: import Foo from '...'
			if importClause.Name() != nil && ast.IsIdentifier(importClause.Name()) {
				names = append(names, importClause.Name().AsIdentifier().Text)
			}

			// Named imports: import { a, b } from '...'
			if importClause.NamedBindings != nil {
				namedBindings := importClause.NamedBindings

				// Namespace import: import * as ns from '...'
				if namedBindings.Kind == ast.KindNamespaceImport {
					nsImport := namedBindings.AsNamespaceImport()
					if nsImport.Name() != nil && ast.IsIdentifier(nsImport.Name()) {
						names = append(names, nsImport.Name().AsIdentifier().Text)
					}
				}

				// Named imports: import { a, b } from '...'
				if namedBindings.Kind == ast.KindNamedImports {
					namedImports := namedBindings.AsNamedImports()
					if namedImports.Elements != nil {
						for _, element := range namedImports.Elements.Nodes {
							if element.Kind == ast.KindImportSpecifier {
								importSpec := element.AsImportSpecifier()
								// Get the imported name (not the local alias)
								if importSpec.PropertyName != nil && ast.IsIdentifier(importSpec.PropertyName) {
									names = append(names, importSpec.PropertyName.AsIdentifier().Text)
								} else if importSpec.Name() != nil && ast.IsIdentifier(importSpec.Name()) {
									names = append(names, importSpec.Name().AsIdentifier().Text)
								}
							}
						}
					}
				}
			}
		}
	case ast.KindExportDeclaration:
		exportDecl := node.AsExportDeclaration()
		if exportDecl.ExportClause != nil && exportDecl.ExportClause.Kind == ast.KindNamedExports {
			namedExports := exportDecl.ExportClause.AsNamedExports()
			if namedExports.Elements != nil {
				for _, element := range namedExports.Elements.Nodes {
					if element.Kind == ast.KindExportSpecifier {
						exportSpec := element.AsExportSpecifier()
						// Get the exported name (from the source module)
						if exportSpec.PropertyName != nil && ast.IsIdentifier(exportSpec.PropertyName) {
							names = append(names, exportSpec.PropertyName.AsIdentifier().Text)
						} else if exportSpec.Name() != nil && ast.IsIdentifier(exportSpec.Name()) {
							names = append(names, exportSpec.Name().AsIdentifier().Text)
						}
					}
				}
			}
		}
	}

	return names
}

// hasTypeOnlySpecifiers checks if specific import specifiers are type-only
func hasTypeOnlySpecifiers(node *ast.Node, restrictedNames []string) bool {
	if node == nil || len(restrictedNames) == 0 {
		return false
	}

	switch node.Kind {
	case ast.KindImportDeclaration:
		importDecl := node.AsImportDeclaration()
		if importDecl.ImportClause != nil {
			importClause := importDecl.ImportClause.AsImportClause()

			// If the entire import is type-only, all specifiers are type-only
			if importClause.IsTypeOnly {
				return true
			}

			// Check for individual type-only imports: import { type A, B } from '...'
			if importClause.NamedBindings != nil && importClause.NamedBindings.Kind == ast.KindNamedImports {
				namedImports := importClause.NamedBindings.AsNamedImports()
				if namedImports.Elements != nil {
					allRestrictedAreTypeOnly := true
					for _, element := range namedImports.Elements.Nodes {
						if element.Kind == ast.KindImportSpecifier {
							importSpec := element.AsImportSpecifier()

							// Get the imported name
							var importedName string
							if importSpec.PropertyName != nil && ast.IsIdentifier(importSpec.PropertyName) {
								importedName = importSpec.PropertyName.AsIdentifier().Text
							} else if importSpec.Name() != nil && ast.IsIdentifier(importSpec.Name()) {
								importedName = importSpec.Name().AsIdentifier().Text
							}

							// Check if this name is in the restricted list
							isRestricted := false
							for _, restrictedName := range restrictedNames {
								if importedName == restrictedName {
									isRestricted = true
									break
								}
							}

							// If it's a restricted name and not type-only, return false
							if isRestricted && !importSpec.IsTypeOnly {
								allRestrictedAreTypeOnly = false
								break
							}
						}
					}
					return allRestrictedAreTypeOnly
				}
			}
		}
	}

	return false
}

// matchesPattern checks if a path matches a pattern
func matchesPattern(pattern *RestrictedPattern, importPath string) bool {
	if pattern.regex == nil {
		return false
	}

	// For case-insensitive matching, convert both to lowercase
	if !pattern.CaseSensitive {
		return pattern.regex.MatchString(strings.ToLower(importPath))
	}

	return pattern.regex.MatchString(importPath)
}

var NoRestrictedImportsRule = rule.CreateRule(rule.Rule{
	Name: "no-restricted-imports",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := NoRestrictedImportsOptions{
			Paths:    []RestrictedPath{},
			Patterns: []RestrictedPattern{},
		}

		// Parse options with dual-format support (handles both array and object formats)
		if options != nil {
			var optsMap map[string]interface{}
			var ok bool

			// Handle array format: [{ option: value }]
			if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
				optsMap, ok = optArray[0].(map[string]interface{})
			} else {
				// Handle direct object format: { option: value }
				optsMap, ok = options.(map[string]interface{})
			}

			if ok {
				// Parse paths - handle both []string and []interface{} formats
				if pathsInterface, ok := optsMap["paths"].([]interface{}); ok {
					for _, pathItem := range pathsInterface {
						// Handle string format
						if pathStr, ok := pathItem.(string); ok {
							opts.Paths = append(opts.Paths, RestrictedPath{
								Name:             pathStr,
								ImportNames:      []string{},
								Message:          "",
								AllowTypeImports: false,
							})
						} else if pathMap, ok := pathItem.(map[string]interface{}); ok {
							// Handle object format
							path := RestrictedPath{}
							if name, ok := pathMap["name"].(string); ok {
								path.Name = name
							}
							if message, ok := pathMap["message"].(string); ok {
								path.Message = message
							}
							if allowTypeImports, ok := pathMap["allowTypeImports"].(bool); ok {
								path.AllowTypeImports = allowTypeImports
							}
							if importNames, ok := pathMap["importNames"].([]interface{}); ok {
								for _, name := range importNames {
									if nameStr, ok := name.(string); ok {
										path.ImportNames = append(path.ImportNames, nameStr)
									}
								}
							} else if importNamesSlice, ok := pathMap["importNames"].([]string); ok {
								// Handle []string directly from Go tests
								path.ImportNames = importNamesSlice
							}
							opts.Paths = append(opts.Paths, path)
						}
					}
				} else if pathsSlice, ok := optsMap["paths"].([]string); ok {
					// Handle []string directly from Go tests
					for _, pathStr := range pathsSlice {
						opts.Paths = append(opts.Paths, RestrictedPath{
							Name:             pathStr,
							ImportNames:      []string{},
							Message:          "",
							AllowTypeImports: false,
						})
					}
				}

				// Parse patterns - handle both []string and []interface{} formats
				if patternsInterface, ok := optsMap["patterns"].([]interface{}); ok {
					for _, patternItem := range patternsInterface {
						// Handle string format
						if patternStr, ok := patternItem.(string); ok {
							// Convert glob pattern to regex
							regexPattern := globToRegex(patternStr)
							if re, err := regexp.Compile(regexPattern); err == nil {
								opts.Patterns = append(opts.Patterns, RestrictedPattern{
									Group:            patternStr,
									Message:          "",
									CaseSensitive:    true,
									AllowTypeImports: false,
									regex:            re,
								})
							}
						} else if patternMap, ok := patternItem.(map[string]interface{}); ok {
							// Handle object format
							pattern := RestrictedPattern{
								CaseSensitive: true, // Default to true
							}
							if group, ok := patternMap["group"].(string); ok {
								pattern.Group = group
							}
							if message, ok := patternMap["message"].(string); ok {
								pattern.Message = message
							}
							if caseSensitive, ok := patternMap["caseSensitive"].(bool); ok {
								pattern.CaseSensitive = caseSensitive
							}
							if allowTypeImports, ok := patternMap["allowTypeImports"].(bool); ok {
								pattern.AllowTypeImports = allowTypeImports
							}
							if importNames, ok := patternMap["importNames"].([]interface{}); ok {
								for _, name := range importNames {
									if nameStr, ok := name.(string); ok {
										pattern.ImportNames = append(pattern.ImportNames, nameStr)
									}
								}
							} else if importNamesSlice, ok := patternMap["importNames"].([]string); ok {
								// Handle []string directly from Go tests
								pattern.ImportNames = importNamesSlice
							}

							// Convert glob pattern to regex
							regexPattern := globToRegex(pattern.Group)
							if !pattern.CaseSensitive {
								regexPattern = "(?i)" + regexPattern
							}
							if re, err := regexp.Compile(regexPattern); err == nil {
								pattern.regex = re
								opts.Patterns = append(opts.Patterns, pattern)
							}
						}
					}
				} else if patternsSlice, ok := optsMap["patterns"].([]string); ok {
					// Handle []string directly from Go tests
					for _, patternStr := range patternsSlice {
						// Convert glob pattern to regex
						regexPattern := globToRegex(patternStr)
						if re, err := regexp.Compile(regexPattern); err == nil {
							opts.Patterns = append(opts.Patterns, RestrictedPattern{
								Group:            patternStr,
								Message:          "",
								CaseSensitive:    true,
								AllowTypeImports: false,
								regex:            re,
							})
						}
					}
				}
			}
		}

		checkImport := func(node *ast.Node) {
			importSource := getImportSource(node)
			if importSource == "" {
				return
			}

			// Remove quotes from the import source
			importSource = strings.Trim(importSource, "\"'")

			// Check against restricted paths
			for _, restrictedPath := range opts.Paths {
				if importSource == restrictedPath.Name {
					// Check if type-only imports are allowed
					if restrictedPath.AllowTypeImports && isTypeOnlyImport(node) {
						continue
					}

					// Check for specific import names
					if len(restrictedPath.ImportNames) > 0 {
						// Only check named imports, not default or namespace imports
						importedNames := getNamedImports(node)

						for _, importedName := range importedNames {
							for _, restrictedName := range restrictedPath.ImportNames {
								if importedName == restrictedName {
									// Check if this specific import is type-only
									if restrictedPath.AllowTypeImports && hasTypeOnlySpecifiers(node, []string{restrictedName}) {
										// This specific import is type-only and allowed, skip it
										continue
									}

									// Report the restricted import name
									if restrictedPath.Message != "" {
										ctx.ReportNode(node, rule.RuleMessage{
											Id:          "importNameWithCustomMessage",
											Description: restrictedPath.Message,
										})
									} else {
										ctx.ReportNode(node, rule.RuleMessage{
											Id:          "importName",
											Description: "'" + restrictedName + "' import from '" + restrictedPath.Name + "' is restricted.",
										})
									}
									return
								}
							}
						}

						// If no restricted names were imported, don't report this path
						// (continue to check other paths/patterns)
						continue
					} else {
						// Report the entire import
						if restrictedPath.Message != "" {
							ctx.ReportNode(node, rule.RuleMessage{
								Id:          "pathWithCustomMessage",
								Description: restrictedPath.Message,
							})
						} else {
							ctx.ReportNode(node, rule.RuleMessage{
								Id:          "path",
								Description: "'" + restrictedPath.Name + "' import is restricted from being used.",
							})
						}
					}
					return
				}
			}

			// Check against restricted patterns
			for _, pattern := range opts.Patterns {
				if matchesPattern(&pattern, importSource) {
					// Check if type-only imports are allowed
					if pattern.AllowTypeImports && isTypeOnlyImport(node) {
						continue
					}

					// Check for specific import names
					if len(pattern.ImportNames) > 0 {
						// Only check named imports, not default or namespace imports
						importedNames := getNamedImports(node)

						for _, importedName := range importedNames {
							for _, restrictedName := range pattern.ImportNames {
								if importedName == restrictedName {
									// Check if this specific import is type-only
									if pattern.AllowTypeImports && hasTypeOnlySpecifiers(node, []string{restrictedName}) {
										// This specific import is type-only and allowed, skip it
										continue
									}

									// Report the restricted import name
									if pattern.Message != "" {
										ctx.ReportNode(node, rule.RuleMessage{
											Id:          "importNameWithCustomMessage",
											Description: pattern.Message,
										})
									} else {
										ctx.ReportNode(node, rule.RuleMessage{
											Id:          "importName",
											Description: "'" + restrictedName + "' import is restricted from being used by a pattern.",
										})
									}
									return
								}
							}
						}

						// If no restricted names were imported, don't report this pattern
						// (continue to check other patterns)
						continue
					} else {
						// Report the entire import
						if pattern.Message != "" {
							ctx.ReportNode(node, rule.RuleMessage{
								Id:          "patternWithCustomMessage",
								Description: pattern.Message,
							})
						} else {
							ctx.ReportNode(node, rule.RuleMessage{
								Id:          "patterns",
								Description: "'" + importSource + "' import is restricted from being used by a pattern.",
							})
						}
					}
					return
				}
			}
		}

		return rule.RuleListeners{
			ast.KindImportDeclaration: checkImport,
			ast.KindExportDeclaration: func(node *ast.Node) {
				// Only check re-exports (export ... from '...')
				exportDecl := node.AsExportDeclaration()
				if exportDecl.ModuleSpecifier != nil {
					checkImport(node)
				}
			},
			ast.KindImportEqualsDeclaration: checkImport,
		}
	},
})

// globToRegex converts a glob pattern to a regex pattern
func globToRegex(pattern string) string {
	// Escape special regex characters except * and ?
	pattern = regexp.QuoteMeta(pattern)

	// Convert glob wildcards to regex
	// ** matches any characters (including directory separators)
	pattern = strings.ReplaceAll(pattern, `\*\*`, ".*")
	// * matches any characters except directory separators
	pattern = strings.ReplaceAll(pattern, `\*`, "[^"+regexp.QuoteMeta(string(filepath.Separator))+"]*")
	// ? matches a single character
	pattern = strings.ReplaceAll(pattern, `\?`, ".")

	// Anchor the pattern
	return "^" + pattern + "$"
}
