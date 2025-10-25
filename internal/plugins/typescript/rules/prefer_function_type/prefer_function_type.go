package prefer_function_type

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// PreferFunctionTypeRule implements the prefer-function-type rule
// Enforce function types instead of interfaces with call signatures
var PreferFunctionTypeRule = rule.CreateRule(rule.Rule{
	Name: "prefer-function-type",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	checkInterfaceOrTypeLiteral := func(node *ast.Node, members *ast.NodeArray, isInterface bool) {
		if members == nil || len(members.Nodes) != 1 {
			return
		}

		// Check if the single member is a call signature
		member := members.Nodes[0]
		if member.Kind != ast.KindCallSignature {
			return
		}

		callSig := member.AsCallSignature()
		if callSig == nil {
			return
		}

		// Check for 'this' type in return type or parameters (cannot be converted)
		if hasThisType(callSig) {
			ctx.ReportNode(member, rule.RuleMessage{
				Id:          "unexpectedThisOnFunctionOnlyInterface",
				Description: "Interface only has a call signature with a 'this' parameter, which can't be converted to a function type.",
			})
			return
		}

		// For interfaces, check if it extends other types (if so, we can't convert)
		if isInterface {
			interfaceDecl := node.AsInterfaceDeclaration()
			if interfaceDecl != nil {
				// Check if interface extends anything
				if interfaceDecl.HeritageClauses != nil && len(interfaceDecl.HeritageClauses.Nodes) > 0 {
					for _, clause := range interfaceDecl.HeritageClauses.Nodes {
						heritageClause := clause.AsHeritageClause()
						if heritageClause == nil {
							continue
						}
						if heritageClause.Token == ast.KindExtendsKeyword {
							// Check if it extends Function specifically (this is ok to convert)
							if len(heritageClause.Types.Nodes) == 1 {
								extendedType := heritageClause.Types.Nodes[0]
								typeRange := utils.TrimNodeTextRange(ctx.SourceFile, extendedType)
								typeText := ctx.SourceFile.Text()[typeRange.Pos():typeRange.End()]
								if strings.TrimSpace(typeText) == "Function" {
									// Extends Function only - OK to convert
									break
								}
							}
							// Extends something else - cannot convert
							return
						}
					}
				}
			}
		}

		// Report the issue with autofix
		ctx.ReportNodeWithFixes(member, rule.RuleMessage{
			Id:          "functionTypeOverCallableType",
			Description: "Interface or type literal has only a call signature — use a function type instead.",
		}, buildFix(ctx, node, callSig, isInterface))
	}

	return rule.RuleListeners{
		ast.KindInterfaceDeclaration: func(node *ast.Node) {
			interfaceDecl := node.AsInterfaceDeclaration()
			if interfaceDecl == nil {
				return
			}
			checkInterfaceOrTypeLiteral(node, interfaceDecl.Members, true)
		},
		ast.KindTypeLiteral: func(node *ast.Node) {
			typeLit := node.AsTypeLiteralNode()
			if typeLit == nil {
				return
			}
			checkInterfaceOrTypeLiteral(node, typeLit.Members, false)
		},
	}
}

// hasThisType checks if a call signature uses 'this' in parameters or return type
func hasThisType(callSig *ast.CallSignature) bool {
	// Check return type for 'this'
	if callSig.Type != nil {
		if hasThisInType(callSig.Type) {
			return true
		}
	}

	// Check parameters for 'this' parameter
	if callSig.Parameters != nil {
		for _, param := range callSig.Parameters.Nodes {
			paramDecl := param.AsParameterDeclaration()
			if paramDecl == nil {
				continue
			}
			// Check if parameter name is 'this'
			if paramDecl.Name != nil && paramDecl.Name.Kind == ast.KindIdentifier {
				id := paramDecl.Name.AsIdentifier()
				if id != nil && id.Text() == "this" {
					return true
				}
			}
		}
	}

	return false
}

// hasThisInType checks if a type node contains a 'this' type
func hasThisInType(typeNode *ast.Node) bool {
	if typeNode == nil {
		return false
	}
	if typeNode.Kind == ast.KindThisType {
		return true
	}
	// Check union/intersection types
	if typeNode.Kind == ast.KindUnionType {
		union := typeNode.AsUnionTypeNode()
		if union != nil && union.Types != nil {
			for _, t := range union.Types.Nodes {
				if hasThisInType(t) {
					return true
				}
			}
		}
	}
	return false
}

// buildFix creates a fix to convert interface/type literal to function type
func buildFix(ctx rule.RuleContext, node *ast.Node, callSig *ast.CallSignature, isInterface bool) *rule.RuleFix {
	// Extract the function signature
	sigRange := utils.TrimNodeTextRange(ctx.SourceFile, callSig)
	sigText := ctx.SourceFile.Text()[sigRange.Pos():sigRange.End()]

	// Remove leading/trailing whitespace and semicolon
	sigText = strings.TrimSpace(sigText)
	sigText = strings.TrimSuffix(sigText, ";")
	sigText = strings.TrimSpace(sigText)

	// Convert call signature to function type
	// Call signature format: (params): ReturnType
	// Function type format: (params) => ReturnType

	// Find the position of the closing paren and colon
	functionType := convertCallSignatureToFunctionType(sigText)

	if isInterface {
		interfaceDecl := node.AsInterfaceDeclaration()
		if interfaceDecl == nil {
			return nil
		}

		// Build the type alias replacement
		var exportText string
		if interfaceDecl.Modifiers() != nil {
			for _, modifier := range interfaceDecl.Modifiers().Nodes {
				if modifier.Kind == ast.KindExportKeyword {
					exportText = "export "
					break
				}
			}
		}

		// Extract interface name
		nameRange := utils.TrimNodeTextRange(ctx.SourceFile, interfaceDecl.Name())
		nameText := ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()]

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

		// Check for JSDoc comments before the interface
		var commentText string
		nodeRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
		sourceText := ctx.SourceFile.Text()

		// Look for leading comments
		pos := nodeRange.Pos()
		// Skip backwards to find comments
		searchStart := 0
		if pos > 200 {
			searchStart = pos - 200
		}

		beforeText := sourceText[searchStart:pos]

		// Find JSDoc comment (/** ... */)
		if idx := strings.LastIndex(beforeText, "/**"); idx >= 0 {
			endIdx := strings.Index(beforeText[idx:], "*/")
			if endIdx >= 0 {
				comment := strings.TrimSpace(beforeText[idx : idx+endIdx+2])
				// Check if there's only whitespace between comment and interface
				afterComment := beforeText[idx+endIdx+2:]
				if strings.TrimSpace(afterComment) == "" || strings.TrimSpace(afterComment) == exportText {
					commentText = comment + " "
				}
			}
		}

		replacement := fmt.Sprintf("%s%stype %s%s = %s", commentText, exportText, nameText, typeParamsText, functionType)
		return rule.RuleFixReplace(ctx.SourceFile, node, replacement)
	} else {
		// For type literals, just replace the type literal with function type
		return rule.RuleFixReplace(ctx.SourceFile, node, functionType)
	}
}

// convertCallSignatureToFunctionType converts a call signature to function type syntax
func convertCallSignatureToFunctionType(callSig string) string {
	// Find the position of ): which separates parameters from return type
	// Need to handle nested parentheses
	depth := 0
	colonPos := -1

	for i, ch := range callSig {
		if ch == '(' {
			depth++
		} else if ch == ')' {
			depth--
			if depth == 0 {
				// Found the closing paren of parameters
				// Look for : after it
				remaining := callSig[i+1:]
				remaining = strings.TrimSpace(remaining)
				if strings.HasPrefix(remaining, ":") {
					colonPos = i + 1 + (len(callSig[i+1:]) - len(remaining))
					break
				}
			}
		}
	}

	if colonPos == -1 {
		// No return type, just add => void
		return callSig + " => void"
	}

	// Split at the colon
	params := strings.TrimSpace(callSig[:colonPos])
	returnType := strings.TrimSpace(callSig[colonPos+1:])

	return params + " => " + returnType
}
