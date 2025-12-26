package prefer_string_starts_ends_with

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"

	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type Options struct {
	AllowSingleElementEquality string `json:"allowSingleElementEquality"` // "never" (default) or "always"
}

func parseOptions(options any) Options {
	opts := Options{
		AllowSingleElementEquality: "never",
	}

	if options == nil {
		return opts
	}

	var optsMap map[string]interface{}
	var ok bool

	if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
		optsMap, ok = optArray[0].(map[string]interface{})
	} else {
		optsMap, ok = options.(map[string]interface{})
	}

	if ok {
		if v, ok := optsMap["allowSingleElementEquality"].(string); ok {
			opts.AllowSingleElementEquality = v
		}
	}
	return opts
}

func buildPreferStartsWithMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferStartsWith",
		Description: "Use String#startsWith method instead.",
	}
}

func buildPreferEndsWithMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferEndsWith",
		Description: "Use String#endsWith method instead.",
	}
}

// isStringType checks if a type is assignable to string (excluding any and generic types)
func isStringType(typeChecker *checker.Checker, t *checker.Type) bool {
	if t == nil {
		return false
	}

	// Exclude 'any' type
	if (checker.Type_flags(t) & checker.TypeFlagsAny) != 0 {
		return false
	}

	// Exclude generic type parameters (T extends string, etc.)
	if (checker.Type_flags(t) & checker.TypeFlagsTypeParameter) != 0 {
		return false
	}

	// Check if it's a union type
	if utils.IsUnionType(t) {
		for _, subType := range t.Types() {
			if !isStringType(typeChecker, subType) {
				return false
			}
		}
		return true
	}

	// Check if it's an intersection type - need at least one string type
	if utils.IsIntersectionType(t) {
		for _, subType := range t.Types() {
			if isStringType(typeChecker, subType) {
				return true
			}
		}
		return false
	}

	// Check for string literal types and string type
	flags := checker.Type_flags(t)
	return (flags&checker.TypeFlagsString) != 0 || (flags&checker.TypeFlagsStringLiteral) != 0
}

// getStringLiteralValue extracts the string value from a string literal node
func getStringLiteralValue(srcFile *ast.SourceFile, n *ast.Node) (string, bool) {
	if n == nil {
		return "", false
	}

	switch n.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		rng := utils.TrimNodeTextRange(srcFile, n)
		text := srcFile.Text()[rng.Pos():rng.End()]
		if len(text) >= 2 {
			quote := text[0]
			if (quote == '\'' || quote == '"' || quote == '`') && text[len(text)-1] == quote {
				return text[1 : len(text)-1], true
			}
		}
		return strings.Trim(text, "'\"`"), true
	}
	return "", false
}

// isSingleCharacter checks if a string is a single Unicode character
func isSingleCharacter(s string) bool {
	return utf8.RuneCountInString(s) == 1
}

