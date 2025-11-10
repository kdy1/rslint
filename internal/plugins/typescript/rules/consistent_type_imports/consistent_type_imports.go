package consistent_type_imports

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type ConsistentTypeImportsOptions struct {
	Prefer                  string `json:"prefer"`
	DisallowTypeAnnotations bool   `json:"disallowTypeAnnotations"`
	FixStyle                string `json:"fixStyle"`
}

// ConsistentTypeImportsRule enforces consistent type imports
var ConsistentTypeImportsRule = rule.CreateRule(rule.Rule{
	Name: "consistent-type-imports",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := ConsistentTypeImportsOptions{
		Prefer:                  "type-imports",
		DisallowTypeAnnotations: true,
		FixStyle:                "separate-type-imports",
	}

	// Parse options
	if options != nil {
		if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
			if optMap, ok := optArray[0].(map[string]interface{}); ok {
				if prefer, ok := optMap["prefer"].(string); ok {
					opts.Prefer = prefer
				}
				if disallow, ok := optMap["disallowTypeAnnotations"].(bool); ok {
					opts.DisallowTypeAnnotations = disallow
				}
				if fixStyle, ok := optMap["fixStyle"].(string); ok {
					opts.FixStyle = fixStyle
				}
			}
		} else if optMap, ok := options.(map[string]interface{}); ok {
			if prefer, ok := optMap["prefer"].(string); ok {
				opts.Prefer = prefer
			}
			if disallow, ok := optMap["disallowTypeAnnotations"].(bool); ok {
				opts.DisallowTypeAnnotations = disallow
			}
			if fixStyle, ok := optMap["fixStyle"].(string); ok {
				opts.FixStyle = fixStyle
			}
		}
	}

	// Helper to check if a symbol is type-only
	isSymbolTypeBased := func(symbol *ast.Symbol) *bool {
		if symbol == nil {
			return nil
		}

		// Follow alias chain
		for symbol != nil && (symbol.Flags&ast.SymbolFlagsAlias) != 0 {
			aliased := ctx.TypeChecker.GetAliasedSymbol(symbol)
			if aliased == nil {
				break
			}
			symbol = aliased

			// Check if any declaration in the chain is type-only
			if symbol != nil && symbol.Declarations != nil {
				for _, decl := range symbol.Declarations {
					if decl.IsTypeOnly() {
						trueVal := true
						return &trueVal
					}
				}
			}
		}

		// Check if the symbol is unknown
		if symbol == nil || ctx.TypeChecker.IsUnknownSymbol(symbol) {
			return nil
		}

		// Check if symbol has Value flag - if not, it's type-only
		hasValue := (symbol.Flags & ast.SymbolFlagsValue) != 0
		isType := !hasValue
		return &isType
	}

	// Helper to check if identifier is used only in type positions
	isIdentifierUsedInTypePosition := func(identifierName string, sourceFile *ast.Node) bool {
		if sourceFile == nil {
			return false
		}

		// Track all usages of this identifier
		hasTypeUsage := false
		hasValueUsage := false

		var visitor func(*ast.Node) bool
		visitor = func(node *ast.Node) bool {
			if node == nil {
				return false
			}

			// Skip import declarations themselves
			if node.Kind == ast.KindImportDeclaration {
				return false
			}

			// Check if this is an identifier matching our import
			if node.Kind == ast.KindIdentifier {
				identifier := node.AsIdentifier()
				if identifier != nil && identifier.Text == identifierName {
					// Determine if this usage is in a type position
					parent := node.Parent
					if parent == nil {
						hasValueUsage = true
						return false
					}

					switch parent.Kind {
					// Type positions
					case ast.KindTypeReference,
						ast.KindInterfaceDeclaration,
						ast.KindTypeAliasDeclaration,
						ast.KindTypeParameter,
						ast.KindTypeQuery,
						ast.KindParameter,
						ast.KindPropertySignature,
						ast.KindPropertyDeclaration,
						ast.KindMethodSignature,
						ast.KindMethodDeclaration,
						ast.KindFunctionDeclaration,
						ast.KindArrowFunction,
						ast.KindFunctionExpression:
						hasTypeUsage = true

					// Check if it's in a type export
					case ast.KindExportSpecifier:
						exportSpec := parent.AsExportSpecifier()
						if exportSpec != nil && exportSpec.IsTypeOnly {
							hasTypeUsage = true
						} else {
							// Check if the parent export declaration is type-only
							grandParent := parent.Parent
							if grandParent != nil && grandParent.Parent != nil {
								exportDecl := grandParent.Parent.AsExportDeclaration()
								if exportDecl != nil && exportDecl.IsTypeOnly {
									hasTypeUsage = true
								} else {
									hasValueUsage = true
								}
							} else {
								hasValueUsage = true
							}
						}

					default:
						// Value usage by default
						hasValueUsage = true
					}
				}
			}

			// Visit children
			node.ForEachChild(visitor)
			return false
		}

		sourceFile.ForEachChild(visitor)

		// It's type-only if it has type usage and no value usage
		return hasTypeUsage && !hasValueUsage
	}

	checkImportDeclaration := func(node *ast.Node) {
		importDecl := node.AsImportDeclaration()
		if importDecl == nil {
			return
		}

		importClauseNode := importDecl.ImportClause
		if importClauseNode == nil {
			return
		}

		importClause := importClauseNode.AsImportClause()
		if importClause == nil {
			return
		}

		// Check for prefer: 'no-type-imports' - avoid type imports
		if opts.Prefer == "no-type-imports" {
			if importClause.IsTypeOnly {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "avoidImportType",
					Description: "Use an `import` instead of an `import type`.",
				})
				return
			}

			// Also check for inline type specifiers
			if importClause.NamedBindings != nil && importClause.NamedBindings.Kind == ast.KindNamedImports {
				namedImports := importClause.NamedBindings.AsNamedImports()
				if namedImports != nil {
					for _, element := range namedImports.Elements.Nodes {
						importSpec := element.AsImportSpecifier()
						if importSpec != nil && importSpec.IsTypeOnly {
							ctx.ReportNode(element, rule.RuleMessage{
								Id:          "avoidImportType",
								Description: "Use an `import` instead of an `import type`.",
							})
						}
					}
				}
			}
			return
		}

		// For prefer: 'type-imports' (default)
		// Skip if entire import is already type-only
		if importClause.IsTypeOnly {
			return
		}

		// Get the source file for usage analysis
		sourceFile := ctx.SourceFile
		if sourceFile == nil {
			return
		}

		// Convert SourceFile to Node for usage analysis
		sourceFileNode := sourceFile.AsNode()

		// Track imports to check
		var hasTypeOnlyImports bool = false
		var hasValueImports bool = false
		var typeOnlySpecifiers []*ast.Node
		var valueSpecifiers []*ast.Node
		var inlineTypeSpecifiers []*ast.Node

		// Check default import
		if importClause.Name() != nil {
			identifier := importClause.Name().AsIdentifier()
			if identifier != nil {
				name := identifier.Text

				// Get symbol to check if it's type-based
				symbol := ctx.TypeChecker.GetSymbolAtLocation(importClause.Name())
				isType := isSymbolTypeBased(symbol)

				// Also check actual usage in the file
				isUsedAsTypeOnly := isIdentifierUsedInTypePosition(name, sourceFileNode)

				if isType != nil && *isType || isUsedAsTypeOnly {
					hasTypeOnlyImports = true
					typeOnlySpecifiers = append(typeOnlySpecifiers, importClause.Name())
				} else {
					hasValueImports = true
					valueSpecifiers = append(valueSpecifiers, importClause.Name())
				}
			}
		}

		// Check named imports
		if importClause.NamedBindings != nil && importClause.NamedBindings.Kind == ast.KindNamedImports {
			namedImports := importClause.NamedBindings.AsNamedImports()
			if namedImports != nil {
				for _, element := range namedImports.Elements.Nodes {
					importSpec := element.AsImportSpecifier()
					if importSpec == nil {
						continue
					}

					// Check if already has inline type specifier
					if importSpec.IsTypeOnly {
						inlineTypeSpecifiers = append(inlineTypeSpecifiers, element)
						hasTypeOnlyImports = true
						continue
					}

					// Get the imported name
					var name string
					if importSpec.PropertyName != nil {
						if id := importSpec.PropertyName.AsIdentifier(); id != nil {
							name = id.Text
						}
					} else {
						if id := importSpec.Name().AsIdentifier(); id != nil {
							name = id.Text
						}
					}

					if name == "" {
						continue
					}

					// Get symbol and check if type-based
					symbol := ctx.TypeChecker.GetSymbolAtLocation(element)
					isType := isSymbolTypeBased(symbol)

					// Also check actual usage
					isUsedAsTypeOnly := isIdentifierUsedInTypePosition(name, sourceFileNode)

					if isType != nil && *isType || isUsedAsTypeOnly {
						hasTypeOnlyImports = true
						typeOnlySpecifiers = append(typeOnlySpecifiers, element)
					} else {
						hasValueImports = true
						valueSpecifiers = append(valueSpecifiers, element)
					}
				}
			}
		}

		// Check namespace imports
		if importClause.NamedBindings != nil && importClause.NamedBindings.Kind == ast.KindNamespaceImport {
			namespaceImport := importClause.NamedBindings.AsNamespaceImport()
			if namespaceImport != nil && namespaceImport.Name() != nil {
				identifier := namespaceImport.Name().AsIdentifier()
				if identifier != nil {
					name := identifier.Text

					symbol := ctx.TypeChecker.GetSymbolAtLocation(namespaceImport.Name())
					isType := isSymbolTypeBased(symbol)
					isUsedAsTypeOnly := isIdentifierUsedInTypePosition(name, sourceFileNode)

					if isType != nil && *isType || isUsedAsTypeOnly {
						hasTypeOnlyImports = true
						typeOnlySpecifiers = append(typeOnlySpecifiers, namespaceImport.Name())
					} else {
						hasValueImports = true
						valueSpecifiers = append(valueSpecifiers, namespaceImport.Name())
					}
				}
			}
		}

		// Report if all imports are type-only but not marked as such
		if hasTypeOnlyImports && !hasValueImports && len(inlineTypeSpecifiers) == 0 {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "typeOverValue",
				Description: "All imports in the declaration are only used as types. Use `import type`.",
			})
			return
		}

		// Report if some imports are type-only (mixed case)
		if hasTypeOnlyImports && hasValueImports {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "someImportsAreOnlyTypes",
				Description: "Some imports are only used as types.",
			})
		}
	}

	checkTSImportType := func(node *ast.Node) {
		if opts.DisallowTypeAnnotations {
			// Check if this is an import type in a type annotation position
			importType := node.AsImportTypeNode()
			if importType != nil {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "noImportTypeAnnotations",
					Description: "`import()` type annotations are forbidden.",
				})
			}
		}
	}

	listeners := rule.RuleListeners{
		ast.KindImportDeclaration: checkImportDeclaration,
	}

	// Only add import type listener if disallowTypeAnnotations is enabled
	if opts.DisallowTypeAnnotations {
		listeners[ast.KindImportType] = checkTSImportType
	}

	return listeners
}
