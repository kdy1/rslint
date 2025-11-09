package consistent_type_definitions

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type DefinitionStyle string

const (
	DefinitionStyleInterface DefinitionStyle = "interface"
	DefinitionStyleType      DefinitionStyle = "type"
)

type ConsistentTypeDefinitionsOptions struct {
	Style DefinitionStyle `json:"style"`
}

// Helper to check if interface is in a "declare global" block
func isInDeclareGlobal(node *ast.Node) bool {
	current := node.Parent
	for current != nil {
		if current.Kind == ast.KindModuleDeclaration {
			moduleDecl := current.AsModuleDeclaration()
			if moduleDecl != nil && moduleDecl.Name() != nil {
				// Check if module has "declare" modifier and name is "global"
				if ast.IsIdentifier(moduleDecl.Name()) {
					ident := moduleDecl.Name().AsIdentifier()
					if ident != nil && ident.Text == "global" {
						// Check if the module declaration has "declare" modifier
						parent := current.Parent
						if parent != nil && parent.Kind == ast.KindModuleBlock {
							// Look for declare keyword
							if current.Modifiers != nil {
								for _, mod := range current.Modifiers.Nodes {
									if mod.Kind == ast.KindDeclareKeyword {
										return true
									}
								}
							}
						}
					}
				}
			}
		}
		current = current.Parent
	}
	return false
}

// Helper to check if node is a default export
func isDefaultExport(node *ast.Node) bool {
	if node.Modifiers == nil {
		return false
	}
	for _, mod := range node.Modifiers.Nodes {
		if mod.Kind == ast.KindDefaultKeyword {
			return true
		}
	}
	return false
}

// Helper to unwrap parenthesized types
func unwrapParenthesizedType(typeNode *ast.Node) *ast.Node {
	current := typeNode
	for current != nil && current.Kind == ast.KindParenthesizedType {
		parenthesized := current.AsParenthesizedTypeNode()
		if parenthesized == nil || parenthesized.Type == nil {
			break
		}
		current = parenthesized.Type
	}
	return current
}

// Convert type alias to interface
func convertTypeToInterface(ctx rule.RuleContext, node *ast.Node, typeAlias *ast.TypeAliasDeclaration) rule.RuleFix {
	sourceText := ctx.SourceFile.Text()

	// Get the name and type parameters
	nameNode := typeAlias.Name()
	var typeParamsText string
	if typeAlias.TypeParameters != nil {
		typeParamsRange := utils.TrimNodeTextRange(ctx.SourceFile, typeAlias.TypeParameters)
		typeParamsText = sourceText[typeParamsRange.Pos():typeParamsRange.End()]
	}

	// Unwrap parentheses from the type
	unwrappedType := unwrapParenthesizedType(typeAlias.Type)

	// Get the body text (the object type literal)
	bodyRange := utils.TrimNodeTextRange(ctx.SourceFile, unwrappedType)
	bodyText := sourceText[bodyRange.Pos():bodyRange.End()]

	// Get the export/declare modifiers
	var modifiers []string
	if node.Modifiers != nil {
		for _, mod := range node.Modifiers.Nodes {
			modRange := utils.TrimNodeTextRange(ctx.SourceFile, mod)
			modText := sourceText[modRange.Pos():modRange.End()]
			modifiers = append(modifiers, modText)
		}
	}

	// Build the interface declaration
	var parts []string
	if len(modifiers) > 0 {
		parts = append(parts, strings.Join(modifiers, " "))
	}
	parts = append(parts, "interface")

	nameRange := utils.TrimNodeTextRange(ctx.SourceFile, nameNode)
	namePath := sourceText[nameRange.Pos():nameRange.End()]
	parts = append(parts, namePath)

	if typeParamsText != "" {
		parts = append(parts, typeParamsText)
	}

	// Join and add body - remove the trailing semicolon if present
	result := strings.Join(parts, " ") + " " + bodyText

	return rule.RuleFixReplace(ctx.SourceFile, node, result)
}

