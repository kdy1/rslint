package prefer_function_type

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildFunctionTypeOverCallableTypeMessage(literalOrInterface string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "functionTypeOverCallableType",
		Description: fmt.Sprintf("%s has only a call signature, you should use a function type instead.", literalOrInterface),
	}
}

func buildUnexpectedThisMessage(interfaceName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedThisOnFunctionOnlyInterface",
		Description: fmt.Sprintf("Interface %s has only a call signature with 'this' parameter or return type, you should convert it to a function type.", interfaceName),
	}
}

// Check if a type node contains 'this' keyword in parameters or return type (but not nested in objects)
func hasThisInSignature(signature *ast.Node) bool {
	if signature == nil || !ast.IsCallSignatureDeclaration(signature) {
		return false
	}

	callSig := signature.AsCallSignatureDeclaration()

	// Check parameters for 'this' type (not 'this' parameter which is allowed)
	if callSig.Parameters != nil {
		for _, param := range callSig.Parameters.Nodes {
			paramDecl := param.AsParameterDeclaration()
			if paramDecl != nil && paramDecl.Type != nil && containsThisTypeAtTopLevel(paramDecl.Type) {
				return true
			}
		}
	}

	// Check return type for 'this' type at top level (not nested)
	if callSig.Type != nil && containsThisTypeAtTopLevel(callSig.Type) {
		return true
	}

	return false
}

// Check if a type contains 'this' keyword at the top level (not nested in object types)
func containsThisTypeAtTopLevel(typeNode *ast.Node) bool {
	if typeNode == nil {
		return false
	}

	// Direct this type
	if ast.IsThisTypeNode(typeNode) {
		return true
	}

	// Union or intersection type - check members
	if typeNode.Kind == ast.KindUnionType {
		unionType := typeNode.AsUnionTypeNode()
		if unionType.Types != nil {
			for _, t := range unionType.Types.Nodes {
				if containsThisTypeAtTopLevel(t) {
					return true
				}
			}
		}
		return false
	}

	if typeNode.Kind == ast.KindIntersectionType {
		intersectionType := typeNode.AsIntersectionTypeNode()
		if intersectionType.Types != nil {
			for _, t := range intersectionType.Types.Nodes {
				if containsThisTypeAtTopLevel(t) {
					return true
				}
			}
		}
		return false
	}

	// Parenthesized type
	if ast.IsParenthesizedTypeNode(typeNode) {
		return containsThisTypeAtTopLevel(typeNode.AsParenthesizedTypeNode().Type)
	}

	// For object types, we don't consider nested 'this' as top-level
	// So we return false for TypeLiteral and other complex types
	return false
}

// Extract call signature's function type representation
func getCallSignatureText(ctx rule.RuleContext, signature *ast.Node) string {
	if signature == nil || !ast.IsCallSignatureDeclaration(signature) {
		return ""
	}

	callSig := signature.AsCallSignatureDeclaration()
	var parts []string

	// Type parameters
	if callSig.TypeParameters != nil && len(callSig.TypeParameters.Nodes) > 0 {
		firstParam := callSig.TypeParameters.Nodes[0]
		lastParam := callSig.TypeParameters.Nodes[len(callSig.TypeParameters.Nodes)-1]
		firstRange := utils.TrimNodeTextRange(ctx.SourceFile, firstParam)
		lastRange := utils.TrimNodeTextRange(ctx.SourceFile, lastParam)
		typeParamsRange := firstRange.WithEnd(lastRange.End())
		// Include the angle brackets
		typeParamsRange = typeParamsRange.WithPos(typeParamsRange.Pos() - 1).WithEnd(typeParamsRange.End() + 1)
		parts = append(parts, ctx.SourceFile.Text()[typeParamsRange.Pos():typeParamsRange.End()])
	}

	// Parameters
	paramsText := "("
	if callSig.Parameters != nil && len(callSig.Parameters.Nodes) > 0 {
		var paramTexts []string
		for _, param := range callSig.Parameters.Nodes {
			paramRange := utils.TrimNodeTextRange(ctx.SourceFile, param)
			paramTexts = append(paramTexts, ctx.SourceFile.Text()[paramRange.Pos():paramRange.End()])
		}
		paramsText += strings.Join(paramTexts, ", ")
	}
	paramsText += ")"
	parts = append(parts, paramsText)

	// Return type
	parts = append(parts, "=>")
	if callSig.Type != nil {
		returnTypeRange := utils.TrimNodeTextRange(ctx.SourceFile, callSig.Type)
		parts = append(parts, ctx.SourceFile.Text()[returnTypeRange.Pos():returnTypeRange.End()])
	} else {
		parts = append(parts, "void")
	}

	return strings.Join(parts, " ")
}

