package no_empty_object_type

import (
	"regexp"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type NoEmptyObjectTypeOptions struct {
	AllowInterfaces  string `json:"allowInterfaces"`  // 'never' | 'always' | 'with-single-extends'
	AllowObjectTypes string `json:"allowObjectTypes"` // 'never' | 'always'
	AllowWithName    string `json:"allowWithName"`    // regex pattern
}

func parseOptions(options any) NoEmptyObjectTypeOptions {
	opts := NoEmptyObjectTypeOptions{
		AllowInterfaces:  "never",
		AllowObjectTypes: "never",
		AllowWithName:    "",
	}

	if options == nil {
		return opts
	}

	// Handle array format: ["error", { allowInterfaces: 'always' }]
	if arr, ok := options.([]any); ok && len(arr) > 0 {
		if objMap, ok := arr[0].(map[string]any); ok {
			if val, ok := objMap["allowInterfaces"].(string); ok {
				opts.AllowInterfaces = val
			}
			if val, ok := objMap["allowObjectTypes"].(string); ok {
				opts.AllowObjectTypes = val
			}
			if val, ok := objMap["allowWithName"].(string); ok {
				opts.AllowWithName = val
			}
		}
		return opts
	}

	// Handle object format: { allowInterfaces: 'always' }
	if objMap, ok := options.(map[string]any); ok {
		if val, ok := objMap["allowInterfaces"].(string); ok {
			opts.AllowInterfaces = val
		}
		if val, ok := objMap["allowObjectTypes"].(string); ok {
			opts.AllowObjectTypes = val
		}
		if val, ok := objMap["allowWithName"].(string); ok {
			opts.AllowWithName = val
		}
	}

	return opts
}

// NoEmptyObjectTypeRule implements the no-empty-object-type rule
// Disallow empty interfaces and empty object types
var NoEmptyObjectTypeRule = rule.CreateRule(rule.Rule{
	Name: "no-empty-object-type",
	Run:  run,
})

func buildNoEmptyInterfaceMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noEmptyInterface",
		Description: "An empty interface is equivalent to `{}`, which accepts any non-nullish value.",
	}
}

func buildNoEmptyInterfaceWithSuperMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noEmptyInterfaceWithSuper",
		Description: "An interface extending a single other interface without adding members is equivalent to the base type itself.",
	}
}

func buildNoEmptyObjectMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noEmptyObject",
		Description: "The `{}` (empty object) type accepts any non-nullish value, which is likely not what you meant. Use `object` for any object or `unknown` for any value.",
	}
}

// Check if a name matches the allowWithName pattern
func matchesPattern(name string, pattern string) bool {
	if pattern == "" {
		return false
	}

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}

	return regex.MatchString(name)
}