// Convert interface to type alias
func convertInterfaceToType(ctx rule.RuleContext, node *ast.Node, interfaceDecl *ast.InterfaceDeclaration) rule.RuleFix {
	sourceText := ctx.SourceFile.Text()

	// Handle default export specially
	if isDefaultExport(node) {
		// For "export default interface Foo", convert to:
		// type Foo = { ... }
		// export default Foo
		nameNode := interfaceDecl.Name()
		nameRange := utils.TrimNodeTextRange(ctx.SourceFile, nameNode)
		nameText := sourceText[nameRange.Pos():nameRange.End()]

		var typeParamsText string
		if interfaceDecl.TypeParameters != nil {
			typeParamsRange := utils.TrimNodeTextRange(ctx.SourceFile, interfaceDecl.TypeParameters)
			typeParamsText = sourceText[typeParamsRange.Pos():typeParamsRange.End()]
		}

		// Get the body
		bodyRange := utils.TrimNodeTextRange(ctx.SourceFile, interfaceDecl.Members)
		bodyText := sourceText[bodyRange.Pos():bodyRange.End()]

		result := fmt.Sprintf("type %s%s = %s\nexport default %s", nameText, typeParamsText, bodyText, nameText)
		return rule.RuleFixReplace(ctx.SourceFile, node, result)
	}

	// Get the name and type parameters
	nameNode := interfaceDecl.Name()
	var typeParamsText string
	if interfaceDecl.TypeParameters != nil {
		typeParamsRange := utils.TrimNodeTextRange(ctx.SourceFile, interfaceDecl.TypeParameters)
		typeParamsText = sourceText[typeParamsRange.Pos():typeParamsRange.End()]
	}

	// Get the body text
	bodyRange := utils.TrimNodeTextRange(ctx.SourceFile, interfaceDecl.Members)
	bodyText := sourceText[bodyRange.Pos():bodyRange.End()]

	// Get the export/declare modifiers (exclude "default")
	var modifiers []string
	if node.Modifiers != nil {
		for _, mod := range node.Modifiers.Nodes {
			if mod.Kind != ast.KindDefaultKeyword {
				modRange := utils.TrimNodeTextRange(ctx.SourceFile, mod)
				modText := sourceText[modRange.Pos():modRange.End()]
				modifiers = append(modifiers, modText)
			}
		}
	}

	// Build the type alias
	var parts []string
	if len(modifiers) > 0 {
		parts = append(parts, strings.Join(modifiers, " "))
	}
	parts = append(parts, "type")

	nameRange := utils.TrimNodeTextRange(ctx.SourceFile, nameNode)
	namePath := sourceText[nameRange.Pos():nameRange.End()]
	parts = append(parts, namePath)

	if typeParamsText != "" {
		parts = append(parts, typeParamsText)
	}

	parts = append(parts, "=")

	// Handle extends clause by converting to intersection type
	var extendsTypes []string
	if interfaceDecl.HeritageClauses != nil {
		for _, heritage := range interfaceDecl.HeritageClauses.Nodes {
			heritageClause := heritage.AsHeritageClause()
			if heritageClause != nil && heritageClause.Token == ast.KindExtendsKeyword {
				if heritageClause.Types != nil {
					for _, exprWithTypeArgs := range heritageClause.Types.Nodes {
						exprRange := utils.TrimNodeTextRange(ctx.SourceFile, exprWithTypeArgs)
						extendsTypes = append(extendsTypes, sourceText[exprRange.Pos():exprRange.End()])
					}
				}
			}
		}
	}

	// Build the result
	result := strings.Join(parts, " ")
	result += " " + bodyText

	// Add extends types as intersection
	if len(extendsTypes) > 0 {
		result += " & " + strings.Join(extendsTypes, " & ")
	}

	return rule.RuleFixReplace(ctx.SourceFile, node, result)
}

