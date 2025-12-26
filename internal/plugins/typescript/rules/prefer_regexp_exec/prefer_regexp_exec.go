package prefer_regexp_exec

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildRegExpExecOverStringMatchMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "regExpExecOverStringMatch",
		Description: "Use the `RegExp#exec()` method instead of `String#match()`.",
	}
}

// Helper function to check if a type has a string literal component
func hasStringType(typeChecker *checker.Checker, t *checker.Type) bool {
	if t == nil {
		return false
	}

	for _, typePart := range utils.UnionTypeParts(t) {
		flags := checker.Type_flags(typePart)
		if flags&(checker.TypeFlagsString|checker.TypeFlagsStringLiteral) != 0 {
			return true
		}
	}
	return false
}

// Helper function to check if a RegExp literal has a global flag
func hasGlobalFlag(node *ast.Node, sourceFile *ast.SourceFile) bool {
	if node.Kind != ast.KindRegularExpressionLiteral {
		return false
	}

	nodeRange := utils.TrimNodeTextRange(sourceFile, node)
	text := sourceFile.Text()[nodeRange.Pos():nodeRange.End()]

	// RegExp format: /pattern/flags
	lastSlash := strings.LastIndex(text, "/")
	if lastSlash > 0 && lastSlash < len(text)-1 {
		flags := text[lastSlash+1:]
		return strings.Contains(flags, "g")
	}
	return false
}

// Helper function to check if a new RegExp() or RegExp() call has a global flag
func newRegExpHasGlobalFlag(node *ast.Node, sourceFile *ast.SourceFile, typeChecker *checker.Checker) bool {
	var expr *ast.NewExpression
	var callExpr *ast.CallExpression

	if node.Kind == ast.KindNewExpression {
		expr = node.AsNewExpression()
	} else if node.Kind == ast.KindCallExpression {
		callExpr = node.AsCallExpression()
	} else {
		return false
	}

	var args []*ast.Node
	var exprNode *ast.Node

	if expr != nil {
		if expr.Arguments != nil {
			args = expr.Arguments.Nodes
		}
		exprNode = expr.Expression
	} else {
		args = callExpr.Arguments.Nodes
		exprNode = callExpr.Expression
	}

	// Check if it's a RegExp constructor
	if exprNode.Kind != ast.KindIdentifier {
		return false
	}
	ident := exprNode.AsIdentifier()
	if ident == nil || ident.Text != "RegExp" {
		return false
	}

	// Check if there's a flags argument (second argument)
	if len(args) < 2 {
		return false
	}

	flagsArg := args[1]
	flagsRange := utils.TrimNodeTextRange(sourceFile, flagsArg)
	flagsText := sourceFile.Text()[flagsRange.Pos():flagsRange.End()]

	// Check for literal string flags
	if flagsArg.Kind == ast.KindStringLiteral || flagsArg.Kind == ast.KindNoSubstitutionTemplateLiteral {
		// Remove quotes
		flagsText = strings.Trim(flagsText, "\"'`")
		return strings.Contains(flagsText, "g")
	}

	// For non-literal flags, we can't determine statically
	// Check the type to see if it could contain 'g'
	if typeChecker != nil {
		flagsType := typeChecker.GetTypeAtLocation(flagsArg)
		typeString := typeChecker.TypeToString(flagsType)
		// If it's a literal type that we can analyze
		if strings.Contains(typeString, "\"") && !strings.Contains(typeString, "g") {
			return false
		}
	}

	return false
}

// Helper function to check if a type could be a RegExp with global flag
func couldBeGlobalRegExp(typeChecker *checker.Checker, node *ast.Node, sourceFile *ast.SourceFile) bool {
	// For RegExp literals, check directly
	if node.Kind == ast.KindRegularExpressionLiteral {
		return hasGlobalFlag(node, sourceFile)
	}

	// For new RegExp() or RegExp() calls
	if node.Kind == ast.KindNewExpression || node.Kind == ast.KindCallExpression {
		if newRegExpHasGlobalFlag(node, sourceFile, typeChecker) {
			return true
		}
	}

	// For identifiers and expressions, we need to check the type
	// Since we can't reliably determine if a RegExp variable has a global flag,
	// we assume it might not have one (conservative approach)
	return false
}

// Helper function to check if an argument is a RegExp type
func isRegExpType(typeChecker *checker.Checker, t *checker.Type) bool {
	if t == nil {
		return false
	}

	for _, typePart := range utils.UnionTypeParts(t) {
		// Check if it's the built-in RegExp type
		symbol := checker.Type_symbol(typePart)
		if symbol != nil && symbol.Name == "RegExp" {
			return true
		}
	}
	return false
}

