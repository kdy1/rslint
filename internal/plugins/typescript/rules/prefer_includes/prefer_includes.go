package prefer_includes

import (
	"regexp"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildPreferIncludesMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferIncludes",
		Description: "Use .includes() instead of .indexOf() !== -1.",
	}
}

func buildPreferStringIncludesMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferStringIncludes",
		Description: "Use String.includes() instead of RegExp.test().",
	}
}

// Check if a type has both indexOf and includes methods with compatible signatures
func typeHasIncludesMethod(ctx rule.RuleContext, typ *checker.Type) bool {
	if typ == nil {
		return false
	}

	// Get non-nullable type to handle optional types
	nnType := ctx.TypeChecker.GetNonNullableType(typ)
	if nnType == nil {
		return false
	}

	// Get apparent type for better property resolution
	appType := checker.Checker_getApparentType(ctx.TypeChecker, nnType)
	if appType == nil {
		return false
	}

	// Check if type has indexOf method
	indexOfSym := checker.Checker_getPropertyOfType(ctx.TypeChecker, appType, "indexOf")
	if indexOfSym == nil {
		return false
	}

	// Check if type has includes method
	includesSym := checker.Checker_getPropertyOfType(ctx.TypeChecker, appType, "includes")
	if includesSym == nil {
		return false
	}

	// Verify includes is a method (not a boolean property)
	includesType := ctx.TypeChecker.GetTypeOfSymbol(includesSym)
	if includesType == nil {
		return false
	}

	// Check if includes is a function/method type
	signatures := includesType.GetCallSignatures()
	if len(signatures) == 0 {
		return false
	}

	// Get indexOf and includes signatures to compare parameters
	indexOfType := ctx.TypeChecker.GetTypeOfSymbol(indexOfSym)
	if indexOfType == nil {
		return false
	}

	indexOfSigs := indexOfType.GetCallSignatures()
	if len(indexOfSigs) == 0 {
		return false
	}

	// Compare parameter counts and types
	// indexOf typically has (searchElement, fromIndex?)
	// includes typically has (searchElement, fromIndex?)
	// They should have compatible first parameters
	for _, indexOfSig := range indexOfSigs {
		for _, includesSig := range signatures {
			indexOfParams := indexOfSig.GetParameters()
			includesParams := includesSig.GetParameters()

			// Both should have at least one parameter
			if len(indexOfParams) == 0 || len(includesParams) == 0 {
				continue
			}

			// Check if the first parameter is required in both
			// and if includes doesn't have a required second parameter that indexOf doesn't have
			indexOfMinParams := indexOfSig.GetMinArgumentCount()
			includesMinParams := includesSig.GetMinArgumentCount()

			// indexOf should have 1 or 2 params, includes should have 1 or 2 params
			// and the required params should match (both should require at least 1)
			if indexOfMinParams >= 1 && includesMinParams >= 1 {
				// If includes requires a second parameter but indexOf doesn't, they're incompatible
				if includesMinParams > indexOfMinParams {
					continue
				}
				// Compatible signatures found
				return true
			}
		}
	}

	return false
}

// Extract string content from a simple regex literal
func extractSimpleRegexPattern(node *ast.Node) (string, bool) {
	if node == nil || node.Kind != ast.KindRegularExpressionLiteral {
		return "", false
	}

	text := node.Text()
	if text == "" {
		return "", false
	}

	// Parse /pattern/flags format
	if !strings.HasPrefix(text, "/") {
		return "", false
	}

	lastSlash := strings.LastIndex(text, "/")
	if lastSlash <= 0 {
		return "", false
	}

	pattern := text[1:lastSlash]
	flags := text[lastSlash+1:]

	// Only allow simple string patterns without regex special chars (except escaped ones)
	// Reject if there are flags (case insensitive, etc.) or special regex syntax
	if flags != "" {
		return "", false
	}

	// Check for regex metacharacters that make it non-trivial
	// Allow only literal characters and common escape sequences
	if containsRegexMetachars(pattern) {
		return "", false
	}

	// Unescape the pattern for use as a string literal
	unescaped := unescapeRegexPattern(pattern)
	return unescaped, true
}

// Check if pattern contains regex metacharacters (excluding basic escapes)
func containsRegexMetachars(pattern string) bool {
	// Regex metacharacters: . * + ? ^ $ { } [ ] ( ) | \
	// We reject patterns with character classes [], alternation |, quantifiers, etc.
	metacharPattern := regexp.MustCompile(`[.*+?^${}[\]()|]`)
	return metacharPattern.MatchString(pattern)
}

