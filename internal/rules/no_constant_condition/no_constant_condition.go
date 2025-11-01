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

// getBooleanValue returns the boolean value of a literal
func getBooleanValue(node *ast.Node) *bool {
	if node == nil {
		return nil
	}

	switch node.Kind {
	case ast.KindTrueKeyword:
		t := true
		return &t
	case ast.KindFalseKeyword:
		f := false
		return &f
	case ast.KindNullKeyword:
		f := false
		return &f
	case ast.KindNumericLiteral:
		// 0 is falsy, other numbers are truthy
		text := node.Text()
		if text == "0" || text == "0.0" || text == "-0" {
			f := false
			return &f
		}
		t := true
		return &t
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		// Empty string is falsy
		text := node.Text()
		// Remove quotes for string literals
		if node.Kind == ast.KindStringLiteral && len(text) >= 2 {
			text = text[1 : len(text)-1]
		}
		// Remove backticks for template literals
		if node.Kind == ast.KindNoSubstitutionTemplateLiteral && len(text) >= 2 {
			text = text[1 : len(text)-1]
		}
		if len(text) == 0 {
			f := false
			return &f
		}
		t := true
		return &t
	case ast.KindIdentifier:
		// undefined is falsy
		if node.Text() == "undefined" {
			f := false
			return &f
		}
	}
	return nil
}

