package no_unsafe_optional_chaining

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var unsafeArithmeticOperators = map[ast.SyntaxKind]bool{
	ast.SyntaxKindPlusToken:         true,
	ast.SyntaxKindMinusToken:        true,
	ast.SyntaxKindSlashToken:        true,
	ast.SyntaxKindAsteriskToken:     true,
	ast.SyntaxKindPercentToken:      true,
	ast.SyntaxKindAsteriskAsteriskToken: true,
}

var unsafeAssignmentOperators = map[ast.SyntaxKind]bool{
	ast.SyntaxKindPlusEqualsToken:         true,
	ast.SyntaxKindMinusEqualsToken:        true,
	ast.SyntaxKindSlashEqualsToken:        true,
	ast.SyntaxKindAsteriskEqualsToken:     true,
	ast.SyntaxKindPercentEqualsToken:      true,
	ast.SyntaxKindAsteriskAsteriskEqualsToken: true,
}

var unsafeRelationalOperators = map[ast.SyntaxKind]bool{
	ast.SyntaxKindInKeyword:         true,
	ast.SyntaxKindInstanceOfKeyword: true,
}

// NoUnsafeOptionalChainingOptions defines the configuration options for this rule
type NoUnsafeOptionalChainingOptions struct {
	DisallowArithmeticOperators bool `json:"disallowArithmeticOperators"`
}

// parseOptions parses and validates the rule options
func parseOptions(options any) NoUnsafeOptionalChainingOptions {
	opts := NoUnsafeOptionalChainingOptions{
		DisallowArithmeticOperators: false,
	}

	if options == nil {
		return opts
	}

	// Handle both array format [{ option: value }] and object format { option: value }
	var optsMap map[string]interface{}
	if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
		optsMap, _ = optArray[0].(map[string]interface{})
	} else {
		optsMap, _ = options.(map[string]interface{})
	}

	if optsMap != nil {
		if v, ok := optsMap["disallowArithmeticOperators"].(bool); ok {
			opts.DisallowArithmeticOperators = v
		}
	}

	return opts
}

// isDestructuringPattern checks if a node is a destructuring pattern
func isDestructuringPattern(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return node.Kind == ast.KindObjectBindingPattern || node.Kind == ast.KindArrayBindingPattern
}

// checkUndefinedShortCircuit recursively checks if a node can short-circuit with undefined
func checkUndefinedShortCircuit(ctx rule.RuleContext, node *ast.Node, reportFunc func(*ast.Node)) {
	if node == nil {
		return
	}

	switch node.Kind {
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary == nil {
			return
		}
		op := binary.OperatorToken.Kind
		// For || and ??, only check right side
		if op == ast.SyntaxKindBarBarToken || op == ast.SyntaxKindQuestionQuestionToken {
			checkUndefinedShortCircuit(ctx, binary.Right, reportFunc)
		} else if op == ast.SyntaxKindAmpersandAmpersandToken {
			// For &&, check both sides
			checkUndefinedShortCircuit(ctx, binary.Left, reportFunc)
			checkUndefinedShortCircuit(ctx, binary.Right, reportFunc)
		}

	case ast.KindCommaListExpression:
		// SequenceExpression - check last expression
		comma := node.AsCommaListExpression()
		if comma != nil && comma.Elements != nil && len(comma.Elements.Nodes) > 0 {
			lastExpr := comma.Elements.Nodes[len(comma.Elements.Nodes)-1]
			checkUndefinedShortCircuit(ctx, lastExpr, reportFunc)
		}

	case ast.KindConditionalExpression:
		cond := node.AsConditionalExpression()
		if cond != nil {
			checkUndefinedShortCircuit(ctx, cond.WhenTrue, reportFunc)
			checkUndefinedShortCircuit(ctx, cond.WhenFalse, reportFunc)
		}

	case ast.KindAwaitExpression:
		await := node.AsAwaitExpression()
		if await != nil {
			checkUndefinedShortCircuit(ctx, await.Expression, reportFunc)
		}

	case ast.KindNonNullExpression:
		// ChainExpression equivalent in TypeScript
		reportFunc(node)

	case ast.KindPropertyAccessExpression:
		// Check if this is optional chaining
		propAccess := node.AsPropertyAccessExpression()
		if propAccess != nil && propAccess.QuestionDotToken != nil {
			reportFunc(node)
		}

	case ast.KindElementAccessExpression:
		// Check if this is optional chaining
		elemAccess := node.AsElementAccessExpression()
		if elemAccess != nil && elemAccess.QuestionDotToken != nil {
			reportFunc(node)
		}

	case ast.KindCallExpression:
		// Check if this is optional chaining
		call := node.AsCallExpression()
		if call != nil && call.QuestionDotToken != nil {
			reportFunc(node)
		}
	}
}

