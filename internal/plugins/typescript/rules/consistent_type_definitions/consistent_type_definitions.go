package consistent_type_definitions

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// ConsistentTypeDefinitionsOptions defines the configuration
type ConsistentTypeDefinitionsOptions struct {
	Style string `json:"style"` // "interface" or "type"
}

func parseOptions(options interface{}) ConsistentTypeDefinitionsOptions {
	opts := ConsistentTypeDefinitionsOptions{
		Style: "interface", // Default
	}

	if options == nil {
		return opts
	}

	switch v := options.(type) {
	case string:
		if v == "type" || v == "interface" {
			opts.Style = v
		}
	case map[string]interface{}:
		if style, ok := v["style"].(string); ok {
			if style == "type" || style == "interface" {
				opts.Style = style
			}
		}
	}

	return opts
}

func buildPreferInterfaceMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "interfaceOverType",
		Description: "Use an 'interface' instead of a 'type' for object type definitions.",
	}
}

func buildPreferTypeMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "typeOverInterface",
		Description: "Use a 'type' instead of an 'interface' for object type definitions.",
	}
}

// Check if a type node is an object type (type literal)
func isObjectType(typeNode *ast.Node) bool {
	if typeNode == nil {
		return false
	}

	return typeNode.Kind == ast.KindTypeLiteral
}

// Convert type alias to interface
func convertTypeToInterface(ctx rule.RuleContext, typeAlias *ast.TypeAliasDeclaration) string {
	if typeAlias == nil || typeAlias.Name == nil {
		return ""
	}

	// Get name
	nameRange := utils.TrimNodeTextRange(ctx.SourceFile, typeAlias.Name)
	name := ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()]

	// Get type parameters if any
	typeParams := ""
	if typeAlias.TypeParameters != nil && len(typeAlias.TypeParameters.Arr) > 0 {
		typeParamsRange := utils.TrimNodeTextRange(ctx.SourceFile, typeAlias.TypeParameters)
		typeParams = ctx.SourceFile.Text()[typeParamsRange.Pos():typeParamsRange.End()]
	}

	// Get type body
	typeBody := ""
	if typeAlias.Type != nil && typeAlias.Type.Kind == ast.KindTypeLiteral {
		typeLiteral := typeAlias.Type.AsTypeLiteralNode()
		if typeLiteral != nil {
			// Get the text of the type literal
			bodyRange := utils.TrimNodeTextRange(ctx.SourceFile, typeAlias.Type)
			typeBody = ctx.SourceFile.Text()[bodyRange.Pos():bodyRange.End()]
		}
	}

	result := "interface " + name + typeParams + " " + typeBody
	return result
}

// Convert interface to type alias
func convertInterfaceToType(ctx rule.RuleContext, interfaceDecl *ast.InterfaceDeclaration) string {
	if interfaceDecl == nil || interfaceDecl.Name == nil {
		return ""
	}

	// Get name
	nameRange := utils.TrimNodeTextRange(ctx.SourceFile, interfaceDecl.Name)
	name := ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()]

	// Get type parameters if any
	typeParams := ""
	if interfaceDecl.TypeParameters != nil && len(interfaceDecl.TypeParameters.Arr) > 0 {
		typeParamsRange := utils.TrimNodeTextRange(ctx.SourceFile, interfaceDecl.TypeParameters)
		typeParams = ctx.SourceFile.Text()[typeParamsRange.Pos():typeParamsRange.End()]
	}

	// Get members and convert to type literal
	membersText := ""
	if interfaceDecl.Members != nil && len(interfaceDecl.Members.Arr) > 0 {
		// Find the opening brace position
		openBrace := -1
		closeBrace := -1
		sourceText := ctx.SourceFile.Text()

		// Search for braces around the members
		for i := interfaceDecl.Name.End(); i < interfaceDecl.End(); i++ {
			if sourceText[i] == '{' && openBrace == -1 {
				openBrace = i
			}
			if sourceText[i] == '}' {
				closeBrace = i
			}
		}

		if openBrace != -1 && closeBrace != -1 {
			membersText = sourceText[openBrace : closeBrace+1]
		}
	} else {
		membersText = "{}"
	}

	result := "type " + name + typeParams + " = " + membersText
	return result
}

var ConsistentTypeDefinitionsRule = rule.CreateRule(rule.Rule{
	Name: "consistent-type-definitions",
	Run: func(ctx rule.RuleContext, options interface{}) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			// Check TypeAliasDeclaration (type X = {...})
			ast.KindTypeAliasDeclaration: func(node *ast.Node) {
				if node.Kind != ast.KindTypeAliasDeclaration {
					return
				}

				typeAlias := node.AsTypeAliasDeclaration()
				if typeAlias == nil || typeAlias.Type == nil {
					return
				}

				// Only check if it's an object type (type literal)
				if !isObjectType(typeAlias.Type) {
					return
				}

				// If preferring interface, report type alias
				if opts.Style == "interface" {
					replacement := convertTypeToInterface(ctx, typeAlias)
					ctx.ReportNodeWithFixes(
						node,
						buildPreferInterfaceMessage(),
						rule.RuleFixReplace(ctx.SourceFile, node, replacement),
					)
				}
			},

			// Check InterfaceDeclaration (interface X {...})
			ast.KindInterfaceDeclaration: func(node *ast.Node) {
				if node.Kind != ast.KindInterfaceDeclaration {
					return
				}

				interfaceDecl := node.AsInterfaceDeclaration()
				if interfaceDecl == nil {
					return
				}

				// If preferring type, report interface
				if opts.Style == "type" {
					replacement := convertInterfaceToType(ctx, interfaceDecl)
					ctx.ReportNodeWithFixes(
						node,
						buildPreferTypeMessage(),
						rule.RuleFixReplace(ctx.SourceFile, node, replacement),
					)
				}
			},
		}
	},
})