// isLogicalIdentity checks if a node is a logical identity element
func isLogicalIdentity(node *ast.Node, operator ast.Kind) bool {
	if node == nil {
		return false
	}

	// Check literals
	boolVal := getBooleanValue(node)
	if boolVal != nil {
		if operator == ast.KindBarBarToken && *boolVal == true {
			return true
		}
		if operator == ast.KindAmpersandAmpersandToken && *boolVal == false {
			return true
		}
	}

	// void operator is identity for &&
	if node.Kind == ast.KindPrefixUnaryExpression {
		prefix := node.AsPrefixUnaryExpression()
		if prefix != nil && prefix.Operator == ast.KindVoidKeyword {
			return operator == ast.KindAmpersandAmpersandToken
		}
	}

	// Logical expressions with same operator
	if node.Kind == ast.KindBinaryExpression {
		binary := node.AsBinaryExpression()
		if binary != nil && binary.OperatorToken != nil {
			nodeOp := binary.OperatorToken.Kind
			if nodeOp == operator && (nodeOp == ast.KindBarBarToken || nodeOp == ast.KindAmpersandAmpersandToken) {
				return isLogicalIdentity(binary.Left, operator) || isLogicalIdentity(binary.Right, operator)
			}
		}
	}

	// Assignment expressions
	if node.Kind == ast.KindBinaryExpression {
		binary := node.AsBinaryExpression()
		if binary != nil && binary.OperatorToken != nil {
			nodeOp := binary.OperatorToken.Kind
			if nodeOp == ast.KindBarBarEqualsToken || nodeOp == ast.KindAmpersandAmpersandEqualsToken {
				// Extract the base operator (|| or &&)
				var baseOp ast.Kind
				if nodeOp == ast.KindBarBarEqualsToken {
					baseOp = ast.KindBarBarToken
				} else {
					baseOp = ast.KindAmpersandAmpersandToken
				}
				return operator == baseOp && isLogicalIdentity(binary.Right, operator)
			}
		}
	}

	return false
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
		// Template literals: constant if any static part has length (in boolean position)
		// or all expressions are constant (not in boolean position)
		template := node.AsTemplateExpression()
		if template == nil {
			return false
		}

		// In boolean position: constant if any quasi has content
		if inBooleanPosition {
			// Check template head
			if template.Head != nil && len(template.Head.Text()) > 0 {
				return true
			}
			// Check template spans
			if template.TemplateSpans != nil {
				for _, span := range template.TemplateSpans.Nodes {
					if span.Kind == ast.KindTemplateSpan {
						templateSpan := span.AsTemplateSpan()
						if templateSpan != nil && templateSpan.Literal != nil {
							text := templateSpan.Literal.Text()
							if len(text) > 0 {
								return true
							}
						}
					}
				}
			}
			return false
		}

		// Not in boolean position: constant if all expressions are constant
		if template.TemplateSpans != nil {
			for _, span := range template.TemplateSpans.Nodes {
				if span.Kind == ast.KindTemplateSpan {
					templateSpan := span.AsTemplateSpan()
					if templateSpan != nil && templateSpan.Expression != nil {
						if !isConstant(templateSpan.Expression, false) {
							return false
						}
					}
				}
			}
		}
		return true

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

		// Comma operator: sequence expression, constant if last expression is constant
		if operator == ast.KindCommaToken {
			return isConstant(binary.Right, inBooleanPosition)
		}

		// Assignment expressions
		if operator == ast.KindEqualsToken {
			// Simple assignment: constant if right side is constant
			return isConstant(binary.Right, inBooleanPosition)
		}

		// Logical assignment operators (||=, &&=)
		if operator == ast.KindBarBarEqualsToken || operator == ast.KindAmpersandAmpersandEqualsToken {
			if inBooleanPosition {
				var baseOp ast.Kind
				if operator == ast.KindBarBarEqualsToken {
					baseOp = ast.KindBarBarToken
				} else {
					baseOp = ast.KindAmpersandAmpersandToken
				}
				return isLogicalIdentity(binary.Right, baseOp)
			}
			return false
		}

		// Logical operators (&&, ||, ??)
		switch operator {
		case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			isLeftConstant := isConstant(binary.Left, inBooleanPosition)
			isRightConstant := isConstant(binary.Right, inBooleanPosition)
			isLeftShortCircuit := isLeftConstant && isLogicalIdentity(binary.Left, operator)
			isRightShortCircuit := inBooleanPosition && isRightConstant && isLogicalIdentity(binary.Right, operator)

			return (isLeftConstant && isRightConstant) || isLeftShortCircuit || isRightShortCircuit
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

// containsYield checks if a node contains a yield expression (not inside nested functions)
func containsYield(node *ast.Node) bool {
	if node == nil {
		return false
	}

	// If this is a yield expression, return true
	if node.Kind == ast.KindYieldExpression {
		return true
	}

	// Don't traverse into nested function bodies
	switch node.Kind {
	case ast.KindFunctionDeclaration,
		ast.KindFunctionExpression,
		ast.KindArrowFunction:
		return false
	}

	// Recursively check children based on node type
	// We need to be careful to avoid calling Children() on unsupported node types
	switch node.Kind {
	case ast.KindBlock:
		block := node.AsBlock()
		if block != nil && block.Statements != nil {
			for _, stmt := range block.Statements.Nodes {
				if containsYield(stmt) {
					return true
				}
			}
		}
	case ast.KindIfStatement:
		ifStmt := node.AsIfStatement()
		if ifStmt != nil {
			if containsYield(ifStmt.Statement) {
				return true
			}
			if ifStmt.ElseStatement != nil && containsYield(ifStmt.ElseStatement) {
				return true
			}
		}
	case ast.KindWhileStatement:
		whileStmt := node.AsWhileStatement()
		if whileStmt != nil && containsYield(whileStmt.Statement) {
			return true
		}
	case ast.KindDoStatement:
		doStmt := node.AsDoStatement()
		if doStmt != nil && containsYield(doStmt.Statement) {
			return true
		}
	case ast.KindForStatement:
		forStmt := node.AsForStatement()
		if forStmt != nil && containsYield(forStmt.Statement) {
			return true
		}
	case ast.KindExpressionStatement:
		exprStmt := node.AsExpressionStatement()
		if exprStmt != nil && exprStmt.Expression != nil {
			if containsYield(exprStmt.Expression) {
				return true
			}
		}
	case ast.KindBinaryExpression,
		ast.KindConditionalExpression,
		ast.KindCallExpression,
		ast.KindPrefixUnaryExpression,
		ast.KindPostfixUnaryExpression:
		// For expressions, try using Children if available
		// But don't panic if it's not supported
		defer func() {
			if r := recover(); r != nil {
				// Ignore panic from unsupported node types
			}
		}()
		children := node.Children()
		if children != nil {
			for _, child := range children.Nodes {
				if containsYield(child) {
					return true
				}
			}
		}
	}

	return false
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

	// Don't check loops in generator functions that contain yield
	// Get the loop body
	var body *ast.Node
	switch node.Kind {
	case ast.KindWhileStatement:
		whileStmt := node.AsWhileStatement()
		if whileStmt != nil {
			body = whileStmt.Statement
		}
	case ast.KindDoStatement:
		doStmt := node.AsDoStatement()
		if doStmt != nil {
			body = doStmt.Statement
		}
	case ast.KindForStatement:
		forStmt := node.AsForStatement()
		if forStmt != nil {
			body = forStmt.Statement
		}
	}

	// If the loop body contains a yield, don't check it
	if body != nil && containsYield(body) {
		return false
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