// Check if an interface is merged with a class
func isMergedWithClass(interfaceName string, sourceFile *ast.SourceFile) bool {
	// Walk through all statements to see if there's a class with the same name
	// For simplicity, we'll skip this check for now and always generate suggestions
	// A full implementation would traverse the AST to find class declarations
	return false
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := parseOptions(options)

	return rule.RuleListeners{
		// Check interface declarations
		ast.KindInterfaceDeclaration: func(node *ast.Node) {
			if opts.AllowInterfaces == "always" {
				return
			}

			interfaceDecl := node.AsInterfaceDeclaration()

			// Get the interface name
			interfaceName := ""
			if interfaceDecl.Name() != nil {
				nameRange := utils.TrimNodeTextRange(ctx.SourceFile, interfaceDecl.Name())
				interfaceName = ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()]
			}

			// Check if name matches the allowWithName pattern
			if opts.AllowWithName != "" && matchesPattern(interfaceName, opts.AllowWithName) {
				return
			}

			// Check if interface has members
			if interfaceDecl.Members != nil && len(interfaceDecl.Members.Nodes) > 0 {
				return
			}

			// Get heritage clauses (extends)
			extendsCount := 0
			var firstExtends *ast.Node

			if interfaceDecl.HeritageClauses != nil {
				for _, clause := range interfaceDecl.HeritageClauses.Nodes {
					heritageClause := clause.AsHeritageClause()
					if heritageClause.Token == ast.KindExtendsKeyword {
						if heritageClause.Types != nil {
							extendsCount = len(heritageClause.Types.Nodes)
							if extendsCount > 0 {
								firstExtends = heritageClause.Types.Nodes[0]
							}
						}
					}
				}
			}

			// If interface extends multiple types, it's allowed (intersection type)
			if extendsCount > 1 {
				return
			}

			// If allowInterfaces is 'with-single-extends' and it extends exactly one type
			if opts.AllowInterfaces == "with-single-extends" && extendsCount == 1 {
				return
			}

			// Case 1: Empty interface with no extends
			if extendsCount == 0 {
				message := buildNoEmptyInterfaceMessage()

				// Check if merged with a class
				if !isMergedWithClass(interfaceName, ctx.SourceFile) {
					// Provide suggestions to replace with type alias
					typeParams := ""
					if interfaceDecl.TypeParameters != nil && len(interfaceDecl.TypeParameters.Nodes) > 0 {
						firstParam := interfaceDecl.TypeParameters.Nodes[0]
						lastParam := interfaceDecl.TypeParameters.Nodes[len(interfaceDecl.TypeParameters.Nodes)-1]
						firstRange := utils.TrimNodeTextRange(ctx.SourceFile, firstParam)
						lastRange := utils.TrimNodeTextRange(ctx.SourceFile, lastParam)
						typeParamsRange := firstRange.WithEnd(lastRange.End())
						typeParamsRange = typeParamsRange.WithPos(typeParamsRange.Pos() - 1).WithEnd(typeParamsRange.End() + 1)
						typeParams = ctx.SourceFile.Text()[typeParamsRange.Pos():typeParamsRange.End()]
					}

					suggestion1 := rule.RuleSuggestion{
						Message: rule.RuleMessage{
							Id:   "replaceEmptyInterface",
						Description: "Replace with `type " + interfaceName + typeParams + " = object`",
						},
						FixesArr: []rule.RuleFix{
							rule.RuleFixReplace(ctx.SourceFile, node, "type "+interfaceName+typeParams+" = object"),
						},
					}

					suggestion2 := rule.RuleSuggestion{
						Message: rule.RuleMessage{
							Id:   "replaceEmptyInterface",
						Description: "Replace with `type " + interfaceName + typeParams + " = unknown`",
						},
						FixesArr: []rule.RuleFix{
							rule.RuleFixReplace(ctx.SourceFile, node, "type "+interfaceName+typeParams+" = unknown"),
						},
					}

					ctx.ReportNodeWithSuggestions(interfaceDecl.Name(), message, suggestion1, suggestion2)
				} else {
					ctx.ReportNode(interfaceDecl.Name(), message)
				}
				return
			}

			// Case 2: Empty interface with single extends
			if extendsCount == 1 && firstExtends != nil {
				message := buildNoEmptyInterfaceWithSuperMessage()

				// Check if merged with a class
				if !isMergedWithClass(interfaceName, ctx.SourceFile) {
					// Get the base type text
					baseTypeRange := utils.TrimNodeTextRange(ctx.SourceFile, firstExtends)
					baseType := ctx.SourceFile.Text()[baseTypeRange.Pos():baseTypeRange.End()]

					typeParams := ""
					if interfaceDecl.TypeParameters != nil && len(interfaceDecl.TypeParameters.Nodes) > 0 {
						firstParam := interfaceDecl.TypeParameters.Nodes[0]
						lastParam := interfaceDecl.TypeParameters.Nodes[len(interfaceDecl.TypeParameters.Nodes)-1]
						firstRange := utils.TrimNodeTextRange(ctx.SourceFile, firstParam)
						lastRange := utils.TrimNodeTextRange(ctx.SourceFile, lastParam)
						typeParamsRange := firstRange.WithEnd(lastRange.End())
						typeParamsRange = typeParamsRange.WithPos(typeParamsRange.Pos() - 1).WithEnd(typeParamsRange.End() + 1)
						typeParams = ctx.SourceFile.Text()[typeParamsRange.Pos():typeParamsRange.End()]
					}

					suggestion := rule.RuleSuggestion{
						Message: rule.RuleMessage{
							Id:   "replaceEmptyInterfaceWithSuper",
						Description: "Replace with `type " + interfaceName + typeParams + " = " + baseType + "`",
						},
						FixesArr: []rule.RuleFix{
							rule.RuleFixReplace(ctx.SourceFile, node, "type "+interfaceName+typeParams+" = "+baseType),
						},
					}

					ctx.ReportNodeWithSuggestions(interfaceDecl.Name(), message, suggestion)
				} else {
					ctx.ReportNode(interfaceDecl.Name(), message)
				}
			}
		},

		// Check type literal (empty object type: {})
		ast.KindTypeLiteral: func(node *ast.Node) {
			if opts.AllowObjectTypes == "always" {
				return
			}

			typeLiteral := node.AsTypeLiteralNode()

			// Check if it has members
			if typeLiteral.Members != nil && len(typeLiteral.Members.Nodes) > 0 {
				return
			}

			// Check if parent is an intersection type (& {})
			parent := node.Parent
			if parent != nil && parent.Kind == ast.KindIntersectionType {
				return
			}

			// Check if parent is a type alias and matches allowWithName
			if parent != nil && parent.Kind == ast.KindTypeAliasDeclaration {
				typeAlias := parent.AsTypeAliasDeclaration()
				if typeAlias.Name() != nil {
					nameRange := utils.TrimNodeTextRange(ctx.SourceFile, typeAlias.Name())
					typeName := ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()]
					if opts.AllowWithName != "" && matchesPattern(typeName, opts.AllowWithName) {
						return
					}
				}
			}

			message := buildNoEmptyObjectMessage()

			// Provide suggestions to replace with object or unknown
			suggestion1 := rule.RuleSuggestion{
				Message: rule.RuleMessage{
							Id:   "replaceEmptyObjectType",
				Description: "Replace `{}` with `object`",
				},
						FixesArr: []rule.RuleFix{
					rule.RuleFixReplace(ctx.SourceFile, node, "object"),
				},
			}

			suggestion2 := rule.RuleSuggestion{
				Message: rule.RuleMessage{
							Id:   "replaceEmptyObjectType",
				Description: "Replace `{}` with `unknown`",
				},
						FixesArr: []rule.RuleFix{
					rule.RuleFixReplace(ctx.SourceFile, node, "unknown"),
				},
			}

			ctx.ReportNodeWithSuggestions(node, message, suggestion1, suggestion2)
		},
	}
}
