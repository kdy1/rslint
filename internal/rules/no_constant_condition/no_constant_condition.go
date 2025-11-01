package no_constant_condition

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Message builder
func buildUnexpectedMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpected",
		Description: "Unexpected constant condition.",
	}
}

// Options structure
type Options struct {
	CheckLoops string // "all", "allExceptWhileTrue", "none"
}

// parseOptions parses the rule options
func parseOptions(options any) Options {
	opts := Options{
		CheckLoops: "allExceptWhileTrue", // default
	}

	if options == nil {
		return opts
	}

	// Handle map[string]interface{}
	if optMap, ok := options.(map[string]interface{}); ok {
		if checkLoops, ok := optMap["checkLoops"].(string); ok {
			opts.CheckLoops = checkLoops
		} else if checkLoopsBool, ok := optMap["checkLoops"].(bool); ok {
			// Handle boolean values: true = "all", false = "none"
			if checkLoopsBool {
				opts.CheckLoops = "all"
			} else {
				opts.CheckLoops = "none"
			}
		}
	}

	return opts
}

// isConstant checks if a node represents a constant value
func isConstant(node *ast.Node, inBooleanPosition bool) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	// Literal values
	case ast.KindNumericLiteral,
		ast.KindStringLiteral,
		ast.KindNoSubstitutionTemplateLiteral,
		ast.KindTrueKeyword,
		ast.KindFalseKeyword,
		ast.KindNullKeyword,
		ast.KindRegularExpressionLiteral,
		ast.KindBigIntLiteral:
		return true

	case ast.KindIdentifier:
		// 'undefined' and 'Infinity' are constants
		text := node.Text()
		return text == "undefined" || text == "Infinity"

	case ast.KindObjectLiteralExpression,
		ast.KindArrayLiteralExpression,
		ast.KindArrowFunction,
		ast.KindFunctionExpression,
		ast.KindClassExpression:
		return true

	case ast.KindTemplateExpression:
		// Template literals with expressions are not constant
		return false

	case ast.KindParenthesizedExpression:
		paren := node.AsParenthesizedExpression()
		if paren != nil && paren.Expression != nil {
			return isConstant(paren.Expression, inBooleanPosition)
		}

	case ast.KindPrefixUnaryExpression:
		prefix := node.AsPrefixUnaryExpression()
		if prefix == nil || prefix.Operand == nil {
			return false
		}

		switch prefix.Operator {
		case ast.KindExclamationToken: // !
			return isConstant(prefix.Operand, true)
		case ast.KindTypeOfKeyword: // typeof
			// typeof always returns a constant string in boolean position
			return inBooleanPosition
		case ast.KindVoidKeyword: // void
			return true
		case ast.KindPlusToken, ast.KindMinusToken, ast.KindTildeToken: // +, -, ~
			return isConstant(prefix.Operand, false)
		}

	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary == nil || binary.OperatorToken == nil {
			return false
		}

		operator := binary.OperatorToken.Kind

		// Logical operators - only constant if left side is constant
		switch operator {
		case ast.KindAmpersandAmpersandToken, // &&
			ast.KindBarBarToken: // ||
			return isConstant(binary.Left, inBooleanPosition)
		}

		// Nullish coalescing operator
		if operator == ast.KindQuestionQuestionToken {
			return isConstant(binary.Left, inBooleanPosition)
		}

		// Comparison operators - both sides must be constant
		switch operator {
		case ast.KindLessThanToken,
			ast.KindLessThanEqualsToken,
			ast.KindGreaterThanToken,
			ast.KindGreaterThanEqualsToken,
			ast.KindEqualsEqualsToken,
			ast.KindExclamationEqualsToken,
			ast.KindEqualsEqualsEqualsToken,
			ast.KindExclamationEqualsEqualsToken,
			ast.KindInKeyword,
			ast.KindInstanceOfKeyword:
			return isConstant(binary.Left, false) && isConstant(binary.Right, false)
		}

		// Arithmetic operators - both sides must be constant
		switch operator {
		case ast.KindPlusToken,
			ast.KindMinusToken,
			ast.KindAsteriskToken,
			ast.KindSlashToken,
			ast.KindPercentToken,
			ast.KindAsteriskAsteriskToken,
			ast.KindLessThanLessThanToken,
			ast.KindGreaterThanGreaterThanToken,
			ast.KindGreaterThanGreaterThanGreaterThanToken,
			ast.KindBarToken,
			ast.KindAmpersandToken,
			ast.KindCaretToken:
			return isConstant(binary.Left, false) && isConstant(binary.Right, false)
		}

	case ast.KindConditionalExpression:
		// Ternary operator - only constant if test is constant
		cond := node.AsConditionalExpression()
		if cond != nil && cond.Condition != nil {
			return isConstant(cond.Condition, false)
		}

	case ast.KindNewExpression:
		// new expressions with certain constructors are constant
		newExpr := node.AsNewExpression()
		if newExpr != nil && newExpr.Expression != nil {
			if newExpr.Expression.Kind == ast.KindIdentifier {
				name := newExpr.Expression.Text()
				// new Boolean(), new String(), new Number() are constant
				if name == "Boolean" || name == "String" || name == "Number" {
					return true
				}
			}
		}
		// Other new expressions create new objects (not constant in terms of identity)
		return true

	case ast.KindCallExpression:
		// Boolean(), String(), Number() with constant arguments are constant
		call := node.AsCallExpression()
		if call != nil && call.Expression != nil && call.Expression.Kind == ast.KindIdentifier {
			name := call.Expression.Text()
			if name == "Boolean" || name == "String" || name == "Number" {
				// With no arguments or constant arguments, these are constant
				if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
					return true
				}
				if len(call.Arguments.Nodes) == 1 {
					return isConstant(call.Arguments.Nodes[0], false)
				}
			}
		}

	case ast.KindPostfixUnaryExpression:
		// ++ and -- are not constant (they modify variables)
		return false
	}

	return false
}

