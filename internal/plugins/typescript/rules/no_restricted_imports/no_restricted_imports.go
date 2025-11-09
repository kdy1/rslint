package no_restricted_imports

import (
	"regexp"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type PathConfig struct {
	Name             string   `json:"name"`
	ImportNames      []string `json:"importNames"`
	Message          string   `json:"message"`
	AllowTypeImports bool     `json:"allowTypeImports"`
}

type PatternConfig struct {
	Group            []string `json:"group"`
	Message          string   `json:"message"`
	CaseSensitive    *bool    `json:"caseSensitive"`
	AllowTypeImports bool     `json:"allowTypeImports"`
	ImportNames      []string `json:"importNames"`
}

type NoRestrictedImportsOptions struct {
	Paths    interface{} `json:"paths"`
	Patterns interface{} `json:"patterns"`
}

// isStringOrTemplateLiteral checks if a node is a string literal or template literal
func isStringOrTemplateLiteral(node *ast.Node) bool {
	return (node.Kind == ast.KindStringLiteral) ||
		(node.Kind == ast.KindTemplateExpression && node.AsTemplateExpression().TemplateSpans == nil) ||
		(node.Kind == ast.KindNoSubstitutionTemplateLiteral)
}

// getStaticStringValue extracts static string value from literal or template
func getStaticStringValue(node *ast.Node) (string, bool) {
	switch node.Kind {
	case ast.KindStringLiteral:
		return node.AsStringLiteral().Text, true
	case ast.KindTemplateExpression:
		// Only handle simple template literals without expressions
		te := node.AsTemplateExpression()
		if te.TemplateSpans == nil || len(te.TemplateSpans.Nodes) == 0 {
			return te.Head.Text(), true
		}
	case ast.KindNoSubstitutionTemplateLiteral:
		// Handle simple template literals `string`
		return node.AsNoSubstitutionTemplateLiteral().Text, true
	}
	return "", false
}

// isTypeImport checks if an import is type-only
func isTypeImport(node *ast.Node) bool {
	// Check for import type declarations
	if node.Kind == ast.KindImportDeclaration {
		importDecl := node.AsImportDeclaration()
		// Check if it's an import type statement
		if importDecl.ImportClause != nil && importDecl.ImportClause.IsTypeOnly() {
			return true
		}
	}

	// Check for import = require with type only
	if node.Kind == ast.KindImportEqualsDeclaration {
		importEq := node.AsImportEqualsDeclaration()
		return importEq.IsTypeOnly
	}

	// Check for export type from
	if node.Kind == ast.KindExportDeclaration {
		exportDecl := node.AsExportDeclaration()
		return exportDecl.IsTypeOnly
	}

	return false
}

// getImportNames extracts the names being imported from an import declaration
func getImportNames(node *ast.Node) []string {
	var names []string

	if node.Kind == ast.KindImportDeclaration {
		importDecl := node.AsImportDeclaration()
		if importDecl.ImportClause != nil {
			clauseNode := importDecl.ImportClause
			clause := clauseNode.AsImportClause()
			if clause == nil {
				return names
			}

			// Default import
			if clause.Name() != nil {
				names = append(names, clause.Name().AsIdentifier().Text)
			}

			// Named imports - check if there are child nodes for named bindings
			// We need to iterate through the clause's children to find NamedImports or NamespaceImport
			for _, child := range clauseNode.Children() {
				if child.Kind == ast.KindNamedImports {
					namedImports := child.AsNamedImports()
					for _, elem := range namedImports.Elements.Nodes {
						importSpec := elem.AsImportSpecifier()
						// Use the property name if it exists (for renamed imports), otherwise use the name
						if importSpec.PropertyName != nil {
							names = append(names, importSpec.PropertyName.AsIdentifier().Text)
						} else if importSpec.Name() != nil {
							names = append(names, importSpec.Name().AsIdentifier().Text)
						}
					}
				} else if child.Kind == ast.KindNamespaceImport {
					// Namespace import: import * as foo
					nsImport := child.AsNamespaceImport()
					if nsImport.Name() != nil {
						names = append(names, nsImport.Name().AsIdentifier().Text)
					}
				}
			}
		}
	} else if node.Kind == ast.KindExportDeclaration {
		exportDecl := node.AsExportDeclaration()
		if exportDecl.ExportClause != nil && exportDecl.ExportClause.Kind == ast.KindNamedExports {
			namedExports := exportDecl.ExportClause.AsNamedExports()
			for _, elem := range namedExports.Elements.Nodes {
				exportSpec := elem.AsExportSpecifier()
				// Use the property name if it exists
				if exportSpec.PropertyName != nil {
					names = append(names, exportSpec.PropertyName.AsIdentifier().Text)
				} else if exportSpec.Name() != nil {
					names = append(names, exportSpec.Name().AsIdentifier().Text)
				}
			}
		}
	}

	return names
}

// matchesPattern checks if a path matches a glob pattern
func matchesPattern(path string, pattern string) bool {
	// Handle negation patterns
	if strings.HasPrefix(pattern, "!") {
		return !matchesPattern(path, pattern[1:])
	}

	// Convert glob pattern to regex
	// Simple glob to regex conversion
	regexPattern := regexp.QuoteMeta(pattern)
	regexPattern = strings.ReplaceAll(regexPattern, `\*`, ".*")
	regexPattern = "^" + regexPattern + "$"

	matched, err := regexp.MatchString(regexPattern, path)
	if err != nil {
		return false
	}
	return matched
}

// matchesPatterns checks if a path matches any pattern in a group
func matchesPatterns(path string, patterns []string) bool {
	// First check all positive patterns
	hasPositiveMatch := false
	for _, pattern := range patterns {
		if !strings.HasPrefix(pattern, "!") {
			if matchesPattern(path, pattern) {
				hasPositiveMatch = true
				break
			}
		}
	}

	if !hasPositiveMatch {
		return false
	}

	// Then check negation patterns
	for _, pattern := range patterns {
		if strings.HasPrefix(pattern, "!") {
			if matchesPattern(path, pattern) {
				return false
			}
		}
	}

	return true
}

var NoRestrictedImportsRule = rule.CreateRule(rule.Rule{
	Name: "no-restricted-imports",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		var pathConfigs []PathConfig
		var patternConfigs []PatternConfig
		var simplePaths []string

		// Parse options
		if options != nil {
			var optsArray []interface{}

			// Handle array format: [{ paths: [], patterns: [] }] or ['path1', 'path2']
			if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
				// Check if first element is a string (simple format)
				if _, isString := optArray[0].(string); isString {
					// Simple string array format
					for _, item := range optArray {
						if str, ok := item.(string); ok {
							simplePaths = append(simplePaths, str)
						}
					}
				} else {
					optsArray = optArray
				}
			}

			// Parse complex options
			if len(optsArray) > 0 {
				if optsMap, ok := optsArray[0].(map[string]interface{}); ok {
					// Parse paths
					if paths, ok := optsMap["paths"].([]interface{}); ok {
						for _, p := range paths {
							if pathStr, ok := p.(string); ok {
								simplePaths = append(simplePaths, pathStr)
							} else if pathMap, ok := p.(map[string]interface{}); ok {
								var pc PathConfig
								if name, ok := pathMap["name"].(string); ok {
									pc.Name = name
								}
								if msg, ok := pathMap["message"].(string); ok {
									pc.Message = msg
								}
								if allowType, ok := pathMap["allowTypeImports"].(bool); ok {
									pc.AllowTypeImports = allowType
								}
								if importNames, ok := pathMap["importNames"].([]interface{}); ok {
									for _, in := range importNames {
										if inStr, ok := in.(string); ok {
											pc.ImportNames = append(pc.ImportNames, inStr)
										}
									}
								}
								pathConfigs = append(pathConfigs, pc)
							}
						}
					}

					// Parse patterns
					if patterns, ok := optsMap["patterns"].([]interface{}); ok {
						for _, p := range patterns {
							if patternStr, ok := p.(string); ok {
								// Simple pattern string
								patternConfigs = append(patternConfigs, PatternConfig{
									Group: []string{patternStr},
								})
							} else if patternMap, ok := p.(map[string]interface{}); ok {
								var pc PatternConfig
								if group, ok := patternMap["group"].([]interface{}); ok {
									for _, g := range group {
										if gStr, ok := g.(string); ok {
											pc.Group = append(pc.Group, gStr)
										}
									}
								}
								if msg, ok := patternMap["message"].(string); ok {
									pc.Message = msg
								}
								if caseSens, ok := patternMap["caseSensitive"].(bool); ok {
									pc.CaseSensitive = &caseSens
								}
								if allowType, ok := patternMap["allowTypeImports"].(bool); ok {
									pc.AllowTypeImports = allowType
								}
								if importNames, ok := patternMap["importNames"].([]interface{}); ok {
									for _, in := range importNames {
										if inStr, ok := in.(string); ok {
											pc.ImportNames = append(pc.ImportNames, inStr)
										}
									}
								}
								patternConfigs = append(patternConfigs, pc)
							}
						}
					}
				}
			}
		}

		checkImport := func(node *ast.Node, source string) {
			// Check if it's a type-only import
			isTypeOnly := isTypeImport(node)

			// Check simple paths
			for _, path := range simplePaths {
				if source == path {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "path",
						Description: "'" + source + "' import is restricted from being used.",
					})
					return
				}
			}

			// Check path configurations
			for _, pathConfig := range pathConfigs {
				if source == pathConfig.Name {
					// If allowTypeImports is true and this is a type import, skip
					if pathConfig.AllowTypeImports && isTypeOnly {
						continue
					}

					// Check if specific import names are restricted
					if len(pathConfig.ImportNames) > 0 {
						importedNames := getImportNames(node)
						for _, importedName := range importedNames {
							for _, restrictedName := range pathConfig.ImportNames {
								if importedName == restrictedName {
									msg := pathConfig.Message
									if msg == "" {
										msg = "'" + restrictedName + "' import from '" + source + "' is restricted."
									}
									ctx.ReportNode(node, rule.RuleMessage{
										Id:          "pathWithCustomMessage",
										Description: msg,
									})
									return
								}
							}
						}
					} else {
						msg := pathConfig.Message
						if msg == "" {
							msg = "'" + source + "' import is restricted from being used."
						}
						ctx.ReportNode(node, rule.RuleMessage{
							Id:          "pathWithCustomMessage",
							Description: msg,
						})
						return
					}
				}
			}

			// Check pattern configurations
			for _, patternConfig := range patternConfigs {
				if matchesPatterns(source, patternConfig.Group) {
					// If allowTypeImports is true and this is a type import, skip
					if patternConfig.AllowTypeImports && isTypeOnly {
						continue
					}

					// Check if specific import names are restricted
					if len(patternConfig.ImportNames) > 0 {
						importedNames := getImportNames(node)
						for _, importedName := range importedNames {
							for _, restrictedName := range patternConfig.ImportNames {
								if importedName == restrictedName {
									msg := patternConfig.Message
									if msg == "" {
										msg = "'" + restrictedName + "' import from '" + source + "' is restricted by pattern."
									}
									ctx.ReportNode(node, rule.RuleMessage{
										Id:          "patterns",
										Description: msg,
									})
									return
								}
							}
						}
					} else {
						msg := patternConfig.Message
						if msg == "" {
							msg = "'" + source + "' import is restricted by pattern."
						}
						ctx.ReportNode(node, rule.RuleMessage{
							Id:          "patterns",
							Description: msg,
						})
						return
					}
				}
			}
		}

		return rule.RuleListeners{
			ast.KindImportDeclaration: func(node *ast.Node) {
				importDecl := node.AsImportDeclaration()
				if importDecl.ModuleSpecifier != nil && isStringOrTemplateLiteral(importDecl.ModuleSpecifier) {
					if source, ok := getStaticStringValue(importDecl.ModuleSpecifier); ok {
						checkImport(node, source)
					}
				}
			},

			ast.KindExportDeclaration: func(node *ast.Node) {
				exportDecl := node.AsExportDeclaration()
				// Only check re-exports (export ... from 'module')
				if exportDecl.ModuleSpecifier != nil && isStringOrTemplateLiteral(exportDecl.ModuleSpecifier) {
					if source, ok := getStaticStringValue(exportDecl.ModuleSpecifier); ok {
						checkImport(node, source)
					}
				}
			},

			ast.KindImportEqualsDeclaration: func(node *ast.Node) {
				importEq := node.AsImportEqualsDeclaration()
				if importEq.ModuleReference != nil && importEq.ModuleReference.Kind == ast.KindExternalModuleReference {
					extModRef := importEq.ModuleReference.AsExternalModuleReference()
					if isStringOrTemplateLiteral(extModRef.Expression) {
						if source, ok := getStaticStringValue(extModRef.Expression); ok {
							checkImport(node, source)
						}
					}
				}
			},
		}
	},
})