// extractRegexLiteral extracts the pattern from a regex literal or RegExp constructor
func extractRegexLiteral(srcFile *ast.SourceFile, node *ast.Node) (pattern string, isStart bool, isEnd bool, ok bool) {
	if node == nil {
		return "", false, false, false
	}

	switch node.Kind {
	case ast.KindRegularExpressionLiteral:
		rng := utils.TrimNodeTextRange(srcFile, node)
		text := srcFile.Text()[rng.Pos():rng.End()]
		// Extract pattern between / /
		lastSlash := strings.LastIndexByte(text, '/')
		if lastSlash <= 0 {
			return "", false, false, false
		}
		pattern = text[1:lastSlash]

		// Check for ^ or $ anchors
		hasStart := strings.HasPrefix(pattern, "^")
		hasEnd := strings.HasSuffix(pattern, "$")

		// Reject patterns with both anchors, or with alternation
		if (hasStart && hasEnd) || strings.Contains(pattern, "|") {
			return "", false, false, false
		}

		// Extract the pattern without anchor
		if hasStart {
			pattern = pattern[1:]
			return pattern, true, false, true
		}
		if hasEnd {
			pattern = pattern[:len(pattern)-1]
			return pattern, false, true, true
		}

	case ast.KindNewExpression:
		newExpr := node.AsNewExpression()
		if newExpr.Expression != nil && newExpr.Expression.Kind == ast.KindIdentifier {
			if newExpr.Expression.AsIdentifier().Text() == "RegExp" {
				if newExpr.Arguments != nil && len(newExpr.Arguments.Nodes) > 0 {
					arg := newExpr.Arguments.Nodes[0]
					if val, ok := getStringLiteralValue(srcFile, arg); ok {
						// Try to parse as regex
						hasStart := strings.HasPrefix(val, "^")
						hasEnd := strings.HasSuffix(val, "$")

						if (hasStart && hasEnd) || strings.Contains(val, "|") {
							return "", false, false, false
						}

						if hasStart {
							pattern = val[1:]
							return pattern, true, false, true
						}
						if hasEnd {
							pattern = val[:len(val)-1]
							return pattern, false, true, true
						}
					}
				}
			}
		}
	}

	return "", false, false, false
}