// isWhileTrueLoop checks if a loop is a `while (true)` loop
func isWhileTrueLoop(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindWhileStatement {
		return false
	}

	whileStmt := node.AsWhileStatement()
	if whileStmt == nil || whileStmt.Expression == nil {
		return false
	}

	// Check if the condition is literally `true`
	if whileStmt.Expression.Kind == ast.KindTrueKeyword {
		return true
	}

	// Check for parenthesized true: (true), ((true)), etc.
	expr := whileStmt.Expression
	for expr.Kind == ast.KindParenthesizedExpression {
		paren := expr.AsParenthesizedExpression()
		if paren == nil || paren.Expression == nil {
			return false
		}
		expr = paren.Expression
	}

	return expr.Kind == ast.KindTrueKeyword
}

// shouldCheckLoop determines if a loop should be checked based on options
func shouldCheckLoop(node *ast.Node, opts Options) bool {
	if opts.CheckLoops == "none" {
		return false
	}

	if opts.CheckLoops == "allExceptWhileTrue" {
		// Allow while (true) loops
		if isWhileTrueLoop(node) {
			return false
		}
	}

	return true
}

// getTestExpression returns the condition/test expression for a statement
func getTestExpression(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}

	switch node.Kind {
	case ast.KindIfStatement:
		ifStmt := node.AsIfStatement()
		if ifStmt != nil {
			return ifStmt.Expression
		}
	case ast.KindWhileStatement:
		whileStmt := node.AsWhileStatement()
		if whileStmt != nil {
			return whileStmt.Expression
		}
	case ast.KindDoStatement:
		doStmt := node.AsDoStatement()
		if doStmt != nil {
			return doStmt.Expression
		}
	case ast.KindForStatement:
		forStmt := node.AsForStatement()
		if forStmt != nil {
			return forStmt.Condition
		}
	case ast.KindConditionalExpression:
		condExpr := node.AsConditionalExpression()
		if condExpr != nil {
			return condExpr.Condition
		}
	}

	return nil
}

// checkCondition reports if a condition is constant
func checkCondition(ctx rule.RuleContext, node *ast.Node, opts Options) {
	testExpr := getTestExpression(node)
	if testExpr == nil {
		return
	}

	// For loops, check if we should skip based on options
	switch node.Kind {
	case ast.KindWhileStatement, ast.KindDoStatement, ast.KindForStatement:
		if !shouldCheckLoop(node, opts) {
			return
		}
	}

	// Check if the test expression is constant
	if isConstant(testExpr, true) {
		ctx.ReportNode(testExpr, buildUnexpectedMessage())
	}
}

// NoConstantConditionRule disallows constant expressions in conditions
var NoConstantConditionRule = rule.CreateRule(rule.Rule{
	Name: "no-constant-condition",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			// Check if statements
			ast.KindIfStatement: func(node *ast.Node) {
				checkCondition(ctx, node, opts)
			},

			// Check conditional expressions (ternary)
			ast.KindConditionalExpression: func(node *ast.Node) {
				checkCondition(ctx, node, opts)
			},

			// Check while loops
			ast.KindWhileStatement: func(node *ast.Node) {
				checkCondition(ctx, node, opts)
			},

			// Check do-while loops
			ast.KindDoStatement: func(node *ast.Node) {
				checkCondition(ctx, node, opts)
			},

			// Check for loops
			ast.KindForStatement: func(node *ast.Node) {
				checkCondition(ctx, node, opts)
			},
		}
	},
})
