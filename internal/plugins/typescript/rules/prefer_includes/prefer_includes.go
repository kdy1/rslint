package prefer_includes

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildPreferIncludesMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferIncludes",
		Description: "Use 'includes()' method instead.",
	}
}

func buildPreferStringIncludesMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferStringIncludes",
		Description: "Use 'String#includes()' method instead.",
	}
}

// hasOptionalChain checks if a node or its ancestors contain optional chaining
func hasOptionalChain(node *ast.Node) bool {
	if node == nil {
		return false
	}

	// Check for QuestionDotToken in the node
	if ast.IsPropertyAccessExpression(node) {
		propAccess := node.AsPropertyAccessExpression()
		if propAccess != nil && propAccess.QuestionDotToken != nil {
			return true
		}
	}

	if ast.IsCallExpression(node) {
		callExpr := node.AsCallExpression()
		if callExpr != nil && callExpr.QuestionDotToken != nil {
			return true
		}
	}

	if ast.IsElementAccessExpression(node) {
		elemAccess := node.AsElementAccessExpression()
		if elemAccess != nil && elemAccess.QuestionDotToken != nil {
			return true
		}
	}

	// Check in the expression recursively
	if ast.IsPropertyAccessExpression(node) {
		propAccess := node.AsPropertyAccessExpression()
		if propAccess != nil {
			return hasOptionalChain(propAccess.Expression)
		}
	}

	if ast.IsCallExpression(node) {
		callExpr := node.AsCallExpression()
		if callExpr != nil {
			return hasOptionalChain(callExpr.Expression)
		}
	}

	return false
}

// getTextForNode returns the text of a node
func getTextForNode(sourceFile *ast.SourceFile, node *ast.Node) string {
	if node == nil || sourceFile == nil {
		return ""
	}
	nodeRange := utils.TrimNodeTextRange(sourceFile, node)
	return sourceFile.Text()[nodeRange.Pos():nodeRange.End()]
}

// parseRegExpPattern extracts the pattern from a RegExp literal or constructor
func parseRegExpPattern(node *ast.Node, ctx rule.RuleContext) (pattern string, flags string, ok bool) {
	if ast.IsRegularExpressionLiteral(node) {
		regexLit := node.AsRegularExpressionLiteral()
		if regexLit == nil {
			return "", "", false
		}
		text := getTextForNode(ctx.SourceFile, node)
		// Parse /pattern/flags format
		if len(text) < 2 || text[0] != '/' {
			return "", "", false
		}
		// Find the closing slash
		lastSlash := strings.LastIndex(text[1:], "/")
		if lastSlash == -1 {
			return "", "", false
		}
		lastSlash += 1 // Adjust for substring offset
		pattern = text[1:lastSlash]
		flags = text[lastSlash+1:]
		return pattern, flags, true
	}

	if ast.IsNewExpression(node) {
		newExpr := node.AsNewExpression()
		if newExpr == nil || newExpr.Expression == nil {
			return "", "", false
		}
		// Check if it's new RegExp(...)
		if ast.IsIdentifier(newExpr.Expression) {
			ident := newExpr.Expression.AsIdentifier()
			if ident != nil && ident.EscapedText == "RegExp" {
				// Get the first argument as pattern
				if newExpr.Arguments != nil && len(newExpr.Arguments.Slice()) > 0 {
					args := newExpr.Arguments.Slice()
					firstArg := args[0]
					if ast.IsStringLiteral(firstArg) {
						strLit := firstArg.AsStringLiteral()
						if strLit != nil {
							pattern = strLit.Text
							// Second argument would be flags
							if len(args) > 1 && ast.IsStringLiteral(args[1]) {
								flagsLit := args[1].AsStringLiteral()
								if flagsLit != nil {
									flags = flagsLit.Text
								}
							}
							return pattern, flags, true
						}
					}
				}
			}
		}
	}

	return "", "", false
}