// NoUnsafeOptionalChainingRule implements the no-unsafe-optional-chaining rule
// Disallow use of optional chaining in contexts where the `undefined` value is not allowed
var NoUnsafeOptionalChainingRule = rule.Rule{
	Name: "no-unsafe-optional-chaining",
	Run:  run,
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := parseOptions(options)

	reportUnsafeUsage := func(node *ast.Node) {
		ctx.ReportNode(node, rule.RuleMessage{
			Id:          "unsafeOptionalChain",
			Description: "Unsafe usage of optional chaining. If it short-circuits with 'undefined' the evaluation will throw TypeError.",
		})
	}

	reportUnsafeArithmetic := func(node *ast.Node) {
		ctx.ReportNode(node, rule.RuleMessage{
			Id:          "unsafeArithmetic",
			Description: "Unsafe arithmetic operation on optional chaining. It can result in NaN.",
		})
	}

	checkUnsafeUsage := func(node *ast.Node) {
		checkUndefinedShortCircuit(ctx, node, reportUnsafeUsage)
	}

	checkUnsafeArithmetic := func(node *ast.Node) {
		checkUndefinedShortCircuit(ctx, node, reportUnsafeArithmetic)
	}

	return rule.RuleListeners{
		// Assignment with destructuring
		ast.KindBinaryExpression: func(node *ast.Node) {
			binary := node.AsBinaryExpression()
			if binary == nil {
				return
			}

			op := binary.OperatorToken.Kind

			// Check for assignment with destructuring
			if op == ast.SyntaxKindEqualsToken && isDestructuringPattern(binary.Left) {
				checkUnsafeUsage(binary.Right)
				return
			}

			// Check for unsafe relational operators
			if unsafeRelationalOperators[op] {
				checkUnsafeUsage(binary.Right)
			}

			// Check for arithmetic operators if option enabled
			if opts.DisallowArithmeticOperators && unsafeArithmeticOperators[op] {
				checkUnsafeArithmetic(binary.Right)
				checkUnsafeArithmetic(binary.Left)
			}

			// Check for arithmetic assignment operators if option enabled
			if opts.DisallowArithmeticOperators && unsafeAssignmentOperators[op] {
				checkUnsafeArithmetic(binary.Right)
			}
		},

		// Class extends
		ast.KindClassDeclaration: func(node *ast.Node) {
			classDecl := node.AsClassDeclaration()
			if classDecl != nil && len(classDecl.HeritageClauses) > 0 {
				for _, heritage := range classDecl.HeritageClauses {
					if heritage != nil && heritage.Types != nil {
						for _, typeExpr := range heritage.Types.Nodes {
							if typeExpr != nil && typeExpr.Expression != nil {
								checkUnsafeUsage(typeExpr.Expression)
							}
						}
					}
				}
			}
		},

		ast.KindClassExpression: func(node *ast.Node) {
			classExpr := node.AsClassExpression()
			if classExpr != nil && len(classExpr.HeritageClauses) > 0 {
				for _, heritage := range classExpr.HeritageClauses {
					if heritage != nil && heritage.Types != nil {
						for _, typeExpr := range heritage.Types.Nodes {
							if typeExpr != nil && typeExpr.Expression != nil {
								checkUnsafeUsage(typeExpr.Expression)
							}
						}
					}
				}
			}
		},

		// Function calls (non-optional)
		ast.KindCallExpression: func(node *ast.Node) {
			call := node.AsCallExpression()
			if call != nil && call.QuestionDotToken == nil && call.Expression != nil {
				checkUnsafeUsage(call.Expression)
			}
		},

		// Constructor calls
		ast.KindNewExpression: func(node *ast.Node) {
			newExpr := node.AsNewExpression()
			if newExpr != nil && newExpr.Expression != nil {
				checkUnsafeUsage(newExpr.Expression)
			}
		},

		// Variable destructuring
		ast.KindVariableDeclaration: func(node *ast.Node) {
			varDecl := node.AsVariableDeclaration()
			if varDecl != nil && isDestructuringPattern(varDecl.Name) && varDecl.Initializer != nil {
				checkUnsafeUsage(varDecl.Initializer)
			}
		},

		// Property access (non-optional)
		ast.KindPropertyAccessExpression: func(node *ast.Node) {
			propAccess := node.AsPropertyAccessExpression()
			if propAccess != nil && propAccess.QuestionDotToken == nil && propAccess.Expression != nil {
				checkUnsafeUsage(propAccess.Expression)
			}
		},

		// Element access (non-optional)
		ast.KindElementAccessExpression: func(node *ast.Node) {
			elemAccess := node.AsElementAccessExpression()
			if elemAccess != nil && elemAccess.QuestionDotToken == nil && elemAccess.Expression != nil {
				checkUnsafeUsage(elemAccess.Expression)
			}
		},

		// Tagged template
		ast.KindTaggedTemplateExpression: func(node *ast.Node) {
			tagged := node.AsTaggedTemplateExpression()
			if tagged != nil && tagged.Tag != nil {
				checkUnsafeUsage(tagged.Tag)
			}
		},

		// For-of statement
		ast.KindForOfStatement: func(node *ast.Node) {
			forOf := node.AsForOfStatement()
			if forOf != nil && forOf.Expression != nil {
				checkUnsafeUsage(forOf.Expression)
			}
		},

		// Spread element
		ast.KindSpreadElement: func(node *ast.Node) {
			spread := node.AsSpreadElement()
			if spread == nil || spread.Expression == nil {
				return
			}
			// Don't check if parent is ObjectExpression (object spread is safe)
			parent := node.Parent
			if parent != nil && parent.Kind != ast.KindObjectLiteralExpression {
				checkUnsafeUsage(spread.Expression)
			}
		},

		// With statement
		ast.KindWithStatement: func(node *ast.Node) {
			withStmt := node.AsWithStatement()
			if withStmt != nil && withStmt.Expression != nil {
				checkUnsafeUsage(withStmt.Expression)
			}
		},

		// Unary expression (arithmetic)
		ast.KindPrefixUnaryExpression: func(node *ast.Node) {
			if !opts.DisallowArithmeticOperators {
				return
			}
			prefix := node.AsPrefixUnaryExpression()
			if prefix != nil && unsafeArithmeticOperators[prefix.Operator] && prefix.Operand != nil {
				checkUnsafeArithmetic(prefix.Operand)
			}
		},

		ast.KindPostfixUnaryExpression: func(node *ast.Node) {
			if !opts.DisallowArithmeticOperators {
				return
			}
			postfix := node.AsPostfixUnaryExpression()
			if postfix != nil && unsafeArithmeticOperators[postfix.Operator] && postfix.Operand != nil {
				checkUnsafeArithmetic(postfix.Operand)
			}
		},
	}
}