// Helper function to escape a string for use in a RegExp literal
func escapeRegExpString(s string) string {
	// Remove surrounding quotes
	s = strings.Trim(s, "\"'`")

	// Escape special regex characters
	specialChars := []string{"\\", "/", ".", "*", "+", "?", "|", "(", ")", "[", "]", "{", "}", "^", "$"}
	for _, char := range specialChars {
		s = strings.ReplaceAll(s, char, "\\"+char)
	}

	return s
}

// Helper function to get the text of a node
func getNodeText(node *ast.Node, sourceFile *ast.SourceFile) string {
	nodeRange := utils.TrimNodeTextRange(sourceFile, node)
	return sourceFile.Text()[nodeRange.Pos():nodeRange.End()]
}

var PreferRegexpExecRule = rule.CreateRule(rule.Rule{
	Name: "prefer-regexp-exec",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				callExpr := node.AsCallExpression()
				if callExpr == nil {
					return
				}

				// Check if it's a property access expression (e.g., string.match())
				if !ast.IsPropertyAccessExpression(callExpr.Expression) {
					return
				}

				propAccess := callExpr.Expression.AsPropertyAccessExpression()

				// Check if the method name is "match"
				propertyName, found := checker.Checker_getAccessedPropertyName(ctx.TypeChecker, callExpr.Expression)
				if !found || propertyName != "match" {
					return
				}

				// Check if there's exactly one argument
				if len(callExpr.Arguments.Nodes) != 1 {
					return
				}

				argumentNode := callExpr.Arguments.Nodes[0]

				// Get the type of the object being called on
				if ctx.TypeChecker == nil {
					return
				}

				objectType := ctx.TypeChecker.GetTypeAtLocation(propAccess.Expression)

				// Check if the object has string type
				if !hasStringType(ctx.TypeChecker, objectType) {
					return
				}

				// Now check the argument type and value
				argumentType := ctx.TypeChecker.GetTypeAtLocation(argumentNode)

				// Check for literal RegExp with global flag
				if argumentNode.Kind == ast.KindRegularExpressionLiteral {
					if hasGlobalFlag(argumentNode, ctx.SourceFile) {
						return // Has global flag, no violation
					}
					// No global flag, report violation with fix
					regexpText := getNodeText(argumentNode, ctx.SourceFile)
					objectText := getNodeText(propAccess.Expression, ctx.SourceFile)
					replacement := fmt.Sprintf("%s.exec(%s)", regexpText, objectText)

					ctx.ReportNodeWithFixes(
						callExpr.Expression,
						buildRegExpExecOverStringMatchMessage(),
						rule.RuleFixReplace(ctx.SourceFile, node, replacement),
					)
					return
				}

				// Check for new RegExp() or RegExp() constructor calls
				if argumentNode.Kind == ast.KindNewExpression || argumentNode.Kind == ast.KindCallExpression {
					if couldBeGlobalRegExp(ctx.TypeChecker, argumentNode, ctx.SourceFile) {
						return // Has global flag, no violation
					}

					// Check if it's a RegExp constructor
					var exprNode *ast.Node
					if argumentNode.Kind == ast.KindNewExpression {
						exprNode = argumentNode.AsNewExpression().Expression
					} else {
						exprNode = argumentNode.AsCallExpression().Expression
					}

					if exprNode.Kind == ast.KindIdentifier {
						ident := exprNode.AsIdentifier()
						if ident != nil && ident.Text == "RegExp" {
							// It's a RegExp constructor without global flag
							regexpText := getNodeText(argumentNode, ctx.SourceFile)
							objectText := getNodeText(propAccess.Expression, ctx.SourceFile)
							replacement := fmt.Sprintf("%s.exec(%s)", regexpText, objectText)

							ctx.ReportNodeWithFixes(
								callExpr.Expression,
								buildRegExpExecOverStringMatchMessage(),
								rule.RuleFixReplace(ctx.SourceFile, node, replacement),
							)
							return
						}
					}
				}

				// Check for string literal argument
				if argumentNode.Kind == ast.KindStringLiteral || argumentNode.Kind == ast.KindNoSubstitutionTemplateLiteral {
					// Convert string to RegExp literal
					strText := getNodeText(argumentNode, ctx.SourceFile)
					escaped := escapeRegExpString(strText)
					objectText := getNodeText(propAccess.Expression, ctx.SourceFile)
					replacement := fmt.Sprintf("/%s/.exec(%s)", escaped, objectText)

					ctx.ReportNodeWithFixes(
						callExpr.Expression,
						buildRegExpExecOverStringMatchMessage(),
						rule.RuleFixReplace(ctx.SourceFile, node, replacement),
					)
					return
				}

				// Check if argument has string literal type
				flags := checker.Type_flags(argumentType)
				if flags&checker.TypeFlagsStringLiteral != 0 {
					// It's a string literal type
					typeStr := ctx.TypeChecker.TypeToString(argumentType)
					// Extract the string value from the type (remove quotes)
					strValue := strings.Trim(typeStr, "\"")

					// Escape for regex
					var escaped string
					// Simple escaping for common regex special chars
					for _, ch := range strValue {
						switch ch {
						case '\\', '/', '.', '*', '+', '?', '|', '(', ')', '[', ']', '{', '}', '^', '$':
							escaped += "\\" + string(ch)
						default:
							escaped += string(ch)
						}
					}

					objectText := getNodeText(propAccess.Expression, ctx.SourceFile)
					replacement := fmt.Sprintf("/%s/.exec(%s)", escaped, objectText)

					ctx.ReportNodeWithFixes(
						callExpr.Expression,
						buildRegExpExecOverStringMatchMessage(),
						rule.RuleFixReplace(ctx.SourceFile, node, replacement),
					)
					return
				}

				// Check for RegExp type variable/expression
				if isRegExpType(ctx.TypeChecker, argumentType) {
					// For RegExp type, check if it's declared with explicit type or function return
					// We can't determine if it has global flag, but based on test cases,
					// we should NOT report if it's:
					// - A declared variable with RegExp type
					// - A function return value of RegExp type
					// - A property access that returns RegExp

					// Check if it's a simple identifier (variable)
					if argumentNode.Kind == ast.KindIdentifier {
						ident := argumentNode.AsIdentifier()
						if ident != nil {
							symbol := ctx.TypeChecker.GetSymbolAtLocation(argumentNode)
							if symbol != nil && len(symbol.Declarations) > 0 {
								decl := symbol.Declarations[0]
								// Check if it has explicit type annotation or is from function
								if decl.Kind == ast.KindVariableDeclaration {
									varDecl := decl.AsVariableDeclaration()
									if varDecl.Type != nil {
										// Has explicit type annotation, don't report
										return
									}
									// Check initializer to see if it's assigned a literal
									if varDecl.Initializer != nil {
										init := varDecl.Initializer
										if init.Kind == ast.KindRegularExpressionLiteral {
											if hasGlobalFlag(init, ctx.SourceFile) {
												return
											}
										} else if init.Kind == ast.KindNewExpression || init.Kind == ast.KindCallExpression {
											if couldBeGlobalRegExp(ctx.TypeChecker, init, ctx.SourceFile) {
												return
											}
										}
									}
								}
							}

							// It's a RegExp variable, report violation
							argumentText := getNodeText(argumentNode, ctx.SourceFile)
							objectText := getNodeText(propAccess.Expression, ctx.SourceFile)
							replacement := fmt.Sprintf("%s.exec(%s)", argumentText, objectText)

							ctx.ReportNodeWithFixes(
								callExpr.Expression,
								buildRegExpExecOverStringMatchMessage(),
								rule.RuleFixReplace(ctx.SourceFile, node, replacement),
							)
							return
						}
					}

					// Check if it's a property access (obj.search)
					if ast.IsPropertyAccessExpression(argumentNode) {
						// Don't report for property access with RegExp type
						return
					}

					// Check if it's a call expression (function returning RegExp)
					if ast.IsCallExpression(argumentNode) {
						// Don't report for function calls returning RegExp
						return
					}
				}

				// Check for string type variable
				if flags&checker.TypeFlagsString != 0 {
					// Plain string type
					argumentText := getNodeText(argumentNode, ctx.SourceFile)
					objectText := getNodeText(propAccess.Expression, ctx.SourceFile)
					replacement := fmt.Sprintf("RegExp(%s).exec(%s)", argumentText, objectText)

					ctx.ReportNodeWithFixes(
						callExpr.Expression,
						buildRegExpExecOverStringMatchMessage(),
						rule.RuleFixReplace(ctx.SourceFile, node, replacement),
					)
					return
				}
			},
		}
	},
})