// isSimpleRegExpPattern checks if a regex pattern is simple enough to convert to includes
// Simple patterns are just literal strings without special regex characters (except escaped ones)
func isSimpleRegExpPattern(pattern string, flags string) bool {
	// Cannot convert if it has flags (other than empty)
	if flags != "" {
		return false
	}

	// Check for regex special characters that make it non-simple
	// We allow escaped characters
	specialChars := `[]{}()*+?|^$.\`
	inEscape := false
	for _, ch := range pattern {
		if inEscape {
			inEscape = false
			continue
		}
		if ch == '\\' {
			inEscape = true
			continue
		}
		if strings.ContainsRune(specialChars, ch) {
			return false
		}
	}
	return true
}

// unescapeRegExpPattern converts regex escape sequences to string literals
func unescapeRegExpPattern(pattern string) string {
	result := strings.Builder{}
	i := 0
	for i < len(pattern) {
		if pattern[i] == '\\' && i+1 < len(pattern) {
			next := pattern[i+1]
			switch next {
			case 'n':
				result.WriteString("\\n")
				i += 2
			case 'r':
				result.WriteString("\\r")
				i += 2
			case 't':
				result.WriteString("\\t")
				i += 2
			case 'v':
				result.WriteString("\\v")
				i += 2
			case 'f':
				result.WriteString("\\f")
				i += 2
			case '0':
				result.WriteString("\\0")
				i += 2
			case '\\':
				result.WriteString("\\\\")
				i += 2
			case '\'':
				result.WriteString("\\'")
				i += 2
			default:
				// Keep other escaped characters as-is
				result.WriteByte(pattern[i+1])
				i += 2
			}
		} else {
			result.WriteByte(pattern[i])
			i++
		}
	}
	return result.String()
}

var PreferIncludesRule = rule.CreateRule(rule.Rule{
	Name: "prefer-includes",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {

		// Helper to check if a type has includes method
		hasIncludesMethod := func(node *ast.Node) bool {
			if node == nil || ctx.TypeChecker == nil {
				return false
			}

			nodeType := ctx.TypeChecker.GetTypeAtLocation(node)
			if nodeType == nil {
				return false
			}

			// Check for includes method
			includesSymbol := nodeType.GetProperty("includes")
			if includesSymbol == nil {
				return false
			}

			// Also check that indexOf exists with compatible signature
			indexOfSymbol := nodeType.GetProperty("indexOf")
			if indexOfSymbol == nil {
				return false
			}

			return true
		}

		// Helper to get the receiver and argument from indexOf call
		parseIndexOfCall := func(node *ast.Node) (receiver *ast.Node, argument *ast.Node, ok bool) {
			if !ast.IsCallExpression(node) {
				return nil, nil, false
			}
			callExpr := node.AsCallExpression()
			if callExpr == nil || callExpr.Expression == nil {
				return nil, nil, false
			}

			// Check if it's a property access expression calling 'indexOf'
			if !ast.IsPropertyAccessExpression(callExpr.Expression) {
				return nil, nil, false
			}
			propAccess := callExpr.Expression.AsPropertyAccessExpression()
			if propAccess == nil || propAccess.Name == nil {
				return nil, nil, false
			}

			// Check if the method name is 'indexOf'
			if !ast.IsIdentifier(propAccess.Name) {
				return nil, nil, false
			}
			nameIdent := propAccess.Name.AsIdentifier()
			if nameIdent == nil || nameIdent.EscapedText != "indexOf" {
				return nil, nil, false
			}

			// Check that there's exactly one argument
			if callExpr.Arguments == nil || len(callExpr.Arguments.Slice()) != 1 {
				return nil, nil, false
			}

			receiver = propAccess.Expression
			argument = callExpr.Arguments.Slice()[0]
			return receiver, argument, true
		}

		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				if !ast.IsBinaryExpression(node) {
					return
				}
				binExpr := node.AsBinaryExpression()
				if binExpr == nil {
					return
				}

				// Check for patterns like: indexOf(...) !== -1, indexOf(...) >= 0, etc.
				var indexOfCall *ast.Node
				var negativeOne bool
				var isPositiveCheck bool // true for !== -1, >= 0, etc.

				switch binExpr.OperatorToken.Kind {
				case ast.KindExclamationEqualsEqualsToken, ast.KindExclamationEqualsToken:
					// !== or !=
					if ast.IsCallExpression(binExpr.Left) {
						indexOfCall = binExpr.Left
						if ast.IsPrefixUnaryExpression(binExpr.Right) {
							prefixUnary := binExpr.Right.AsPrefixUnaryExpression()
							if prefixUnary != nil && prefixUnary.Operator == ast.KindMinusToken {
								if ast.IsNumericLiteral(prefixUnary.Operand) {
									numLit := prefixUnary.Operand.AsNumericLiteral()
									if numLit != nil && numLit.Text == "1" {
										negativeOne = true
										isPositiveCheck = true
									}
								}
							}
						}
					} else if ast.IsCallExpression(binExpr.Right) {
						indexOfCall = binExpr.Right
						if ast.IsPrefixUnaryExpression(binExpr.Left) {
							prefixUnary := binExpr.Left.AsPrefixUnaryExpression()
							if prefixUnary != nil && prefixUnary.Operator == ast.KindMinusToken {
								if ast.IsNumericLiteral(prefixUnary.Operand) {
									numLit := prefixUnary.Operand.AsNumericLiteral()
									if numLit != nil && numLit.Text == "1" {
										negativeOne = true
										isPositiveCheck = true
									}
								}
							}
						}
					}

				case ast.KindEqualsEqualsEqualsToken, ast.KindEqualsEqualsToken:
					// === or ==
					if ast.IsCallExpression(binExpr.Left) {
						indexOfCall = binExpr.Left
						if ast.IsPrefixUnaryExpression(binExpr.Right) {
							prefixUnary := binExpr.Right.AsPrefixUnaryExpression()
							if prefixUnary != nil && prefixUnary.Operator == ast.KindMinusToken {
								if ast.IsNumericLiteral(prefixUnary.Operand) {
									numLit := prefixUnary.Operand.AsNumericLiteral()
									if numLit != nil && numLit.Text == "1" {
										negativeOne = true
										isPositiveCheck = false
									}
								}
							}
						}
					} else if ast.IsCallExpression(binExpr.Right) {
						indexOfCall = binExpr.Right
						if ast.IsPrefixUnaryExpression(binExpr.Left) {
							prefixUnary := binExpr.Left.AsPrefixUnaryExpression()
							if prefixUnary != nil && prefixUnary.Operator == ast.KindMinusToken {
								if ast.IsNumericLiteral(prefixUnary.Operand) {
									numLit := prefixUnary.Operand.AsNumericLiteral()
									if numLit != nil && numLit.Text == "1" {
										negativeOne = true
										isPositiveCheck = false
									}
								}
							}
						}
					}

				case ast.KindGreaterThanToken:
					// >
					if ast.IsCallExpression(binExpr.Left) {
						indexOfCall = binExpr.Left
						if ast.IsPrefixUnaryExpression(binExpr.Right) {
							prefixUnary := binExpr.Right.AsPrefixUnaryExpression()
							if prefixUnary != nil && prefixUnary.Operator == ast.KindMinusToken {
								if ast.IsNumericLiteral(prefixUnary.Operand) {
									numLit := prefixUnary.Operand.AsNumericLiteral()
									if numLit != nil && numLit.Text == "1" {
										negativeOne = true
										isPositiveCheck = true
									}
								}
							}
						}
					} else if ast.IsCallExpression(binExpr.Right) {
						indexOfCall = binExpr.Right
						if ast.IsPrefixUnaryExpression(binExpr.Left) {
							prefixUnary := binExpr.Left.AsPrefixUnaryExpression()
							if prefixUnary != nil && prefixUnary.Operator == ast.KindMinusToken {
								if ast.IsNumericLiteral(prefixUnary.Operand) {
									numLit := prefixUnary.Operand.AsNumericLiteral()
									if numLit != nil && numLit.Text == "1" {
										negativeOne = true
										isPositiveCheck = false // -1 > indexOf means indexOf < -1
									}
								}
							}
						}
					}

				case ast.KindLessThanToken:
					// <
					if ast.IsCallExpression(binExpr.Left) {
						indexOfCall = binExpr.Left
						if ast.IsNumericLiteral(binExpr.Right) {
							numLit := binExpr.Right.AsNumericLiteral()
							if numLit != nil && numLit.Text == "0" {
								negativeOne = true
								isPositiveCheck = false
							}
						}
					} else if ast.IsCallExpression(binExpr.Right) {
						indexOfCall = binExpr.Right
						if ast.IsNumericLiteral(binExpr.Left) {
							numLit := binExpr.Left.AsNumericLiteral()
							if numLit != nil && numLit.Text == "0" {
								negativeOne = true
								isPositiveCheck = true // 0 < indexOf
							}
						}
					}

				case ast.KindGreaterThanEqualsToken:
					// >=
					if ast.IsCallExpression(binExpr.Left) {
						indexOfCall = binExpr.Left
						if ast.IsNumericLiteral(binExpr.Right) {
							numLit := binExpr.Right.AsNumericLiteral()
							if numLit != nil && numLit.Text == "0" {
								negativeOne = true
								isPositiveCheck = true
							}
						}
					} else if ast.IsCallExpression(binExpr.Right) {
						indexOfCall = binExpr.Right
						if ast.IsNumericLiteral(binExpr.Left) {
							numLit := binExpr.Left.AsNumericLiteral()
							if numLit != nil && numLit.Text == "0" {
								negativeOne = true
								isPositiveCheck = false // 0 >= indexOf
							}
						}
					}

				case ast.KindLessThanEqualsToken:
					// <=
					if ast.IsCallExpression(binExpr.Left) {
						indexOfCall = binExpr.Left
						if ast.IsPrefixUnaryExpression(binExpr.Right) {
							prefixUnary := binExpr.Right.AsPrefixUnaryExpression()
							if prefixUnary != nil && prefixUnary.Operator == ast.KindMinusToken {
								if ast.IsNumericLiteral(prefixUnary.Operand) {
									numLit := prefixUnary.Operand.AsNumericLiteral()
									if numLit != nil && numLit.Text == "1" {
										negativeOne = true
										isPositiveCheck = false
									}
								}
							}
						}
					} else if ast.IsCallExpression(binExpr.Right) {
						indexOfCall = binExpr.Right
						if ast.IsPrefixUnaryExpression(binExpr.Left) {
							prefixUnary := binExpr.Left.AsPrefixUnaryExpression()
							if prefixUnary != nil && prefixUnary.Operator == ast.KindMinusToken {
								if ast.IsNumericLiteral(prefixUnary.Operand) {
									numLit := prefixUnary.Operand.AsNumericLiteral()
									if numLit != nil && numLit.Text == "1" {
										negativeOne = true
										isPositiveCheck = true // -1 <= indexOf
									}
								}
							}
						}
					}
				}

				if indexOfCall == nil || !negativeOne {
					return
				}

				// Parse the indexOf call
				receiver, argument, ok := parseIndexOfCall(indexOfCall)
				if !ok {
					return
				}

				// Check if the receiver has includes method
				if !hasIncludesMethod(receiver) {
					return
				}

				// Check for optional chaining
				if hasOptionalChain(indexOfCall) {
					// Report without fix
					ctx.ReportNode(node, buildPreferIncludesMessage())
					return
				}

				// Build the fix
				receiverText := getTextForNode(ctx.SourceFile, receiver)
				argumentText := getTextForNode(ctx.SourceFile, argument)

				var replacement string
				if isPositiveCheck {
					replacement = receiverText + ".includes(" + argumentText + ")"
				} else {
					replacement = "!" + receiverText + ".includes(" + argumentText + ")"
				}

				ctx.ReportNodeWithFixes(node, buildPreferIncludesMessage(),
					rule.RuleFixReplace(ctx.SourceFile, node, replacement))
			},

			ast.KindCallExpression: func(node *ast.Node) {
				if !ast.IsCallExpression(node) {
					return
				}
				callExpr := node.AsCallExpression()
				if callExpr == nil || callExpr.Expression == nil {
					return
				}

				// Check for RegExp.test() pattern
				if !ast.IsPropertyAccessExpression(callExpr.Expression) {
					return
				}
				propAccess := callExpr.Expression.AsPropertyAccessExpression()
				if propAccess == nil || propAccess.Name == nil {
					return
				}

				// Check if method is 'test'
				if !ast.IsIdentifier(propAccess.Name) {
					return
				}
				nameIdent := propAccess.Name.AsIdentifier()
				if nameIdent == nil || nameIdent.EscapedText != "test" {
					return
				}

				// The receiver should be a RegExp
				regexpNode := propAccess.Expression

				// Try to parse the regexp pattern
				pattern, flags, ok := parseRegExpPattern(regexpNode, ctx)
				if !ok {
					return
				}

				// Check if it's a simple pattern
				if !isSimpleRegExpPattern(pattern, flags) {
					return
				}

				// Get the argument to test()
				if callExpr.Arguments == nil || len(callExpr.Arguments.Slice()) != 1 {
					return
				}
				testArg := callExpr.Arguments.Slice()[0]

				// Build the fix
				testArgText := getTextForNode(ctx.SourceFile, testArg)

				// Handle parentheses for complex expressions
				needsParens := false
				if ast.IsBinaryExpression(testArg) || ast.IsConditionalExpression(testArg) {
					needsParens = true
				}
				// Check for comma/sequence expressions
				if ast.IsCommaListExpression(testArg) {
					needsParens = true
				}

				if needsParens && !strings.HasPrefix(testArgText, "(") {
					testArgText = "(" + testArgText + ")"
				}

				// Convert pattern to string literal
				unescapedPattern := unescapeRegExpPattern(pattern)
				stringLiteral := strconv.Quote(unescapedPattern)

				replacement := testArgText + ".includes(" + stringLiteral + ")"

				ctx.ReportNodeWithFixes(node, buildPreferStringIncludesMessage(),
					rule.RuleFixReplace(ctx.SourceFile, node, replacement))
			},
		}
	},
})
