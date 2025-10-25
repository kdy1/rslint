package consistent_type_definitions

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// ConsistentTypeDefinitionsOptions defines the configuration options for this rule
type ConsistentTypeDefinitionsOptions struct {
	Prefer string `json:"prefer"` // "interface" or "type"
}

// parseOptions parses and validates the rule options
func parseOptions(options any) ConsistentTypeDefinitionsOptions {
	opts := ConsistentTypeDefinitionsOptions{
		Prefer: "interface", // default to interface
	}

	if options == nil {
		return opts
	}

	// Handle both array format ["interface"] and object format { prefer: "interface" }
	if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
		if str, ok := optArray[0].(string); ok {
			opts.Prefer = str
		} else if optsMap, ok := optArray[0].(map[string]interface{}); ok {
			if v, ok := optsMap["prefer"].(string); ok {
				opts.Prefer = v
			}
		}
	} else if optsMap, ok := options.(map[string]interface{}); ok {
		if v, ok := optsMap["prefer"].(string); ok {
			opts.Prefer = v
		}
	}

	return opts
}

// isObjectType checks if a type node represents an object type
func isObjectType(typeNode *ast.Node) bool {
	if typeNode == nil {
		return false
	}
	return typeNode.Kind == ast.KindTypeLiteral
}

// convertTypeToInterface converts a type alias declaration to an interface
func convertTypeToInterface(ctx rule.RuleContext, node *ast.Node) string {
	typeAlias := node.AsTypeAliasDeclaration()
	if typeAlias == nil {
		return ""
	}

	sourceText := ctx.SourceFile.Text()

	// Get the name
	name := ""
	if nameNode := typeAlias.Name(); nameNode != nil {
		name = sourceText[nameNode.Pos():nameNode.End()]
	}

	// Get type parameters if any
	typeParams := ""
	if typeAlias.TypeParameters != nil {
		typeParams = sourceText[typeAlias.TypeParameters.Pos():typeAlias.TypeParameters.End()]
	}

	// Get the type body (the object literal)
	typeBody := ""
	if typeAlias.Type != nil {
		typeBody = sourceText[typeAlias.Type.Pos():typeAlias.Type.End()]
	}

	// Get export and declare modifiers if any
	prefix := ""
	if modifiers := typeAlias.Modifiers(); modifiers != nil {
		for _, mod := range modifiers.Nodes {
			if mod.Kind == ast.KindExportKeyword || mod.Kind == ast.KindDeclareKeyword {
				prefix += sourceText[mod.Pos():mod.End()] + " "
			}
		}
	}

	return prefix + "interface " + name + typeParams + " " + typeBody
}

// convertInterfaceToType converts an interface declaration to a type alias
func convertInterfaceToType(ctx rule.RuleContext, node *ast.Node) string {
	interfaceDecl := node.AsInterfaceDeclaration()
	if interfaceDecl == nil {
		return ""
	}

	sourceText := ctx.SourceFile.Text()

	// Get the name
	name := ""
	if nameNode := interfaceDecl.Name(); nameNode != nil {
		name = sourceText[nameNode.Pos():nameNode.End()]
	}

	// Get type parameters if any
	typeParams := ""
	if interfaceDecl.TypeParameters != nil {
		typeParams = sourceText[interfaceDecl.TypeParameters.Pos():interfaceDecl.TypeParameters.End()]
	}

	// Get the body
	body := ""
	if len(interfaceDecl.Members.Nodes) > 0 {
		firstMember := interfaceDecl.Members.Nodes[0]
		lastMember := interfaceDecl.Members.Nodes[len(interfaceDecl.Members.Nodes)-1]
		body = sourceText[firstMember.Pos():lastMember.End()]
	}

	// Get heritage clause (extends)
	heritage := ""
	if interfaceDecl.HeritageClauses != nil && len(interfaceDecl.HeritageClauses.Nodes) > 0 {
		// For interfaces with extends, we need to convert to intersection type
		for i, clause := range interfaceDecl.HeritageClauses.Nodes {
			heritageClause := clause.AsHeritageClause()
			if heritageClause != nil && len(heritageClause.Types.Nodes) > 0 {
				for j, typeNode := range heritageClause.Types.Nodes {
					if i > 0 || j > 0 {
						heritage += " & "
					}
					heritage += sourceText[typeNode.Pos():typeNode.End()]
				}
			}
		}
		if body != "" {
			heritage += " & { " + body + " }"
		} else {
			heritage = "{ " + heritage + " }"
		}
	} else {
		heritage = "{ " + body + " }"
	}

	// Get export and declare modifiers if any
	prefix := ""
	if modifiers := interfaceDecl.Modifiers(); modifiers != nil {
		for _, mod := range modifiers.Nodes {
			if mod.Kind == ast.KindExportKeyword || mod.Kind == ast.KindDeclareKeyword {
				prefix += sourceText[mod.Pos():mod.End()] + " "
			}
		}
	}

	return prefix + "type " + name + typeParams + " = " + heritage
}

// ConsistentTypeDefinitionsRule implements the consistent-type-definitions rule
// Enforce type definitions with interface or type
var ConsistentTypeDefinitionsRule = rule.CreateRule(rule.Rule{
	Name: "consistent-type-definitions",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := parseOptions(options)

	return rule.RuleListeners{
		ast.KindInterfaceDeclaration: func(node *ast.Node) {
			// If prefer is "type", convert interface to type
			if opts.Prefer == "type" {
				interfaceDecl := node.AsInterfaceDeclaration()
				if interfaceDecl == nil {
					return
				}

				// Skip interfaces in declare global blocks (they can't be converted)
				parent := node.Parent
				if parent != nil && parent.Kind == ast.KindModuleBlock {
					grandparent := parent.Parent
					if grandparent != nil && grandparent.Kind == ast.KindModuleDeclaration {
						moduleDecl := grandparent.AsModuleDeclaration()
						if moduleDecl != nil {
							// Check if it's a global augmentation
							if moduleName := moduleDecl.Name(); moduleName != nil && moduleName.Kind == ast.KindIdentifier {
								sourceText := ctx.SourceFile.Text()
								moduleNameText := sourceText[moduleName.Pos():moduleName.End()]
								if strings.Contains(moduleNameText, "global") {
									return // Skip global augmentations
								}
							}
						}
					}
				}

				fixedCode := convertInterfaceToType(ctx, node)
				if fixedCode != "" {
					ctx.ReportNodeWithFixes(node, rule.RuleMessage{
						Id:          "interfaceOverType",
						Description: "Use type instead of interface",
					}, rule.RuleFixReplace(ctx.SourceFile, node, fixedCode))
				}
			}
		},
		ast.KindTypeAliasDeclaration: func(node *ast.Node) {
			// If prefer is "interface", convert object type to interface
			if opts.Prefer == "interface" {
				typeAlias := node.AsTypeAliasDeclaration()
				if typeAlias == nil || typeAlias.Type == nil {
					return
				}

				// Only convert if the type is an object literal
				if !isObjectType(typeAlias.Type) {
					return
				}

				fixedCode := convertTypeToInterface(ctx, node)
				if fixedCode != "" {
					ctx.ReportNodeWithFixes(node, rule.RuleMessage{
						Id:          "typeOverInterface",
						Description: "Use interface instead of type",
					}, rule.RuleFixReplace(ctx.SourceFile, node, fixedCode))
				}
			}
		},
	}
}
