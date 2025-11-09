package no_empty_object_type

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type NoEmptyObjectTypeOptions struct {
	AllowInterfaces  string `json:"allowInterfaces"`
	AllowObjectTypes string `json:"allowObjectTypes"`
	AllowWithName    string `json:"allowWithName"`
}

func parseOptions(options any) NoEmptyObjectTypeOptions {
	opts := NoEmptyObjectTypeOptions{
		AllowInterfaces:  "never",
		AllowObjectTypes: "never",
		AllowWithName:    "",
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
			if allowInterfaces, ok := optsMap["allowInterfaces"].(string); ok {
				opts.AllowInterfaces = allowInterfaces
			}
			if allowObjectTypes, ok := optsMap["allowObjectTypes"].(string); ok {
				opts.AllowObjectTypes = allowObjectTypes
			}
			if allowWithName, ok := optsMap["allowWithName"].(string); ok {
				opts.AllowWithName = allowWithName
			}
		}
	}

	return opts
}

func isEmptyTypeLiteral(node *ast.Node) bool {
	if !ast.IsTypeLiteralNode(node) {
		return false
	}
	typeLiteral := node.AsTypeLiteralNode()
	if typeLiteral == nil {
		return false
	}
	return typeLiteral.Members == nil || len(typeLiteral.Members.Nodes) == 0
}

func isInIntersectionType(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	return ast.IsIntersectionTypeNode(parent)
}

func matchesAllowedName(name string, pattern string) bool {
	if pattern == "" {
		return false
	}
	matched, err := regexp.MatchString(pattern, name)
	if err != nil {
		return false
	}
	return matched
}