// escapeForStringLiteral escapes a regex pattern for use in a string literal
func escapeForStringLiteral(pattern string) string {
	// Escape backslashes and quotes
	result := strings.ReplaceAll(pattern, `\`, `\\`)
	result = strings.ReplaceAll(result, `"`, `\"`)
	return result
}

var PreferStringStartsEndsWithRule = rule.CreateRule(rule.Rule{
	Name: "prefer-string-starts-ends-with",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				binExpr := node.AsBinaryExpression()
				if binExpr == nil {
					return
				}

				// Check for equality operators: ===, !==, ==, !=
				isEquality := false
				isInequality := false
				switch binExpr.OperatorToken.Kind {
				case ast.KindEqualsEqualsEqualsToken, ast.KindEqualsEqualsToken:
					isEquality = true
				case ast.KindExclamationEqualsEqualsToken, ast.KindExclamationEqualsToken:
					isInequality = true
				default:
					return
				}

				// Pattern 1: s[0] === 'a' or s[s.length - 1] === 'a'
				checkElementAccess := func(left, right *ast.Node) {
					if left.Kind == ast.KindElementAccessExpression {
						elemAccess := left.AsElementAccessExpression()
						if elemAccess.ArgumentExpression == nil {
							return
						}

						// Check if the object is a string type
						objType := ctx.TypeChecker.GetTypeAtLocation(elemAccess.Expression)
						if !isStringType(ctx.TypeChecker, objType) {
							return
						}

						// Extract the string literal being compared
						stringVal, ok := getStringLiteralValue(ctx.SourceFile, right)
						if !ok {
							return
						}

						// Check for single-element equality option
						if opts.AllowSingleElementEquality == "always" && isSingleCharacter(stringVal) {
							// Check if this is s[0] or s[s.length - 1]
							if elemAccess.ArgumentExpression.Kind == ast.KindNumericLiteral {
								numLit := elemAccess.ArgumentExpression.AsNumericLiteral()
								if numLit.Text() == "0" {
									return
								}
							}
							// Check for s.length - 1 pattern
							if elemAccess.ArgumentExpression.Kind == ast.KindBinaryExpression {
								binExp := elemAccess.ArgumentExpression.AsBinaryExpression()
								if binExp.OperatorToken.Kind == ast.KindMinusToken {
									return
								}
							}
						}

						// Check for multi-byte character (cannot be safely auto-fixed)
						if !isSingleCharacter(stringVal) {
							if isEquality {
								ctx.ReportNode(node, buildPreferStartsWithMessage())
							} else {
								ctx.ReportNode(node, buildPreferStartsWithMessage())
							}
							return
						}

						// Pattern: s[0]
						if elemAccess.ArgumentExpression.Kind == ast.KindNumericLiteral {
							numLit := elemAccess.ArgumentExpression.AsNumericLiteral()
							if numLit.Text() == "0" {
								// Generate fix: s.startsWith('a') or !s.startsWith('a')
								objText := ctx.SourceFile.Text()[utils.TrimNodeTextRange(ctx.SourceFile, elemAccess.Expression).Pos():utils.TrimNodeTextRange(ctx.SourceFile, elemAccess.Expression).End()]

								// Handle optional chaining
								optionalChain := ""
								if elemAccess.QuestionDotToken != nil {
									optionalChain = "?."
								} else {
									optionalChain = "."
								}

								replacement := fmt.Sprintf("%s%sstartsWith(%s)", objText, optionalChain, ctx.SourceFile.Text()[utils.TrimNodeTextRange(ctx.SourceFile, right).Pos():utils.TrimNodeTextRange(ctx.SourceFile, right).End()])
								if isInequality {
									replacement = "!" + replacement
								}

								fix := rule.RuleFixReplace(ctx.SourceFile, node, replacement)
								ctx.ReportNodeWithFixes(node, buildPreferStartsWithMessage(), fix)
								return
							}
						}

						// Pattern: s[s.length - 1]
						if elemAccess.ArgumentExpression.Kind == ast.KindBinaryExpression {
							indexExpr := elemAccess.ArgumentExpression.AsBinaryExpression()
							if indexExpr.OperatorToken.Kind == ast.KindMinusToken {
								// Check if left side is s.length
								if indexExpr.Left.Kind == ast.KindPropertyAccessExpression {
									propAccess := indexExpr.Left.AsPropertyAccessExpression()
									if propAccess.Name() != nil && propAccess.Name().Text() == "length" {
										// Check if right side is 1
										if indexExpr.Right.Kind == ast.KindNumericLiteral {
											numLit := indexExpr.Right.AsNumericLiteral()
											if numLit.Text() == "1" {
												// Generate fix: s.endsWith('a') or !s.endsWith('a')
												objText := ctx.SourceFile.Text()[utils.TrimNodeTextRange(ctx.SourceFile, elemAccess.Expression).Pos():utils.TrimNodeTextRange(ctx.SourceFile, elemAccess.Expression).End()]

												// Handle optional chaining
												optionalChain := "."
												if elemAccess.QuestionDotToken != nil {
													optionalChain = "?."
												}

												replacement := fmt.Sprintf("%s%sendsWith(%s)", objText, optionalChain, ctx.SourceFile.Text()[utils.TrimNodeTextRange(ctx.SourceFile, right).Pos():utils.TrimNodeTextRange(ctx.SourceFile, right).End()])
												if isInequality {
													replacement = "!" + replacement
												}

												fix := rule.RuleFixReplace(ctx.SourceFile, node, replacement)
												ctx.ReportNodeWithFixes(node, buildPreferEndsWithMessage(), fix)
												return
											}
										}
									}
								}
							}
						}
					}
				}

				// Pattern 2: s.charAt(0) === 'a' or s.charAt(s.length - 1) === 'a'
				checkCharAt := func(left, right *ast.Node) {
					if left.Kind == ast.KindCallExpression {
						callExpr := left.AsCallExpression()
						if callExpr.Expression == nil {
							return
						}

						// Check if it's .charAt() call
						if callExpr.Expression.Kind == ast.KindPropertyAccessExpression {
							propAccess := callExpr.Expression.AsPropertyAccessExpression()
							if propAccess.Name() == nil || propAccess.Name().Text() != "charAt" {
								return
							}

							// Check if the object is a string type
							objType := ctx.TypeChecker.GetTypeAtLocation(propAccess.Expression)
							if !isStringType(ctx.TypeChecker, objType) {
								return
							}

							// Extract the string literal being compared
							stringVal, ok := getStringLiteralValue(ctx.SourceFile, right)
							if !ok {
								return
							}

							// Check for multi-byte character
							if !isSingleCharacter(stringVal) {
								ctx.ReportNode(node, buildPreferStartsWithMessage())
								return
							}

							// Get charAt argument
							if callExpr.Arguments == nil || len(callExpr.Arguments.Nodes) == 0 {
								return
							}
							arg := callExpr.Arguments.Nodes[0]

							// Pattern: s.charAt(0)
							if arg.Kind == ast.KindNumericLiteral {
								numLit := arg.AsNumericLiteral()
								if numLit.Text() == "0" {
									objText := ctx.SourceFile.Text()[utils.TrimNodeTextRange(ctx.SourceFile, propAccess.Expression).Pos():utils.TrimNodeTextRange(ctx.SourceFile, propAccess.Expression).End()]

									// Handle optional chaining
									optionalChain := "."
									if callExpr.QuestionDotToken != nil {
										optionalChain = "?."
									}

									replacement := fmt.Sprintf("%s%sstartsWith(%s)", objText, optionalChain, ctx.SourceFile.Text()[utils.TrimNodeTextRange(ctx.SourceFile, right).Pos():utils.TrimNodeTextRange(ctx.SourceFile, right).End()])
									if isInequality {
										replacement = "!" + replacement
									}

									fix := rule.RuleFixReplace(ctx.SourceFile, node, replacement)
									ctx.ReportNodeWithFixes(node, buildPreferStartsWithMessage(), fix)
									return
								}
							}

							// Pattern: s.charAt(s.length - 1)
							if arg.Kind == ast.KindBinaryExpression {
								binExp := arg.AsBinaryExpression()
								if binExp.OperatorToken.Kind == ast.KindMinusToken {
									if binExp.Left.Kind == ast.KindPropertyAccessExpression {
										propAcc := binExp.Left.AsPropertyAccessExpression()
										if propAcc.Name() != nil && propAcc.Name().Text() == "length" {
											if binExp.Right.Kind == ast.KindNumericLiteral {
												numLit := binExp.Right.AsNumericLiteral()
												if numLit.Text() == "1" {
													objText := ctx.SourceFile.Text()[utils.TrimNodeTextRange(ctx.SourceFile, propAccess.Expression).Pos():utils.TrimNodeTextRange(ctx.SourceFile, propAccess.Expression).End()]

													optionalChain := "."
													if callExpr.QuestionDotToken != nil {
														optionalChain = "?."
													}

													replacement := fmt.Sprintf("%s%sendsWith(%s)", objText, optionalChain, ctx.SourceFile.Text()[utils.TrimNodeTextRange(ctx.SourceFile, right).Pos():utils.TrimNodeTextRange(ctx.SourceFile, right).End()])
													if isInequality {
														replacement = "!" + replacement
													}

													fix := rule.RuleFixReplace(ctx.SourceFile, node, replacement)
													ctx.ReportNodeWithFixes(node, buildPreferEndsWithMessage(), fix)
													return
												}
											}
										}
									}
								}
							}
						}
					}
				}

				// Pattern 3: s.indexOf(needle) === 0
				checkIndexOf := func(left, right *ast.Node) {
					if left.Kind == ast.KindCallExpression {
						callExpr := left.AsCallExpression()
						if callExpr.Expression == nil {
							return
						}

						if callExpr.Expression.Kind == ast.KindPropertyAccessExpression {
							propAccess := callExpr.Expression.AsPropertyAccessExpression()
							if propAccess.Name() == nil || propAccess.Name().Text() != "indexOf" {
								return
							}

							// Check if the object is a string type
							objType := ctx.TypeChecker.GetTypeAtLocation(propAccess.Expression)
							if !isStringType(ctx.TypeChecker, objType) {
								return
							}

							// Check if comparing to 0
							if right.Kind == ast.KindNumericLiteral {
								numLit := right.AsNumericLiteral()
								if numLit.Text() == "0" {
									objText := ctx.SourceFile.Text()[utils.TrimNodeTextRange(ctx.SourceFile, propAccess.Expression).Pos():utils.TrimNodeTextRange(ctx.SourceFile, propAccess.Expression).End()]

									// Get the needle argument
									needleText := ""
									if callExpr.Arguments != nil && len(callExpr.Arguments.Nodes) > 0 {
										needleText = ctx.SourceFile.Text()[utils.TrimNodeTextRange(ctx.SourceFile, callExpr.Arguments.Nodes[0]).Pos():utils.TrimNodeTextRange(ctx.SourceFile, callExpr.Arguments.Nodes[0]).End()]
									}

									optionalChain := "."
									if callExpr.QuestionDotToken != nil {
										optionalChain = "?."
									}

									replacement := fmt.Sprintf("%s%sstartsWith(%s)", objText, optionalChain, needleText)
									if isInequality {
										replacement = "!" + replacement
									}

									fix := rule.RuleFixReplace(ctx.SourceFile, node, replacement)
									ctx.ReportNodeWithFixes(node, buildPreferStartsWithMessage(), fix)
									return
								}
							}
						}
					}
				}

				// Pattern 4: s.lastIndexOf('bar') === s.length - 3
				checkLastIndexOf := func(left, right *ast.Node) {
					if left.Kind == ast.KindCallExpression {
						callExpr := left.AsCallExpression()
						if callExpr.Expression == nil {
							return
						}

						if callExpr.Expression.Kind == ast.KindPropertyAccessExpression {
							propAccess := callExpr.Expression.AsPropertyAccessExpression()
							if propAccess.Name() == nil || propAccess.Name().Text() != "lastIndexOf" {
								return
							}

							// Check if the object is a string type
							objType := ctx.TypeChecker.GetTypeAtLocation(propAccess.Expression)
							if !isStringType(ctx.TypeChecker, objType) {
								return
							}

							// Get the needle argument
							if callExpr.Arguments == nil || len(callExpr.Arguments.Nodes) == 0 {
								return
							}
							needleArg := callExpr.Arguments.Nodes[0]
							needleText := ctx.SourceFile.Text()[utils.TrimNodeTextRange(ctx.SourceFile, needleArg).Pos():utils.TrimNodeTextRange(ctx.SourceFile, needleArg).End()]

							// Check if right side is s.length - needle.length or s.length - 3
							if right.Kind == ast.KindBinaryExpression {
								binExp := right.AsBinaryExpression()
								if binExp.OperatorToken.Kind == ast.KindMinusToken {
									if binExp.Left.Kind == ast.KindPropertyAccessExpression {
										propAcc := binExp.Left.AsPropertyAccessExpression()
										if propAcc.Name() != nil && propAcc.Name().Text() == "length" {
											// Valid pattern found
											objText := ctx.SourceFile.Text()[utils.TrimNodeTextRange(ctx.SourceFile, propAccess.Expression).Pos():utils.TrimNodeTextRange(ctx.SourceFile, propAccess.Expression).End()]

											optionalChain := "."
											if callExpr.QuestionDotToken != nil {
												optionalChain = "?."
											}

											replacement := fmt.Sprintf("%s%sendsWith(%s)", objText, optionalChain, needleText)
											if isInequality {
												replacement = "!" + replacement
											}

											fix := rule.RuleFixReplace(ctx.SourceFile, node, replacement)
											ctx.ReportNodeWithFixes(node, buildPreferEndsWithMessage(), fix)
											return
										}
									}
								}
							}
						}
					}
				}

				// Pattern 5: s.slice(0, 3) === 'bar'
				checkSlice := func(left, right *ast.Node) {
					if left.Kind == ast.KindCallExpression {
						callExpr := left.AsCallExpression()
						if callExpr.Expression == nil {
							return
						}

						if callExpr.Expression.Kind == ast.KindPropertyAccessExpression {
							propAccess := callExpr.Expression.AsPropertyAccessExpression()
							if propAccess.Name() == nil {
								return
							}
							methodName := propAccess.Name().Text()
							if methodName != "slice" && methodName != "substring" {
								return
							}

							// Check if the object is a string type
							objType := ctx.TypeChecker.GetTypeAtLocation(propAccess.Expression)
							if !isStringType(ctx.TypeChecker, objType) {
								return
							}

							if callExpr.Arguments == nil || len(callExpr.Arguments.Nodes) == 0 {
								return
							}

							args := callExpr.Arguments.Nodes
							objText := ctx.SourceFile.Text()[utils.TrimNodeTextRange(ctx.SourceFile, propAccess.Expression).Pos():utils.TrimNodeTextRange(ctx.SourceFile, propAccess.Expression).End()]

							optionalChain := "."
							if callExpr.QuestionDotToken != nil {
								optionalChain = "?."
							}

							// Pattern: s.slice(0, 3) === 'bar' or s.slice(0, needle.length) === needle
							if len(args) >= 2 {
								if args[0].Kind == ast.KindNumericLiteral {
									numLit := args[0].AsNumericLiteral()
									if numLit.Text() == "0" {
										// Get the comparison value
										rightText := ctx.SourceFile.Text()[utils.TrimNodeTextRange(ctx.SourceFile, right).Pos():utils.TrimNodeTextRange(ctx.SourceFile, right).End()]

										// Check if we can auto-fix
										canFix := true

										// If using == (loose equality) with needle.length, don't auto-fix
										if binExpr.OperatorToken.Kind == ast.KindEqualsEqualsToken || binExpr.OperatorToken.Kind == ast.KindExclamationEqualsToken {
											if args[1].Kind == ast.KindPropertyAccessExpression {
												canFix = false
											}
										}

										if canFix {
											replacement := fmt.Sprintf("%s%sstartsWith(%s)", objText, optionalChain, rightText)
											if isInequality {
												replacement = "!" + replacement
											}

											fix := rule.RuleFixReplace(ctx.SourceFile, node, replacement)
											ctx.ReportNodeWithFixes(node, buildPreferStartsWithMessage(), fix)
										} else {
											ctx.ReportNode(node, buildPreferStartsWithMessage())
										}
										return
									}
								}
							}

							// Pattern: s.slice(-3) === 'bar' or s.slice(-needle.length) === needle
							if len(args) >= 1 {
								// Check for negative number or unary minus
								isNegative := false
								if args[0].Kind == ast.KindPrefixUnaryExpression {
									prefixExpr := args[0].AsPrefixUnaryExpression()
									if prefixExpr.Operator == ast.KindMinusToken {
										isNegative = true
									}
								} else if args[0].Kind == ast.KindNumericLiteral {
									numLit := args[0].AsNumericLiteral()
									if strings.HasPrefix(numLit.Text(), "-") {
										isNegative = true
									}
								}

								if isNegative {
									rightText := ctx.SourceFile.Text()[utils.TrimNodeTextRange(ctx.SourceFile, right).Pos():utils.TrimNodeTextRange(ctx.SourceFile, right).End()]

									replacement := fmt.Sprintf("%s%sendsWith(%s)", objText, optionalChain, rightText)
									if isInequality {
										replacement = "!" + replacement
									}

									fix := rule.RuleFixReplace(ctx.SourceFile, node, replacement)
									ctx.ReportNodeWithFixes(node, buildPreferEndsWithMessage(), fix)
									return
								}

								// Pattern: s.slice(s.length - needle.length) === needle
								if args[0].Kind == ast.KindBinaryExpression {
									binExp := args[0].AsBinaryExpression()
									if binExp.OperatorToken.Kind == ast.KindMinusToken {
										if binExp.Left.Kind == ast.KindPropertyAccessExpression {
											propAcc := binExp.Left.AsPropertyAccessExpression()
											if propAcc.Name() != nil && propAcc.Name().Text() == "length" {
												rightText := ctx.SourceFile.Text()[utils.TrimNodeTextRange(ctx.SourceFile, right).Pos():utils.TrimNodeTextRange(ctx.SourceFile, right).End()]

												replacement := fmt.Sprintf("%s%sendsWith(%s)", objText, optionalChain, rightText)
												if isInequality {
													replacement = "!" + replacement
												}

												fix := rule.RuleFixReplace(ctx.SourceFile, node, replacement)
												ctx.ReportNodeWithFixes(node, buildPreferEndsWithMessage(), fix)
												return
											}
										}
									}
								}
							}

							// Pattern: s.substring(-3) === 'bar' (probably a mistake, no auto-fix)
							if methodName == "substring" && len(args) >= 1 {
								if args[0].Kind == ast.KindPrefixUnaryExpression {
									prefixExpr := args[0].AsPrefixUnaryExpression()
									if prefixExpr.Operator == ast.KindMinusToken {
										ctx.ReportNode(node, buildPreferEndsWithMessage())
										return
									}
								} else if args[0].Kind == ast.KindNumericLiteral {
									numLit := args[0].AsNumericLiteral()
									if strings.HasPrefix(numLit.Text(), "-") {
										ctx.ReportNode(node, buildPreferEndsWithMessage())
										return
									}
								}

								// Pattern: s.substring(s.length - 3, s.length) === 'bar'
								if len(args) >= 2 {
									if args[0].Kind == ast.KindBinaryExpression && args[1].Kind == ast.KindPropertyAccessExpression {
										binExp := args[0].AsBinaryExpression()
										propAcc := args[1].AsPropertyAccessExpression()
										if binExp.OperatorToken.Kind == ast.KindMinusToken && propAcc.Name() != nil && propAcc.Name().Text() == "length" {
											rightText := ctx.SourceFile.Text()[utils.TrimNodeTextRange(ctx.SourceFile, right).Pos():utils.TrimNodeTextRange(ctx.SourceFile, right).End()]

											replacement := fmt.Sprintf("%s%sendsWith(%s)", objText, optionalChain, rightText)
											if isInequality {
												replacement = "!" + replacement
											}

											fix := rule.RuleFixReplace(ctx.SourceFile, node, replacement)
											ctx.ReportNodeWithFixes(node, buildPreferEndsWithMessage(), fix)
											return
										}
									}
								}
							}
						}
					}
				}

				// Pattern 6: s.match(/^foo/) !== null
				checkMatch := func(left, right *ast.Node) {
					if left.Kind == ast.KindCallExpression {
						callExpr := left.AsCallExpression()
						if callExpr.Expression == nil {
							return
						}

						if callExpr.Expression.Kind == ast.KindPropertyAccessExpression {
							propAccess := callExpr.Expression.AsPropertyAccessExpression()
							if propAccess.Name() == nil || propAccess.Name().Text() != "match" {
								return
							}

							// Check if the object is a string type
							objType := ctx.TypeChecker.GetTypeAtLocation(propAccess.Expression)
							if !isStringType(ctx.TypeChecker, objType) {
								return
							}

							// Check if comparing to null
							if right.Kind != ast.KindNullKeyword {
								return
							}

							// Get the regex argument
							if callExpr.Arguments == nil || len(callExpr.Arguments.Nodes) == 0 {
								return
							}
							regexArg := callExpr.Arguments.Nodes[0]

							// Extract regex pattern
							pattern, isStart, isEnd, ok := extractRegexLiteral(ctx.SourceFile, regexArg)
							if !ok || (!isStart && !isEnd) {
								return
							}

							objText := ctx.SourceFile.Text()[utils.TrimNodeTextRange(ctx.SourceFile, propAccess.Expression).Pos():utils.TrimNodeTextRange(ctx.SourceFile, propAccess.Expression).End()]

							optionalChain := "."
							if callExpr.QuestionDotToken != nil {
								optionalChain = "?."
							}

							// Escape pattern for use in string literal
							escapedPattern := escapeForStringLiteral(pattern)

							var replacement string
							if isStart {
								replacement = fmt.Sprintf("%s%sstartsWith(\"%s\")", objText, optionalChain, escapedPattern)
							} else {
								replacement = fmt.Sprintf("%s%sendsWith(\"%s\")", objText, optionalChain, escapedPattern)
							}

							// Handle !== null vs === null
							if isEquality {
								replacement = "!" + replacement
							}

							fix := rule.RuleFixReplace(ctx.SourceFile, node, replacement)
							if isStart {
								ctx.ReportNodeWithFixes(node, buildPreferStartsWithMessage(), fix)
							} else {
								ctx.ReportNodeWithFixes(node, buildPreferEndsWithMessage(), fix)
							}
							return
						}
					}
				}

				// Try all patterns
				checkElementAccess(binExpr.Left, binExpr.Right)
				checkElementAccess(binExpr.Right, binExpr.Left)
				checkCharAt(binExpr.Left, binExpr.Right)
				checkCharAt(binExpr.Right, binExpr.Left)
				checkIndexOf(binExpr.Left, binExpr.Right)
				checkIndexOf(binExpr.Right, binExpr.Left)
				checkLastIndexOf(binExpr.Left, binExpr.Right)
				checkLastIndexOf(binExpr.Right, binExpr.Left)
				checkSlice(binExpr.Left, binExpr.Right)
				checkSlice(binExpr.Right, binExpr.Left)
				checkMatch(binExpr.Left, binExpr.Right)
				checkMatch(binExpr.Right, binExpr.Left)
			},

			// Pattern 7: /^foo/.test(s) or /foo$/.test(s)
			ast.KindCallExpression: func(node *ast.Node) {
				callExpr := node.AsCallExpression()
				if callExpr.Expression == nil {
					return
				}

				// Check for pattern.test(s) or /^foo/.test(s)
				if callExpr.Expression.Kind == ast.KindPropertyAccessExpression {
					propAccess := callExpr.Expression.AsPropertyAccessExpression()
					if propAccess.Name() == nil || propAccess.Name().Text() != "test" {
						return
					}

					// Get the regex object
					regexNode := propAccess.Expression

					// Extract regex pattern
					pattern, isStart, isEnd, ok := extractRegexLiteral(ctx.SourceFile, regexNode)
					if !ok || (!isStart && !isEnd) {
						return
					}

					// Get the string argument
					if callExpr.Arguments == nil || len(callExpr.Arguments.Nodes) == 0 {
						return
					}
					stringArg := callExpr.Arguments.Nodes[0]

					// Check if the argument is a string type
					argType := ctx.TypeChecker.GetTypeAtLocation(stringArg)
					if !isStringType(ctx.TypeChecker, argType) {
						return
					}

					stringText := ctx.SourceFile.Text()[utils.TrimNodeTextRange(ctx.SourceFile, stringArg).Pos():utils.TrimNodeTextRange(ctx.SourceFile, stringArg).End()]

					// Check if we need to wrap in parentheses (for binary expressions like a + b)
					needsParens := false
					if stringArg.Kind == ast.KindBinaryExpression {
						needsParens = true
					}

					if needsParens {
						stringText = "(" + stringText + ")"
					}

					optionalChain := "."
					if propAccess.QuestionDotToken != nil {
						optionalChain = "?."
					}

					// Escape pattern for use in string literal
					escapedPattern := escapeForStringLiteral(pattern)

					var replacement string
					if isStart {
						replacement = fmt.Sprintf("%s%sstartsWith(\"%s\")", stringText, optionalChain, escapedPattern)
					} else {
						replacement = fmt.Sprintf("%s%sendsWith(\"%s\")", stringText, optionalChain, escapedPattern)
					}

					fix := rule.RuleFixReplace(ctx.SourceFile, node, replacement)
					if isStart {
						ctx.ReportNodeWithFixes(node, buildPreferStartsWithMessage(), fix)
					} else {
						ctx.ReportNodeWithFixes(node, buildPreferEndsWithMessage(), fix)
					}
					return
				}
			},
		}
	},
})
