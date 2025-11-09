package method_signature_style

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type SignatureStyle string

const (
	SignatureStyleProperty SignatureStyle = "property"
	SignatureStyleMethod   SignatureStyle = "method"
)

type MethodSignatureStyleOptions struct {
	Style SignatureStyle `json:"style"`
}

var MethodSignatureStyleRule = rule.CreateRule(rule.Rule{
	Name: "method-signature-style",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := MethodSignatureStyleOptions{
		Style: SignatureStyleProperty,
	}

	// Parse options
	if options != nil {
		if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
			if str, ok := optArray[0].(string); ok {
				opts.Style = SignatureStyle(str)
			}
		} else if str, ok := options.(string); ok {
			opts.Style = SignatureStyle(str)
		}
	}

	// Helper to check if a member is in a globally-scoped module
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

	// Helper to get the trailing delimiter (semicolon, comma, or nothing)
	getTrailingDelimiter := func(node *ast.Node) string {
		nodeRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
		text := ctx.SourceFile.Text()

		// Look ahead after the node to find delimiter
		pos := nodeRange.End()
		for pos < len(text) {
			ch := text[pos]
			if ch == ';' {
				return ";"
			} else if ch == ',' {
				return ","
			} else if ch == '\n' || ch == '\r' {
				// Newline without delimiter means no delimiter
				return ""
			} else if ch != ' ' && ch != '\t' {
				// Non-whitespace character that's not a delimiter
				return ""
			}
			pos++
		}
		return ""
	}

	// Convert method signature to property signature
	convertMethodToProperty := func(node *ast.Node) string {
		methodSig := node.AsMethodSignature()
		if methodSig == nil {
			return ""
		}

		// Get the name text (could be identifier, string literal, or computed property)
		var nameText string
		nameRange := utils.TrimNodeTextRange(ctx.SourceFile, methodSig.Name())
		nameText = ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()]

		// Handle optional modifier
		optionalToken := ""
		if methodSig.QuestionToken != nil {
			optionalToken = "?"
		}

		// Get type parameters if present
		var typeParamsText string
		if methodSig.TypeParameters != nil && len(methodSig.TypeParameters.Nodes) > 0 {
			firstParam := methodSig.TypeParameters.Nodes[0]
			lastParam := methodSig.TypeParameters.Nodes[len(methodSig.TypeParameters.Nodes)-1]
			firstRange := utils.TrimNodeTextRange(ctx.SourceFile, firstParam)
			lastRange := utils.TrimNodeTextRange(ctx.SourceFile, lastParam)

			// Find the opening < before the first type parameter
			openPos := firstRange.Pos() - 1
			for openPos > 0 && ctx.SourceFile.Text()[openPos] != '<' {
				openPos--
			}

			// Find the closing > after the last type parameter
			closePos := lastRange.End()
			for closePos < len(ctx.SourceFile.Text()) && ctx.SourceFile.Text()[closePos] != '>' {
				closePos++
			}
			if closePos < len(ctx.SourceFile.Text()) {
				closePos++ // Include the >
			}

			typeParamsText = ctx.SourceFile.Text()[openPos:closePos]
		}

		// Get parameters text
		var paramsText string
		if methodSig.Parameters != nil && len(methodSig.Parameters.Nodes) > 0 {
			firstParam := methodSig.Parameters.Nodes[0]
			lastParam := methodSig.Parameters.Nodes[len(methodSig.Parameters.Nodes)-1]
			firstRange := utils.TrimNodeTextRange(ctx.SourceFile, firstParam)
			lastRange := utils.TrimNodeTextRange(ctx.SourceFile, lastParam)

			// Find the opening ( before the first parameter
			openPos := firstRange.Pos() - 1
			for openPos > 0 && ctx.SourceFile.Text()[openPos] != '(' {
				openPos--
			}

			// Find the closing ) after the last parameter
			closePos := lastRange.End()
			for closePos < len(ctx.SourceFile.Text()) && ctx.SourceFile.Text()[closePos] != ')' {
				closePos++
			}
			if closePos < len(ctx.SourceFile.Text()) {
				closePos++ // Include the )
			}

			paramsText = ctx.SourceFile.Text()[openPos:closePos]
		} else {
			paramsText = "()"
		}

		// Get return type text
		var returnTypeText string
		if methodSig.Type != nil {
			typeRange := utils.TrimNodeTextRange(ctx.SourceFile, methodSig.Type)
			returnTypeText = ctx.SourceFile.Text()[typeRange.Pos():typeRange.End()]
		} else {
			// If no return type is specified, it's implicitly 'any'
			returnTypeText = "any"
		}

		// Build the property signature
		return fmt.Sprintf("%s%s: %s%s => %s", nameText, optionalToken, typeParamsText, paramsText, returnTypeText)
	}

	// Convert property signature to method signature
	convertPropertyToMethod := func(node *ast.Node) string {
		propertySig := node.AsPropertySignature()
		if propertySig == nil {
			return ""
		}

		// Get the name text
		var nameText string
		nameRange := utils.TrimNodeTextRange(ctx.SourceFile, propertySig.Name())
		nameText = ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()]

		// Handle optional modifier
		optionalToken := ""
		if propertySig.QuestionToken != nil {
			optionalToken = "?"
		}

		// The type must be a function type
		if propertySig.Type == nil {
			return ""
		}

		funcType := propertySig.Type.AsFunctionTypeNode()
		if funcType == nil {
			return ""
		}

		// Get type parameters if present
		var typeParamsText string
		if funcType.TypeParameters != nil && len(funcType.TypeParameters.Nodes) > 0 {
			firstParam := funcType.TypeParameters.Nodes[0]
			lastParam := funcType.TypeParameters.Nodes[len(funcType.TypeParameters.Nodes)-1]
			firstRange := utils.TrimNodeTextRange(ctx.SourceFile, firstParam)
			lastRange := utils.TrimNodeTextRange(ctx.SourceFile, lastParam)

			// Find the opening < before the first type parameter
			openPos := firstRange.Pos() - 1
			for openPos > 0 && ctx.SourceFile.Text()[openPos] != '<' {
				openPos--
			}

			// Find the closing > after the last type parameter
			closePos := lastRange.End()
			for closePos < len(ctx.SourceFile.Text()) && ctx.SourceFile.Text()[closePos] != '>' {
				closePos++
			}
			if closePos < len(ctx.SourceFile.Text()) {
				closePos++ // Include the >
			}

			typeParamsText = ctx.SourceFile.Text()[openPos:closePos]
		}

		// Get parameters text
		var paramsText string
		if funcType.Parameters != nil && len(funcType.Parameters.Nodes) > 0 {
			firstParam := funcType.Parameters.Nodes[0]
			lastParam := funcType.Parameters.Nodes[len(funcType.Parameters.Nodes)-1]
			firstRange := utils.TrimNodeTextRange(ctx.SourceFile, firstParam)
			lastRange := utils.TrimNodeTextRange(ctx.SourceFile, lastParam)

			// Find the opening ( before the first parameter
			openPos := firstRange.Pos() - 1
			for openPos > 0 && ctx.SourceFile.Text()[openPos] != '(' {
				openPos--
			}

			// Find the closing ) after the last parameter
			closePos := lastRange.End()
			for closePos < len(ctx.SourceFile.Text()) && ctx.SourceFile.Text()[closePos] != ')' {
				closePos++
			}
			if closePos < len(ctx.SourceFile.Text()) {
				closePos++ // Include the )
			}

			paramsText = ctx.SourceFile.Text()[openPos:closePos]
		} else {
			paramsText = "()"
		}

		// Get return type text
		var returnTypeText string
		if funcType.Type != nil {
			typeRange := utils.TrimNodeTextRange(ctx.SourceFile, funcType.Type)
			returnTypeText = ctx.SourceFile.Text()[typeRange.Pos():typeRange.End()]
		} else {
			returnTypeText = "void"
		}

		// Build the method signature
		return fmt.Sprintf("%s%s%s%s: %s", nameText, optionalToken, typeParamsText, paramsText, returnTypeText)
	}

	// Group overloaded method signatures by name
	groupOverloadedMethods := func(members []*ast.Node) map[string][]*ast.Node {
		groups := make(map[string][]*ast.Node)

		for _, member := range members {
			if member.Kind != ast.KindMethodSignature {
				continue
			}

			methodSig := member.AsMethodSignature()
			if methodSig == nil {
				continue
			}

			// Get the method name key
			nameRange := utils.TrimNodeTextRange(ctx.SourceFile, methodSig.Name())
			nameText := ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()]

			groups[nameText] = append(groups[nameText], member)
		}

		return groups
	}

	// Convert overloaded method signatures to a single property signature with intersection types
	convertOverloadedMethodsToProperty := func(methods []*ast.Node) string {
		if len(methods) == 0 {
			return ""
		}

		// Get the name from the first method
		firstMethod := methods[0].AsMethodSignature()
		if firstMethod == nil {
			return ""
		}

		nameRange := utils.TrimNodeTextRange(ctx.SourceFile, firstMethod.Name())
		nameText := ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()]

		// Build intersection of function types
		var functionTypes []string
		for _, method := range methods {
			methodSig := method.AsMethodSignature()
			if methodSig == nil {
				continue
			}

			// Get type parameters if present
			var typeParamsText string
			if methodSig.TypeParameters != nil && len(methodSig.TypeParameters.Nodes) > 0 {
				firstParam := methodSig.TypeParameters.Nodes[0]
				lastParam := methodSig.TypeParameters.Nodes[len(methodSig.TypeParameters.Nodes)-1]
				firstRange := utils.TrimNodeTextRange(ctx.SourceFile, firstParam)
				lastRange := utils.TrimNodeTextRange(ctx.SourceFile, lastParam)

				openPos := firstRange.Pos() - 1
				for openPos > 0 && ctx.SourceFile.Text()[openPos] != '<' {
					openPos--
				}

				closePos := lastRange.End()
				for closePos < len(ctx.SourceFile.Text()) && ctx.SourceFile.Text()[closePos] != '>' {
					closePos++
				}
				if closePos < len(ctx.SourceFile.Text()) {
					closePos++
				}

				typeParamsText = ctx.SourceFile.Text()[openPos:closePos]
			}

			// Get parameters text
			var paramsText string
			if methodSig.Parameters != nil && len(methodSig.Parameters.Nodes) > 0 {
				firstParam := methodSig.Parameters.Nodes[0]
				lastParam := methodSig.Parameters.Nodes[len(methodSig.Parameters.Nodes)-1]
				firstRange := utils.TrimNodeTextRange(ctx.SourceFile, firstParam)
				lastRange := utils.TrimNodeTextRange(ctx.SourceFile, lastParam)

				openPos := firstRange.Pos() - 1
				for openPos > 0 && ctx.SourceFile.Text()[openPos] != '(' {
					openPos--
				}

				closePos := lastRange.End()
				for closePos < len(ctx.SourceFile.Text()) && ctx.SourceFile.Text()[closePos] != ')' {
					closePos++
				}
				if closePos < len(ctx.SourceFile.Text()) {
					closePos++
				}

				paramsText = ctx.SourceFile.Text()[openPos:closePos]
			} else {
				paramsText = "()"
			}

			// Get return type text
			var returnTypeText string
			if methodSig.Type != nil {
				typeRange := utils.TrimNodeTextRange(ctx.SourceFile, methodSig.Type)
				returnTypeText = ctx.SourceFile.Text()[typeRange.Pos():typeRange.End()]
			} else {
				returnTypeText = "any"
			}

			functionTypes = append(functionTypes, fmt.Sprintf("(%s%s => %s)", typeParamsText, paramsText, returnTypeText))
		}

		// Join with intersection operator
		intersectionType := strings.Join(functionTypes, " & ")
		return fmt.Sprintf("%s: %s", nameText, intersectionType)
	}

	checkTypeLiteralOrInterface := func(node *ast.Node) {
		var members []*ast.Node

		if node.Kind == ast.KindTypeLiteral {
			typeLiteral := node.AsTypeLiteralNode()
			if typeLiteral == nil || typeLiteral.Members == nil {
				return
			}
			members = typeLiteral.Members.Nodes
		} else if node.Kind == ast.KindInterfaceDeclaration {
			interfaceDecl := node.AsInterfaceDeclaration()
			if interfaceDecl == nil || interfaceDecl.Members == nil {
				return
			}
			members = interfaceDecl.Members.Nodes
		} else {
			return
		}

		if opts.Style == SignatureStyleProperty {
			// Check for method signatures and convert them to property signatures
			overloadGroups := groupOverloadedMethods(members)

			for _, methodGroup := range overloadGroups {
				if len(methodGroup) > 1 {
					// Handle overloaded methods - convert to intersection type
					// Check if we're in a global module (can't auto-fix in this case)
					inGlobalModule := isInGlobalModule(methodGroup[0])

					if inGlobalModule {
						// Report each overload without fix
						for _, method := range methodGroup {
							ctx.ReportNode(method, rule.RuleMessage{
								Id:          "errorMethod",
								Description: "Signature should be a property.",
							})
						}
					} else {
						// Report each overload with fix
						for i, method := range methodGroup {
							message := rule.RuleMessage{
								Id:          "errorMethod",
								Description: "Signature should be a property.",
							}

							if i == 0 {
								// First overload: replace with the intersection type
								replacement := convertOverloadedMethodsToProperty(methodGroup)
								ctx.ReportNodeWithFixes(method, message,
									rule.RuleFixReplace(ctx.SourceFile, method, replacement))
							} else {
								// Other overloads: remove them
								ctx.ReportNodeWithFixes(method, message,
									rule.RuleFixRemove(ctx.SourceFile, method))
							}
						}
					}
				} else if len(methodGroup) == 1 {
					// Single method signature
					method := methodGroup[0]
					methodSig := method.AsMethodSignature()
					if methodSig == nil {
						continue
					}

					// Skip getters and setters
					if methodSig.Kind == ast.KindGetAccessor || methodSig.Kind == ast.KindSetAccessor {
						continue
					}

					replacement := convertMethodToProperty(method)
					delimiter := getTrailingDelimiter(method)
					if delimiter != "" {
						replacement += delimiter
					}

					ctx.ReportNodeWithFixes(method, rule.RuleMessage{
						Id:          "errorMethod",
						Description: "Signature should be a property.",
					}, rule.RuleFixReplace(ctx.SourceFile, method, replacement))
				}
			}

			// Also check for non-grouped method signatures (shouldn't happen, but just in case)
			for _, member := range members {
				if member.Kind == ast.KindMethodSignature {
					methodSig := member.AsMethodSignature()
					if methodSig == nil {
						continue
					}

					// Get the method name
					nameRange := utils.TrimNodeTextRange(ctx.SourceFile, methodSig.Name())
					nameText := ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()]

					// Skip if already handled in overload groups
					if group, exists := overloadGroups[nameText]; exists && len(group) > 0 {
						continue
					}

					replacement := convertMethodToProperty(member)
					delimiter := getTrailingDelimiter(member)
					if delimiter != "" {
						replacement += delimiter
					}

					ctx.ReportNodeWithFixes(member, rule.RuleMessage{
						Id:          "errorMethod",
						Description: "Signature should be a property.",
					}, rule.RuleFixReplace(ctx.SourceFile, member, replacement))
				}
			}
		} else {
			// Style is "method" - check for property signatures with function types
			for _, member := range members {
				if member.Kind != ast.KindPropertySignature {
					continue
				}

				propertySig := member.AsPropertySignature()
				if propertySig == nil || propertySig.Type == nil {
					continue
				}

				// Check if the type is a function type
				if propertySig.Type.Kind != ast.KindFunctionType {
					continue
				}

				replacement := convertPropertyToMethod(member)
				delimiter := getTrailingDelimiter(member)
				if delimiter != "" {
					replacement += delimiter
				}

				ctx.ReportNodeWithFixes(member, rule.RuleMessage{
					Id:          "errorProperty",
					Description: "Signature should be a method.",
				}, rule.RuleFixReplace(ctx.SourceFile, member, replacement))
			}
		}
	}

	return rule.RuleListeners{
		ast.KindTypeLiteral:          checkTypeLiteralOrInterface,
		ast.KindInterfaceDeclaration: checkTypeLiteralOrInterface,
	}
}
