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
	originalText := sourceText[node.Pos():node.End()]

	// Get the name
	name := ""
	nameNode := typeAlias.Name()
	if nameNode != nil {
		name = sourceText[nameNode.Pos():nameNode.End()]
	}

	// Get type parameters if any
	typeParams := ""
	if typeAlias.TypeParameters != nil {
		typeParams = sourceText[typeAlias.TypeParameters.Pos():typeAlias.TypeParameters.End()]
	}

	// Get the type body (the object literal) preserving exact whitespace
	typeBody := ""
	if typeAlias.Type != nil {
		typeBody = sourceText[typeAlias.Type.Pos():typeAlias.Type.End()]
	}

	// Get export and declare modifiers if any
	prefix := ""
	modStart := node.Pos()
	if modifiers := typeAlias.Modifiers(); modifiers != nil && len(modifiers.Nodes) > 0 {
		lastMod := modifiers.Nodes[len(modifiers.Nodes)-1]
		prefix = sourceText[node.Pos():lastMod.End()]
		// Include space after modifiers
		modStart = lastMod.End()
	}

	// Extract whitespace between parts from original text
	// Find "type" keyword position in original
	typeKeywordStart := strings.Index(originalText[modStart-node.Pos():], "type")
	if typeKeywordStart == -1 {
		return prefix + "interface " + name + typeParams + " " + typeBody
	}

	// Get whitespace between "type" and name
	afterTypeKeyword := modStart - node.Pos() + typeKeywordStart + 4 // 4 = len("type")
	whitespaceAfterType := sourceText[node.Pos()+afterTypeKeyword : nameNode.Pos()]

	// Get whitespace after name/type params (before the {)
	// We want to preserve this whitespace
	whitespaceBeforeBody := " " // default single space
	if nameNode != nil {
		afterNamePos := nameNode.End()
		if typeAlias.TypeParameters != nil {
			afterNamePos = typeAlias.TypeParameters.End()
		}
		// Find the "{" in the type body
		// The whitespace we want is after name/params but before "="
		// Then after "=", we want the whitespace before "{"
		textBetween := sourceText[afterNamePos:typeAlias.Type.Pos()]
		equalPos := strings.Index(textBetween, "=")
		if equalPos != -1 {
			// Get whitespace after "=" and before the type body
			afterEqual := textBetween[equalPos+1:]
			whitespaceBeforeBody = strings.TrimLeft(afterEqual, " \t")
			// If there's whitespace, we'll use a single space
			if len(whitespaceBeforeBody) < len(afterEqual) {
				whitespaceBeforeBody = " "
			}
		}
	}

	// Construct the result
	result := prefix
	if prefix != "" {
		result += " "
	}
	result += "interface" + whitespaceAfterType + name + typeParams + whitespaceBeforeBody + typeBody

	return result
}

// convertInterfaceToType converts an interface declaration to a type alias
func convertInterfaceToType(ctx rule.RuleContext, node *ast.Node) string {
	interfaceDecl := node.AsInterfaceDeclaration()
	if interfaceDecl == nil {
		return ""
	}

	sourceText := ctx.SourceFile.Text()
	originalText := sourceText[node.Pos():node.End()]

	// Get the name
	name := ""
	nameNode := interfaceDecl.Name()
	if nameNode != nil {
		name = sourceText[nameNode.Pos():nameNode.End()]
	}

	// Get type parameters if any
	typeParams := ""
	if interfaceDecl.TypeParameters != nil {
		typeParams = sourceText[interfaceDecl.TypeParameters.Pos():interfaceDecl.TypeParameters.End()]
	}

	// Get the body - extract from opening brace to closing brace
	bodyText := ""
	if len(interfaceDecl.Members.Nodes) > 0 {
		// Get from opening brace to closing brace
		firstMember := interfaceDecl.Members.Nodes[0]
		lastMember := interfaceDecl.Members.Nodes[len(interfaceDecl.Members.Nodes)-1]
		// Find the { before first member and } after last member
		// Look backwards from first member to find {
		textBeforeFirst := sourceText[node.Pos():firstMember.Pos()]
		openBracePos := strings.LastIndex(textBeforeFirst, "{")
		if openBracePos != -1 {
			openBraceAbsPos := node.Pos() + openBracePos
			// Find } after last member
			textAfterLast := sourceText[lastMember.End():node.End()]
			closeBracePos := strings.Index(textAfterLast, "}")
			if closeBracePos != -1 {
				closeBraceAbsPos := lastMember.End() + closeBracePos
				bodyText = sourceText[openBraceAbsPos : closeBraceAbsPos+1]
			}
		}
	} else {
		// Empty interface - find the {}
		nodeText := sourceText[node.Pos():node.End()]
		openBrace := strings.Index(nodeText, "{")
		closeBrace := strings.LastIndex(nodeText, "}")
		if openBrace != -1 && closeBrace != -1 {
			bodyText = nodeText[openBrace : closeBrace+1]
		}
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
		if bodyText != "" {
			heritage += " & " + bodyText
		} else {
			heritage = "{ " + heritage + " }"
		}
	} else {
		heritage = bodyText
	}

	// Get export and declare modifiers if any
	prefix := ""
	modStart := node.Pos()
	if modifiers := interfaceDecl.Modifiers(); modifiers != nil && len(modifiers.Nodes) > 0 {
		lastMod := modifiers.Nodes[len(modifiers.Nodes)-1]
		prefix = sourceText[node.Pos():lastMod.End()]
		modStart = lastMod.End()
	}

	// Extract whitespace between parts from original text
	// Find "interface" keyword position
	interfaceKeywordStart := strings.Index(originalText[modStart-node.Pos():], "interface")
	if interfaceKeywordStart == -1 {
		return prefix + " type " + name + typeParams + " = " + heritage
	}

	// Get whitespace between "interface" and name
	afterInterfaceKeyword := modStart - node.Pos() + interfaceKeywordStart + 9 // 9 = len("interface")
	whitespaceAfterInterface := sourceText[node.Pos()+afterInterfaceKeyword : nameNode.Pos()]

	// Get whitespace after name (or type parameters)
	whitespaceAfterName := ""
	if nameNode != nil {
		afterNamePos := nameNode.End()
		if interfaceDecl.TypeParameters != nil {
			afterNamePos = interfaceDecl.TypeParameters.End()
		}
		// Find the opening brace
		nodeText := sourceText[node.Pos():node.End()]
		openBracePos := strings.Index(nodeText, "{")
		if openBracePos != -1 {
			openBraceAbsPos := node.Pos() + openBracePos
			whitespaceAfterName = strings.TrimRight(sourceText[afterNamePos:openBraceAbsPos], " \t")
		}
	}

	// Construct the result
	result := prefix
	if prefix != "" {
		result += " "
	}
	result += "type" + whitespaceAfterInterface + name + typeParams + whitespaceAfterName + " = " + heritage

	return result
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