// Extract comments from the call signature member
func extractComments(ctx rule.RuleContext, member *ast.Node) string {
	if member == nil {
		return ""
	}

	// Get the text before the member to find comments
	memberRange := utils.TrimNodeTextRange(ctx.SourceFile, member)
	sourceText := ctx.SourceFile.Text()

	// Look backwards from the member start to find comments
	start := memberRange.Pos()
	
	// Find the previous non-whitespace content
	searchStart := start - 1
	for searchStart >= 0 && (sourceText[searchStart] == ' ' || sourceText[searchStart] == '\t' || sourceText[searchStart] == '\n' || sourceText[searchStart] == '\r') {
		searchStart--
	}

	if searchStart < 0 {
		return ""
	}

	// Check for block comment ending with */
	if searchStart >= 1 && sourceText[searchStart-1:searchStart+1] == "*/" {
		commentEnd := searchStart + 1
		commentStart := commentEnd - 2
		for commentStart >= 1 && sourceText[commentStart-1:commentStart+1] != "/*" {
			commentStart--
		}
		if commentStart >= 1 {
			commentStart-- // Include the '/'
			comment := sourceText[commentStart:commentEnd]
			return strings.TrimSpace(comment)
		}
	}

	// Check for line comment
	// Look for // at the beginning of the line
	lineStart := start
	for lineStart > 0 && sourceText[lineStart-1] != '\n' {
		lineStart--
	}
	
	// Skip whitespace at line start
	linePos := lineStart
	for linePos < start && (sourceText[linePos] == ' ' || sourceText[linePos] == '\t') {
		linePos++
	}
	
	if linePos+1 < start && sourceText[linePos:linePos+2] == "//" {
		lineEnd := start
		for lineEnd > linePos && (sourceText[lineEnd-1] == ' ' || sourceText[lineEnd-1] == '\t' || sourceText[lineEnd-1] == '\n' || sourceText[lineEnd-1] == '\r') {
			lineEnd--
		}
		comment := sourceText[linePos:lineEnd]
		return strings.TrimSpace(comment)
	}

	return ""
}

// Check if parent is a union or intersection type
func isInUnionOrIntersection(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}

	parent := node.Parent
	return parent.Kind == ast.KindUnionType || parent.Kind == ast.KindIntersectionType
}