var NoEmptyObjectTypeRule = rule.CreateRule(rule.Rule{
	Name: "no-empty-object-type",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			ast.KindInterfaceDeclaration: func(node *ast.Node) {
				interfaceDecl := node.AsInterfaceDeclaration()
				if interfaceDecl == nil {
					return
				}

				// Check if interface has members
				if interfaceDecl.Members != nil && len(interfaceDecl.Members.Nodes) > 0 {
					return
				}

				// Extract interface name for allowWithName check
				nameRange := utils.TrimNodeTextRange(ctx.SourceFile, interfaceDecl.Name())
				nameText := ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()]

				if matchesAllowedName(nameText, opts.AllowWithName) {
					return
				}

				// Count extended interfaces
				extendCount := 0
				var extendClause *ast.HeritageClause
				if interfaceDecl.HeritageClauses != nil {
					for _, clause := range interfaceDecl.HeritageClauses.Nodes {
						heritageClause := clause.AsHeritageClause()
						if heritageClause == nil {
							continue
						}
						if heritageClause.Token == ast.KindExtendsKeyword {
							extendClause = heritageClause
							extendCount = len(heritageClause.Types.Nodes)
							break
						}
					}
				}

				// Empty interface with multiple extends is allowed (union type alternative)
				if extendCount > 1 {
					return
				}

				// Handle empty interface with single extend
				if extendCount == 1 {
					// Check allowInterfaces setting
					if opts.AllowInterfaces == "always" || opts.AllowInterfaces == "with-single-extends" {
						return
					}

					// Check for merged class declaration
					mergedWithClassDeclaration := false
					if ctx.TypeChecker != nil {
						symbol := ctx.TypeChecker.GetSymbolAtLocation(interfaceDecl.Name())
						if symbol != nil {
							for _, decl := range symbol.Declarations {
								if decl.Kind == ast.KindClassDeclaration {
									mergedWithClassDeclaration = true
									break
								}
							}
						}
					}

					// Check if in ambient declaration (.d.ts file)
					isInAmbientDeclaration := false
					if strings.HasSuffix(ctx.SourceFile.FileName(), ".d.ts") {
						parent := node.Parent
						for parent != nil {
							if parent.Kind == ast.KindModuleDeclaration {
								moduleDecl := parent.AsModuleDeclaration()
								if moduleDecl == nil {
									parent = parent.Parent
									continue
								}
								modifiers := moduleDecl.Modifiers()
								if modifiers != nil {
									for _, modifier := range modifiers.Nodes {
										if modifier.Kind == ast.KindDeclareKeyword {
											isInAmbientDeclaration = true
											break
										}
									}
								}
							}
							if isInAmbientDeclaration {
								break
							}
							parent = parent.Parent
						}
					}

					message := rule.RuleMessage{
						Id:          "noEmptyInterfaceWithSuper",
						Description: "An empty interface extending a single type is equivalent to a type alias.",
					}

					// Check for export modifier
					var exportText string
					if interfaceDecl.Modifiers() != nil {
						for _, modifier := range interfaceDecl.Modifiers().Nodes {
							if modifier.Kind == ast.KindExportKeyword {
								exportText = "export "
								break
							}
						}
					}

					// Extract type parameters if present
					var typeParamsText string
					if interfaceDecl.TypeParameters != nil && len(interfaceDecl.TypeParameters.Nodes) > 0 {
						firstParam := interfaceDecl.TypeParameters.Nodes[0]
						lastParam := interfaceDecl.TypeParameters.Nodes[len(interfaceDecl.TypeParameters.Nodes)-1]
						firstRange := utils.TrimNodeTextRange(ctx.SourceFile, firstParam)
						lastRange := utils.TrimNodeTextRange(ctx.SourceFile, lastParam)
						typeParamsRange := firstRange.WithEnd(lastRange.End())
						typeParamsRange = typeParamsRange.WithPos(typeParamsRange.Pos() - 1).WithEnd(typeParamsRange.End() + 1)
						typeParamsText = ctx.SourceFile.Text()[typeParamsRange.Pos():typeParamsRange.End()]
					}

					extendedTypeRange := utils.TrimNodeTextRange(ctx.SourceFile, extendClause.Types.Nodes[0])
					extendedTypeText := ctx.SourceFile.Text()[extendedTypeRange.Pos():extendedTypeRange.End()]

					replacement := fmt.Sprintf("%stype %s%s = %s", exportText, nameText, typeParamsText, extendedTypeText)

					if isInAmbientDeclaration || mergedWithClassDeclaration {
						ctx.ReportNode(interfaceDecl.Name(), message)
					} else {
						ctx.ReportNodeWithSuggestions(interfaceDecl.Name(), message,
							rule.RuleSuggestion{
								Message: rule.RuleMessage{
									Id:          "replaceEmptyInterfaceWithSuper",
									Description: "Replace empty interface with a type alias.",
								},
								FixesArr: []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, replacement)},
							})
					}
					return
				}

				// Empty interface with no extends
				if opts.AllowInterfaces == "always" {
					return
				}

				message := rule.RuleMessage{
					Id:          "noEmptyInterface",
					Description: fmt.Sprintf("The %s type (empty interface type) allows any non-nullish value. If this is intentional, use `object` instead. Otherwise, use `unknown`.", "{}"),
				}

				// Create suggestions for object and unknown
				suggestions := []rule.RuleSuggestion{
					{
						Message: rule.RuleMessage{
							Id:          "replaceEmptyInterface",
							Description: fmt.Sprintf("Replace empty interface with %s.", "`object`"),
						},
						FixesArr: []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, fmt.Sprintf("type %s = object", nameText))},
					},
					{
						Message: rule.RuleMessage{
							Id:          "replaceEmptyInterface",
							Description: fmt.Sprintf("Replace empty interface with %s.", "`unknown`"),
						},
						FixesArr: []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, fmt.Sprintf("type %s = unknown", nameText))},
					},
				}

				ctx.ReportNodeWithSuggestions(interfaceDecl.Name(), message, suggestions...)
			},

			ast.KindTypeLiteral: func(node *ast.Node) {
				if !isEmptyTypeLiteral(node) {
					return
				}

				// Allow empty object type in intersection types (e.g., T & {})
				if isInIntersectionType(node) {
					return
				}

				// Check if this is part of a type alias to apply allowWithName
				var typeAliasName string
				parent := node.Parent
				if parent != nil && ast.IsTypeAliasDeclaration(parent) {
					typeAlias := parent.AsTypeAliasDeclaration()
					if typeAlias != nil {
						nameRange := utils.TrimNodeTextRange(ctx.SourceFile, typeAlias.Name())
						typeAliasName = ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()]
					}
				}

				if typeAliasName != "" && matchesAllowedName(typeAliasName, opts.AllowWithName) {
					return
				}

				if opts.AllowObjectTypes == "always" {
					return
				}

				message := rule.RuleMessage{
					Id:          "noEmptyObject",
					Description: fmt.Sprintf("The %s type (empty object type) allows any non-nullish value. If this is intentional, use `object` instead. Otherwise, use `unknown`.", "{}"),
				}

				suggestions := []rule.RuleSuggestion{
					{
						Message: rule.RuleMessage{
							Id:          "replaceEmptyObjectType",
							Description: fmt.Sprintf("Replace %s with %s.", "{}", "`object`"),
						},
						FixesArr: []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, "object")},
					},
					{
						Message: rule.RuleMessage{
							Id:          "replaceEmptyObjectType",
							Description: fmt.Sprintf("Replace %s with %s.", "{}", "`unknown`"),
						},
						FixesArr: []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, "unknown")},
					},
				}

				ctx.ReportNodeWithSuggestions(node, message, suggestions...)
			},
		}
	},
})