// Unescape regex pattern to string literal format
func unescapeRegexPattern(pattern string) string {
	// This handles common escape sequences like \n, \t, \0, etc.
	// For the purposes of String.includes, we need to convert regex escapes to string escapes
	result := strings.Builder{}
	i := 0
	for i < len(pattern) {
		if pattern[i] == '\\' && i+1 < len(pattern) {
			next := pattern[i+1]
			// Keep the escape sequence as-is for string literal
			result.WriteByte('\\')
			result.WriteByte(next)
			i += 2
		} else {
			result.WriteByte(pattern[i])
			i++
		}
	}
	return result.String()
}

// Extract string from new RegExp('pattern') call
func extractNewRegExpPattern(ctx rule.RuleContext, node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}

	// Look for the variable declaration with new RegExp('pattern')
	// We need to trace back from the identifier to its declaration
	if node.Kind != ast.KindIdentifier {
		return "", false
	}

	sym := ctx.TypeChecker.GetSymbolAtLocation(node)
	if sym == nil || len(sym.Declarations) == 0 {
		return "", false
	}

	for _, decl := range sym.Declarations {
		if decl == nil || decl.Kind != ast.KindVariableDeclaration {
			continue
		}

		varDecl := decl.AsVariableDeclaration()
		if varDecl == nil || varDecl.Initializer == nil {
			continue
		}

		init := varDecl.Initializer
		if init.Kind != ast.KindNewExpression {
			continue
		}

		newExpr := init.AsNewExpression()
		if newExpr == nil || newExpr.Expression == nil {
			continue
		}

		if newExpr.Expression.Kind != ast.KindIdentifier || newExpr.Expression.Text() != "RegExp" {
			continue
		}

		args := newExpr.Arguments()
		if len(args) == 0 {
			continue
		}

		firstArg := args[0]
		if firstArg == nil {
			continue
		}

		// Extract string literal from first argument
		if firstArg.Kind == ast.KindStringLiteral {
			rng := utils.TrimNodeTextRange(ctx.SourceFile, firstArg)
			text := ctx.SourceFile.Text()[rng.Pos():rng.End()]
			if len(text) >= 2 && (text[0] == '\'' || text[0] == '"') {
				return text[1 : len(text)-1], true
			}
		}
	}

	return "", false
}