var PreferFunctionTypeRule = rule.CreateRule(rule.Rule{
	Name: "prefer-function-type",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		
		checkInterfaceDeclaration := func(node *ast.Node) {
			interfaceDecl := node.AsInterfaceDeclaration()
			if interfaceDecl == nil {
				return
			}

			// Must have exactly one member
			if interfaceDecl.Members == nil || len(interfaceDecl.Members.Nodes) != 1 {
				return
			}

			member := interfaceDecl.Members.Nodes[0]
			if !ast.IsCallSignatureDeclaration(member) {
				return
			}

			// Check for additional extends clauses (more than just Function)
			hasNonFunctionExtends := false
			if interfaceDecl.HeritageClauses != nil {
				for _, clause := range interfaceDecl.HeritageClauses.Nodes {
					heritageClause := clause.AsHeritageClause()
					if heritageClause == nil || heritageClause.Token != ast.KindExtendsKeyword {
						continue
					}
					
					if heritageClause.Types != nil {
						for _, extendType := range heritageClause.Types.Nodes {
							// Check if it's extending something other than just Function
							if ast.IsExpressionWithTypeArguments(extendType) {
								exprWithType := extendType.AsExpressionWithTypeArguments()
								if ast.IsIdentifier(exprWithType.Expression) {
									ident := exprWithType.Expression.AsIdentifier()
									if ident.Text != "Function" {
										hasNonFunctionExtends = true
										break
									}
								} else {
									hasNonFunctionExtends = true
									break
								}
							}
						}
					}
				}
			}

			if hasNonFunctionExtends {
				return
			}

			// Check if signature has 'this' in parameters or return type
			if hasThisInSignature(member) {
				// Get interface name
				nameRange := utils.TrimNodeTextRange(ctx.SourceFile, interfaceDecl.Name())
				nameText := ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()]
				
				// Report without fix
				ctx.ReportNode(member, buildUnexpectedThisMessage(nameText))
				return
			}

			// Build the replacement
			var exportText string
			if interfaceDecl.Modifiers() != nil {
				for _, modifier := range interfaceDecl.Modifiers().Nodes {
					if modifier.Kind == ast.KindExportKeyword {
						exportText = "export "
						break
					}
				}
			}

			// Check for default export (cannot be converted easily with comments)
			isDefaultExport := false
			if interfaceDecl.Modifiers() != nil {
				for _, modifier := range interfaceDecl.Modifiers().Nodes {
					if modifier.Kind == ast.KindDefaultKeyword {
						isDefaultExport = true
						break
					}
				}
			}

			// Extract comments
			comment := extractComments(ctx, member)

			// If it's a default export with comments, we can't convert it properly
			if isDefaultExport && comment != "" {
				ctx.ReportNode(member, buildFunctionTypeOverCallableTypeMessage("Interface"))
				return
			}

			// Extract interface name
			nameRange := utils.TrimNodeTextRange(ctx.SourceFile, interfaceDecl.Name())
			nameText := ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()]

			// Extract type parameters
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

			// Get the function type text
			functionTypeText := getCallSignatureText(ctx, member)

			// Build replacement
			var replacement string
			if comment != "" {
				replacement = fmt.Sprintf("%s\n%stype %s%s = %s", comment, exportText, nameText, typeParamsText, functionTypeText)
			} else {
				replacement = fmt.Sprintf("%stype %s%s = %s", exportText, nameText, typeParamsText, functionTypeText)
			}

			ctx.ReportNodeWithFixes(member, buildFunctionTypeOverCallableTypeMessage("Interface"),
				rule.RuleFixReplace(ctx.SourceFile, node, replacement))
		}

		checkTypeLiteral := func(node *ast.Node) {
			typeLiteral := node.AsTypeLiteralNode()
			if typeLiteral == nil {
				return
			}

			// Must have exactly one member
			if typeLiteral.Members == nil || len(typeLiteral.Members.Nodes) != 1 {
				return
			}

			member := typeLiteral.Members.Nodes[0]
			if !ast.IsCallSignatureDeclaration(member) {
				return
			}

			// Get the function type text
			functionTypeText := getCallSignatureText(ctx, member)

			// Check for comments inside the type literal
			comment := extractComments(ctx, member)

			// Check if we need to wrap in parentheses (if in union or intersection)
			needsParens := isInUnionOrIntersection(node)

			var replacement string
			if comment != "" {
				if needsParens {
					replacement = fmt.Sprintf("%s (%s)", comment, functionTypeText)
				} else {
					replacement = fmt.Sprintf("%s %s", comment, functionTypeText)
				}
			} else {
				if needsParens {
					replacement = fmt.Sprintf("(%s)", functionTypeText)
				} else {
					replacement = functionTypeText
				}
			}

			ctx.ReportNodeWithFixes(member, buildFunctionTypeOverCallableTypeMessage("Type literal"),
				rule.RuleFixReplace(ctx.SourceFile, node, replacement))
		}

		return rule.RuleListeners{
			ast.KindInterfaceDeclaration: checkInterfaceDeclaration,
			ast.KindTypeLiteral:           checkTypeLiteral,
		}
	},
})