// ConsistentTypeDefinitionsRule enforces consistent type definitions
var ConsistentTypeDefinitionsRule = rule.CreateRule(rule.Rule{
	Name: "consistent-type-definitions",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := ConsistentTypeDefinitionsOptions{
		Style: DefinitionStyleInterface,
	}

	// Parse options
	if options != nil {
		if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
			if str, ok := optArray[0].(string); ok {
				opts.Style = DefinitionStyle(str)
			}
		} else if str, ok := options.(string); ok {
			opts.Style = DefinitionStyle(str)
		}
	}

	// Helper to check if a type is an object type literal (without index signatures or mapped types)
	isObjectTypeLiteral := func(typeNode *ast.Node) bool {
		if typeNode == nil {
			return false
		}
		if typeNode.Kind != ast.KindTypeLiteral {
			return false
		}

		// Check if type literal contains index signatures or mapped types
		typeLiteral := typeNode.AsTypeLiteralNode()
		if typeLiteral == nil || typeLiteral.Members == nil {
			return true
		}

		// If any member is an index signature, this is not a simple object type
		for _, member := range typeLiteral.Members.Nodes {
			if member.Kind == ast.KindIndexSignature {
				return false
			}
		}

		return true
	}

	// Helper to check if a type alias is a simple object type (not a union, intersection, etc.)
	isSimpleObjectType := func(typeNode *ast.Node) bool {
		if typeNode == nil {
			return false
		}

		// Check if it's a parenthesized type wrapping an object type
		if typeNode.Kind == ast.KindParenthesizedType {
			parenthesized := typeNode.AsParenthesizedTypeNode()
			if parenthesized != nil {
				return isObjectTypeLiteral(parenthesized.Type)
			}
		}

		return isObjectTypeLiteral(typeNode)
	}

	// Helper to check if interface is in a globally-scoped module
	isInGlobalModule := func(node *ast.Node) bool {
		current := node.Parent
		for current != nil {
			if current.Kind == ast.KindModuleDeclaration {
				moduleDecl := current.AsModuleDeclaration()
				if moduleDecl != nil && moduleDecl.Name() != nil {
					// Check if module name is 'global'
					if ast.IsIdentifier(moduleDecl.Name()) {
						ident := moduleDecl.Name().AsIdentifier()
						if ident != nil && ident.Text == "global" {
							return true
						}
					}
				}
			}
			current = current.Parent
		}
		return false
	}

	checkTypeAlias := func(node *ast.Node) {
		if opts.Style != DefinitionStyleInterface {
			return
		}

		typeAlias := node.AsTypeAliasDeclaration()
		if typeAlias == nil {
			return
		}

		// Only report if it's a simple object type literal
		if !isSimpleObjectType(typeAlias.Type) {
			return
		}

		// Generate auto-fix: convert type to interface
		fix := convertTypeToInterface(ctx, node, typeAlias)

		ctx.ReportNodeWithFixes(node, rule.RuleMessage{
			Id:          "interfaceOverType",
			Description: "Use an interface instead of a type literal.",
		}, fix)
	}

	checkInterface := func(node *ast.Node) {
		if opts.Style != DefinitionStyleType {
			return
		}

		interfaceDecl := node.AsInterfaceDeclaration()
		if interfaceDecl == nil {
			return
		}

		// Check if we can provide an auto-fix
		canFix := !isInGlobalModule(node) && !isInDeclareGlobal(node) && !isDefaultExport(node)

		if canFix {
			// Generate auto-fix: convert interface to type
			fix := convertInterfaceToType(ctx, node, interfaceDecl)

			ctx.ReportNodeWithFixes(node, rule.RuleMessage{
				Id:          "typeOverInterface",
				Description: "Use a type literal instead of an interface.",
			}, fix)
		} else {
			// Report without fix for special cases
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "typeOverInterface",
				Description: "Use a type literal instead of an interface.",
			})
		}
	}

	return rule.RuleListeners{
		ast.KindTypeAliasDeclaration: checkTypeAlias,
		ast.KindInterfaceDeclaration: checkInterface,
	}
}