var PreferIncludesRule = rule.CreateRule(rule.Rule{
	Name: "prefer-includes",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			// Check for a.indexOf(b) !== -1 and similar patterns
			ast.KindBinaryExpression: func(node *ast.Node) {
				binExpr := node.AsBinaryExpression()
				if binExpr == nil {
					return
				}

				// Check for comparison operators: !==, !=, ===, ==, >, >=, <, <=
				var isNegativeCheck bool
				var canFix bool = true

				switch binExpr.OperatorToken.Kind {
				case ast.KindExclamationEqualsEqualsToken, ast.KindExclamationEqualsToken:
					// a.indexOf(b) !== -1 or != -1
					isNegativeCheck = false
				case ast.KindEqualsEqualsEqualsToken, ast.KindEqualsEqualsToken:
					// a.indexOf(b) === -1 or == -1
					isNegativeCheck = true
				case ast.KindGreaterThanToken:
					// a.indexOf(b) > -1
					isNegativeCheck = false
				case ast.KindGreaterThanEqualsToken:
					// a.indexOf(b) >= 0
					isNegativeCheck = false
				case ast.KindLessThanToken:
					// a.indexOf(b) < 0
					isNegativeCheck = true
				case ast.KindLessThanEqualsToken:
					// a.indexOf(b) <= -1
					isNegativeCheck = true
				default:
					return
				}

				var indexOfCall *ast.Node
				var compareValue *ast.Node

				// Check if left side is indexOf call
				if binExpr.Left != nil && binExpr.Left.Kind == ast.KindCallExpression {
					indexOfCall = binExpr.Left
					compareValue = binExpr.Right
				} else if binExpr.Right != nil && binExpr.Right.Kind == ast.KindCallExpression {
					indexOfCall = binExpr.Right
					compareValue = binExpr.Left
				} else {
					return
				}

				if indexOfCall == nil || compareValue == nil {
					return
				}

				callExpr := indexOfCall.AsCallExpression()
				if callExpr == nil || callExpr.Expression == nil {
					return
				}

				// Check if it's a property access for 'indexOf'
				if callExpr.Expression.Kind != ast.KindPropertyAccessExpression {
					return
				}

				propAccess := callExpr.Expression.AsPropertyAccessExpression()
				if propAccess == nil || propAccess.Name() == nil {
					return
				}

				if propAccess.Name().Text() != "indexOf" {
					return
				}

				// Check the compare value is -1 or 0
				var expectedValue int
				if binExpr.OperatorToken.Kind == ast.KindGreaterThanEqualsToken || binExpr.OperatorToken.Kind == ast.KindLessThanToken {
					expectedValue = 0
				} else {
					expectedValue = -1
				}

				// Verify the compare value
				if compareValue.Kind == ast.KindPrefixUnaryExpression {
					unary := compareValue.AsPrefixUnaryExpression()
					if unary != nil && unary.Operator == ast.KindMinusToken && unary.Operand != nil {
						if unary.Operand.Kind == ast.KindNumericLiteral && unary.Operand.Text() == "1" {
							// This is -1
							if expectedValue != -1 {
								return
							}
						} else {
							return
						}
					} else {
						return
					}
				} else if compareValue.Kind == ast.KindNumericLiteral {
					if compareValue.Text() != "0" && compareValue.Text() != "-1" {
						return
					}
					if compareValue.Text() == "0" && expectedValue != 0 {
						return
					}
					if compareValue.Text() == "-1" && expectedValue != -1 {
						return
					}
				} else {
					return
				}

				// Check if the type has includes method
				if propAccess.Expression == nil {
					return
				}

				objType := ctx.TypeChecker.GetTypeAtLocation(propAccess.Expression)
				if !typeHasIncludesMethod(ctx, objType) {
					return
				}

				// Check for optional chaining - we can't fix those
				if propAccess.QuestionDotToken != nil {
					canFix = false
				}

				// Build the fix
				if canFix {
					objRange := utils.TrimNodeTextRange(ctx.SourceFile, propAccess.Expression)
					objText := ctx.SourceFile.Text()[objRange.Pos():objRange.End()]

					// Get the arguments from indexOf call
					args := callExpr.Arguments()
					var argsText string
					if len(args) > 0 {
						argRange := utils.TrimNodeTextRange(ctx.SourceFile, args[0])
						argsText = ctx.SourceFile.Text()[argRange.Pos():argRange.End()]
					}

					var replacement string
					if isNegativeCheck {
						replacement = "!" + objText + ".includes(" + argsText + ")"
					} else {
						replacement = objText + ".includes(" + argsText + ")"
					}

					ctx.ReportNodeWithFixes(node, buildPreferIncludesMessage(),
						rule.RuleFixReplace(ctx.SourceFile, node, replacement))
				} else {
					ctx.ReportNode(node, buildPreferIncludesMessage())
				}
			},

			// Check for /pattern/.test(str) -> str.includes('pattern')
			ast.KindCallExpression: func(node *ast.Node) {
				callExpr := node.AsCallExpression()
				if callExpr == nil || callExpr.Expression == nil {
					return
				}

				// Check if it's a property access for 'test'
				if callExpr.Expression.Kind != ast.KindPropertyAccessExpression {
					return
				}

				propAccess := callExpr.Expression.AsPropertyAccessExpression()
				if propAccess == nil || propAccess.Name() == nil {
					return
				}

				if propAccess.Name().Text() != "test" {
					return
				}

				// Check if the object is a regex literal or RegExp constructor
				if propAccess.Expression == nil {
					return
				}

				var pattern string
				var ok bool

				if propAccess.Expression.Kind == ast.KindRegularExpressionLiteral {
					pattern, ok = extractSimpleRegexPattern(propAccess.Expression)
				} else if propAccess.Expression.Kind == ast.KindIdentifier {
					pattern, ok = extractNewRegExpPattern(ctx, propAccess.Expression)
				} else {
					return
				}

				if !ok {
					return
				}

				// Check if test has an argument
				args := callExpr.Arguments()
				if len(args) == 0 {
					return
				}

				strArg := args[0]
				if strArg == nil {
					return
				}

				// Unwrap sequence expressions (comma operator)
				actualArg := strArg
				for actualArg.Kind == ast.KindCommaListExpression {
					commaList := actualArg.AsCommaListExpression()
					if commaList == nil || len(commaList.Elements.Nodes) == 0 {
						return
					}
					actualArg = commaList.Elements.Nodes[len(commaList.Elements.Nodes)-1]
				}

				// Get the type of the argument to ensure it's a string
				argType := ctx.TypeChecker.GetTypeAtLocation(actualArg)
				if argType != nil {
					// Check if it's a string type
					if !utils.IsStringType(argType, ctx.TypeChecker) {
						return
					}
				}

				// Build the fix
				argRange := utils.TrimNodeTextRange(ctx.SourceFile, strArg)
				argText := ctx.SourceFile.Text()[argRange.Pos():argRange.End()]

				// Wrap in parens if it's a binary expression or similar
				if strArg.Kind == ast.KindBinaryExpression {
					argText = "(" + argText + ")"
				}

				replacement := argText + ".includes('" + pattern + "')"

				ctx.ReportNodeWithFixes(node, buildPreferStringIncludesMessage(),
					rule.RuleFixReplace(ctx.SourceFile, node, replacement))
			},
		}
	},
})
