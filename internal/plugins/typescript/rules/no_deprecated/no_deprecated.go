package no_deprecated

import (
	"regexp"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type NoDeprecatedOptions struct {
	Allow []interface{} `json:"allow"`
}

var NoDeprecatedRule = rule.CreateRule(rule.Rule{
	Name: "no-deprecated",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := NoDeprecatedOptions{
			Allow: nil,
		}

		// Parse options with dual-format support
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
				if allow, ok := optsMap["allow"].([]interface{}); ok {
					opts.Allow = allow
				}
			}
		}

		// Helper to check if a symbol is in the allow list
		isAllowed := func(symbolName string) bool {
			if opts.Allow == nil {
				return false
			}
			for _, item := range opts.Allow {
				// Support both string format and object format
				if str, ok := item.(string); ok && str == symbolName {
					return true
				}
				if obj, ok := item.(map[string]interface{}); ok {
					if name, ok := obj["name"].(string); ok && name == symbolName {
						return true
					}
				}
			}
			return false
		}

		// Helper to extract deprecation info from JSDoc comments
		getDeprecationReason := func(node *ast.Node) (bool, string) {
			if node == nil {
				return false, ""
			}

			// Get JSDoc comments from the node
			sourceText := ctx.SourceFile.Text()
			nodeStart := node.Pos()

			// Look backwards for JSDoc comments
			if nodeStart <= 0 {
				return false, ""
			}

			// Search backwards for /** ... */ comment before the node
			searchStart := nodeStart - 1
			if searchStart >= len(sourceText) {
				return false, ""
			}

			// Find the start of the line to check for JSDoc
			lineStart := searchStart
			for lineStart > 0 && sourceText[lineStart-1] != '\n' {
				lineStart--
			}

			// Look for JSDoc comment before this line
			commentEnd := lineStart
			for commentEnd > 0 && (sourceText[commentEnd-1] == ' ' || sourceText[commentEnd-1] == '\t' || sourceText[commentEnd-1] == '\n' || sourceText[commentEnd-1] == '\r') {
				commentEnd--
			}

			if commentEnd < 2 {
				return false, ""
			}

			// Check if there's a */ before the node
			if commentEnd >= 2 && sourceText[commentEnd-2:commentEnd] == "*/" {
				// Find the start of the comment
				commentStart := commentEnd - 2
				for commentStart > 1 && !(sourceText[commentStart-2:commentStart] == "/*") {
					commentStart--
				}

				if commentStart > 1 && sourceText[commentStart-2:commentStart] == "/*" {
					commentText := sourceText[commentStart-2 : commentEnd]

					// Check if it's a JSDoc comment (starts with /**)
					if len(commentText) >= 3 && commentText[:3] == "/**" {
						// Look for @deprecated tag
						deprecatedRegex := regexp.MustCompile(`@deprecated\s*([^\n@]*)`)
						matches := deprecatedRegex.FindStringSubmatch(commentText)
						if matches != nil {
							reason := strings.TrimSpace(matches[1])
							return true, reason
						}
					}
				}
			}

			return false, ""
		}

		// Helper to check if a symbol's declaration is deprecated
		isSymbolDeprecated := func(symbol *ast.Symbol) (bool, string) {
			if symbol == nil || symbol.Declarations == nil {
				return false, ""
			}

			// Check all declarations for deprecation
			for _, decl := range symbol.Declarations {
				if deprecated, reason := getDeprecationReason(decl); deprecated {
					return true, reason
				}
			}

			return false, ""
		}

		// Helper to get symbol name
		getSymbolName := func(symbol *ast.Symbol) string {
			if symbol == nil {
				return ""
			}
			// Try to get the escaped name
			if symbol.EscapedName != "" {
				return symbol.EscapedName
			}
			// Fallback to reading from declarations
			if symbol.Declarations != nil && len(symbol.Declarations) > 0 {
				decl := symbol.Declarations[0]
				if decl.Kind == ast.KindVariableDeclaration {
					varDecl := decl.AsVariableDeclaration()
					if varDecl != nil && varDecl.Name() != nil {
						nameRange := utils.TrimNodeTextRange(ctx.SourceFile, varDecl.Name())
						return ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()]
					}
				}
			}
			return ""
		}

		// Helper to report deprecation
		reportDeprecation := func(node *ast.Node, symbolName string, reason string) {
			if isAllowed(symbolName) {
				return
			}

			if reason != "" {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "deprecatedWithReason",
					Description: "'" + symbolName + "' is deprecated: " + reason,
					Data: map[string]interface{}{
						"name":   symbolName,
						"reason": reason,
					},
				})
			} else {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "deprecated",
					Description: "'" + symbolName + "' is deprecated.",
					Data: map[string]interface{}{
						"name": symbolName,
					},
				})
			}
		}

		// Helper to check if node is in a declaration position
		isDeclaration := func(node *ast.Node) bool {
			if node == nil {
				return false
			}

			parent := node.Parent
			if parent == nil {
				return false
			}

			switch parent.Kind {
			case ast.KindVariableDeclaration:
				varDecl := parent.AsVariableDeclaration()
				return varDecl != nil && varDecl.Name() == node
			case ast.KindFunctionDeclaration:
				funcDecl := parent.AsFunctionDeclaration()
				return funcDecl != nil && funcDecl.Name() == node
			case ast.KindClassDeclaration:
				classDecl := parent.AsClassDeclaration()
				return classDecl != nil && classDecl.Name() == node
			case ast.KindInterfaceDeclaration:
				interfaceDecl := parent.AsInterfaceDeclaration()
				return interfaceDecl != nil && interfaceDecl.Name() == node
			case ast.KindTypeAliasDeclaration:
				typeAlias := parent.AsTypeAliasDeclaration()
				return typeAlias != nil && typeAlias.Name() == node
			case ast.KindEnumDeclaration:
				enumDecl := parent.AsEnumDeclaration()
				return enumDecl != nil && enumDecl.Name() == node
			case ast.KindParameter:
				return true
			case ast.KindPropertyDeclaration, ast.KindPropertySignature,
				ast.KindMethodDeclaration, ast.KindMethodSignature:
				return true
			}

			return false
		}

		// Helper to check if node is inside an import
		isInsideImport := func(node *ast.Node) bool {
			current := node
			for current != nil {
				switch current.Kind {
				case ast.KindImportDeclaration, ast.KindImportEqualsDeclaration:
					return true
				case ast.KindExportDeclaration:
					exportDecl := current.AsExportDeclaration()
					// Allow if it's a re-export without a name
					if exportDecl != nil && exportDecl.ExportClause == nil {
						return false
					}
					return true
				}
				current = current.Parent
			}
			return false
		}

		return rule.RuleListeners{
			ast.KindIdentifier: func(node *ast.Node) {
				// Skip if type checker is not available
				if ctx.TypeChecker == nil {
					return
				}

				// Skip declarations
				if isDeclaration(node) {
					return
				}

				// Skip inside imports (but not usage of imported items)
				if isInsideImport(node) {
					return
				}

				// Get the symbol for this identifier
				symbol := ctx.TypeChecker.GetSymbolAtLocation(node)
				if symbol == nil {
					return
				}

				// Check if the symbol is deprecated
				deprecated, reason := isSymbolDeprecated(symbol)
				if deprecated {
					symbolName := getSymbolName(symbol)
					if symbolName == "" {
						// Try to get name from identifier node
						idRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
						symbolName = ctx.SourceFile.Text()[idRange.Pos():idRange.End()]
					}
					reportDeprecation(node, symbolName, reason)
				}
			},

			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}

				propAccess := node.AsPropertyAccessExpression()
				if propAccess == nil || propAccess.Name == nil {
					return
				}

				// Get the type of the expression being accessed
				exprType := ctx.TypeChecker.GetTypeAtLocation(propAccess.Expression)
				if exprType == nil {
					return
				}

				// Get the property name
				nameRange := utils.TrimNodeTextRange(ctx.SourceFile, propAccess.Name)
				propName := ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()]

				// Get the property symbol
				prop := ctx.TypeChecker.GetPropertyOfType(exprType, propName)
				if prop == nil {
					return
				}

				// Check if the property is deprecated
				deprecated, reason := isSymbolDeprecated(prop)
				if deprecated {
					reportDeprecation(propAccess.Name, propName, reason)
				}
			},

			ast.KindElementAccessExpression: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}

				elemAccess := node.AsElementAccessExpression()
				if elemAccess == nil || elemAccess.ArgumentExpression == nil {
					return
				}

				// Try to resolve the argument to a string literal
				argType := ctx.TypeChecker.GetTypeAtLocation(elemAccess.ArgumentExpression)
				if argType == nil {
					return
				}

				// Check if it's a string literal type
				if !utils.IsTypeFlagSet(argType, checker.TypeFlagsStringLiteral) {
					// Try to resolve template literals or const assertions
					if elemAccess.ArgumentExpression.Kind == ast.KindStringLiteral ||
						elemAccess.ArgumentExpression.Kind == ast.KindNoSubstitutionTemplateLiteral {
						// Get the literal value
						argRange := utils.TrimNodeTextRange(ctx.SourceFile, elemAccess.ArgumentExpression)
						argText := ctx.SourceFile.Text()[argRange.Pos():argRange.End()]
						// Remove quotes
						if len(argText) >= 2 {
							propName := argText[1 : len(argText)-1]

							// Get the type of the expression being accessed
							exprType := ctx.TypeChecker.GetTypeAtLocation(elemAccess.Expression)
							if exprType != nil {
								prop := ctx.TypeChecker.GetPropertyOfType(exprType, propName)
								if prop != nil {
									deprecated, reason := isSymbolDeprecated(prop)
									if deprecated {
										reportDeprecation(elemAccess.ArgumentExpression, propName, reason)
									}
								}
							}
						}
					}
					return
				}

				// Get the string value from the type
				propName := argType.AsStringLiteralType().Value

				// Get the type of the expression being accessed
				exprType := ctx.TypeChecker.GetTypeAtLocation(elemAccess.Expression)
				if exprType == nil {
					return
				}

				// Get the property symbol
				prop := ctx.TypeChecker.GetPropertyOfType(exprType, propName)
				if prop == nil {
					return
				}

				// Check if the property is deprecated
				deprecated, reason := isSymbolDeprecated(prop)
				if deprecated {
					reportDeprecation(elemAccess.ArgumentExpression, propName, reason)
				}
			},

			ast.KindNewExpression: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}

				newExpr := node.AsNewExpression()
				if newExpr == nil || newExpr.Expression == nil {
					return
				}

				// Get the symbol being instantiated
				symbol := ctx.TypeChecker.GetSymbolAtLocation(newExpr.Expression)
				if symbol == nil {
					return
				}

				// Check if the class itself is deprecated
				deprecated, reason := isSymbolDeprecated(symbol)
				if deprecated {
					symbolName := getSymbolName(symbol)
					if symbolName == "" {
						nameRange := utils.TrimNodeTextRange(ctx.SourceFile, newExpr.Expression)
						symbolName = ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()]
					}
					reportDeprecation(newExpr.Expression, symbolName, reason)
					return
				}

				// Check if the constructor is deprecated
				exprType := ctx.TypeChecker.GetTypeAtLocation(newExpr.Expression)
				if exprType == nil {
					return
				}

				// Get construct signatures
				signatures := utils.GetConstructSignatures(ctx.TypeChecker, exprType)
				for _, sig := range signatures {
					if sig.Declaration != nil {
						if deprecated, reason := getDeprecationReason(sig.Declaration); deprecated {
							symbolName := getSymbolName(symbol)
							if symbolName == "" {
								nameRange := utils.TrimNodeTextRange(ctx.SourceFile, newExpr.Expression)
								symbolName = ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()]
							}
							reportDeprecation(newExpr.Expression, symbolName, reason)
							break
						}
					}
				}
			},

			ast.KindJsxOpeningElement: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}

				openingElem := node.AsJsxOpeningElement()
				if openingElem == nil {
					return
				}

				// Check if the JSX element itself is deprecated
				tagName := openingElem.TagName
				if tagName != nil && tagName.Kind == ast.KindIdentifier {
					symbol := ctx.TypeChecker.GetSymbolAtLocation(tagName)
					if symbol != nil {
						deprecated, reason := isSymbolDeprecated(symbol)
						if deprecated {
							symbolName := getSymbolName(symbol)
							if symbolName == "" {
								nameRange := utils.TrimNodeTextRange(ctx.SourceFile, tagName)
								symbolName = ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()]
							}
							reportDeprecation(tagName, symbolName, reason)
						}
					}
				}
			},

			ast.KindJsxSelfClosingElement: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}

				selfClosing := node.AsJsxSelfClosingElement()
				if selfClosing == nil {
					return
				}

				// Check if the JSX element itself is deprecated
				tagName := selfClosing.TagName
				if tagName != nil && tagName.Kind == ast.KindIdentifier {
					symbol := ctx.TypeChecker.GetSymbolAtLocation(tagName)
					if symbol != nil {
						deprecated, reason := isSymbolDeprecated(symbol)
						if deprecated {
							symbolName := getSymbolName(symbol)
							if symbolName == "" {
								nameRange := utils.TrimNodeTextRange(ctx.SourceFile, tagName)
								symbolName = ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()]
							}
							reportDeprecation(tagName, symbolName, reason)
						}
					}
				}
			},
		}
	},
})
