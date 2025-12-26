package prefer_for_of

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildPreferForOfMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferForOf",
		Description: "Expected a `for-of` loop instead of a `for` loop with this simple iteration.",
	}
}

var PreferForOfRule = rule.CreateRule(rule.Rule{
	Name: "prefer-for-of",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			rule.ListenerOnExit(ast.KindForStatement): func(node *ast.Node) {
				forStmt := node.AsForStatement()
				if forStmt == nil {
					return
				}

				// Check for single variable declaration
				if !isSingleVariableDeclaration(forStmt.Initializer) {
					return
				}

				varDecl := forStmt.Initializer.AsVariableDeclarationList()
				if len(varDecl.Declarations.Nodes) != 1 {
					return
				}

				declarator := varDecl.Declarations.Nodes[0]
				if declarator == nil {
					return
				}

				// Check for zero initialization
				init := declarator.Initializer()
				if !isZeroInitialized(init) {
					return
				}

				// Check that the declarator ID is an identifier
				declaratorName := declarator.Name()
				if !ast.IsIdentifier(declaratorName) {
					return
				}

				indexName := declaratorName.AsIdentifier().Text

				// Check for `i < array.length` pattern
				arrayExpression := isLessThanLengthExpression(forStmt.Condition, indexName)
				if arrayExpression == nil {
					return
				}

				// Check for increment pattern (i++, ++i, i+=1, i=i+1, i=1+i)
				if !isIncrement(forStmt.Incrementor, indexName) {
					return
				}

				// Check that index is only used with array access
				if !isIndexOnlyUsedWithArray(ctx, forStmt.Statement, indexName, arrayExpression) {
					return
				}

				// Report the violation
				headLoc := utils.GetForStatementHeadLoc(ctx.SourceFile, node)
				ctx.ReportRange(headLoc, buildPreferForOfMessage())
			},
		}
	},
})

// isSingleVariableDeclaration checks if node is a single variable declaration (not const)
func isSingleVariableDeclaration(node *ast.Node) bool {
	if node == nil || !ast.IsVariableDeclarationList(node) {
		return false
	}
	varDecl := node.AsVariableDeclarationList()
	// Check that it's not const and has exactly one declaration
	return varDecl.Flags&ast.NodeFlagsConst == 0 && len(varDecl.Declarations.Nodes) == 1
}

// isZeroInitialized checks if the variable is initialized to 0
func isZeroInitialized(initializer *ast.Node) bool {
	if initializer == nil {
		return false
	}
	return isLiteral(initializer, 0)
}

// isLiteral checks if node is a numeric literal with the given value
func isLiteral(node *ast.Node, value int) bool {
	if !ast.IsNumericLiteral(node) {
		return false
	}
	// Check the text representation
	return node.Text() == "0"
}

// isMatchingIdentifier checks if node is an identifier with the given name
func isMatchingIdentifier(node *ast.Node, name string) bool {
	if !ast.IsIdentifier(node) {
		return false
	}
	return node.AsIdentifier().Text == name
}

// isLessThanLengthExpression checks for `i < array.length` pattern
func isLessThanLengthExpression(node *ast.Node, name string) *ast.Node {
	if node == nil || !ast.IsBinaryExpression(node) {
		return nil
	}

	binExpr := node.AsBinaryExpression()
	if binExpr.OperatorToken.Kind != ast.KindLessThanToken {
		return nil
	}

	// Check left side is the index variable
	if !isMatchingIdentifier(binExpr.Left, name) {
		return nil
	}

	// Check right side is `something.length`
	if !ast.IsPropertyAccessExpression(binExpr.Right) {
		return nil
	}

	propAccess := binExpr.Right.AsPropertyAccessExpression()
	propName := propAccess.Name()
	if !ast.IsIdentifier(propName) {
		return nil
	}

	if propName.AsIdentifier().Text != "length" {
		return nil
	}

	return propAccess.Expression
}

// isIncrement checks for i++, ++i, i+=1, i=i+1, or i=1+i patterns
func isIncrement(node *ast.Node, name string) bool {
	if node == nil {
		return false
	}

	// Check for x++ or ++x
	if node.Kind == ast.KindPostfixUnaryExpression || ast.IsPrefixUnaryExpression(node) {
		var operand *ast.Node
		var operator ast.Kind

		if node.Kind == ast.KindPostfixUnaryExpression {
			postfix := node.AsPostfixUnaryExpression()
			operand = postfix.Operand
			operator = postfix.Operator
		} else {
			prefix := node.AsPrefixUnaryExpression()
			operand = prefix.Operand
			operator = prefix.Operator
		}

		return operator == ast.KindPlusPlusToken && isMatchingIdentifier(operand, name)
	}

	// Check for assignment expressions
	if ast.IsBinaryExpression(node) {
		binExpr := node.AsBinaryExpression()

		// Check left side is the index variable
		if !isMatchingIdentifier(binExpr.Left, name) {
			return false
		}

		// x += 1
		if binExpr.OperatorToken.Kind == ast.KindPlusEqualsToken {
			return isLiteral(binExpr.Right, 1)
		}

		// x = x + 1 or x = 1 + x
		if binExpr.OperatorToken.Kind == ast.KindEqualsToken {
			if !ast.IsBinaryExpression(binExpr.Right) {
				return false
			}

			rightBinExpr := binExpr.Right.AsBinaryExpression()
			if rightBinExpr.OperatorToken.Kind != ast.KindPlusToken {
				return false
			}

			// x = x + 1
			if isMatchingIdentifier(rightBinExpr.Left, name) && isLiteral(rightBinExpr.Right, 1) {
				return true
			}

			// x = 1 + x
			if isLiteral(rightBinExpr.Left, 1) && isMatchingIdentifier(rightBinExpr.Right, name) {
				return true
			}
		}
	}

	return false
}

// contains checks if the outer node contains the inner node
func contains(outer *ast.Node, inner *ast.Node) bool {
	return outer.Pos() <= inner.Pos() && outer.End() >= inner.End()
}

// getNodeText returns the text representation of a node
func getNodeText(sourceFile *ast.SourceFile, node *ast.Node) string {
	r := utils.TrimNodeTextRange(sourceFile, node)
	return sourceFile.Text()[r.Pos():r.End()]
}

// isIndexOnlyUsedWithArray checks if the index variable is only used for array access
func isIndexOnlyUsedWithArray(ctx rule.RuleContext, body *ast.Node, indexName string, arrayExpression *ast.Node) bool {
	arrayText := getNodeText(ctx.SourceFile, arrayExpression)
	allValid := true

	// Traverse the body to find all uses of the index variable
	var checkNode func(*ast.Node) bool
	checkNode = func(node *ast.Node) bool {
		if node == nil {
			return false
		}

		// If this is an identifier matching the index name
		if ast.IsIdentifier(node) && node.AsIdentifier().Text == indexName {
			// Check if it's within the body
			if !contains(body, node) {
				return false
			}

			parent := node.Parent

			// The identifier must be used as the property in a member expression
			// like array[i] where:
			// - parent is ElementAccessExpression
			// - parent.object is not 'this'
			// - parent.object text matches the array
			// - the identifier is the ArgumentExpression (the index)
			// - it's not being assigned to

			if ast.IsElementAccessExpression(parent) {
				elemAccess := parent.AsElementAccessExpression()

				// Check that the object is not 'this'
				if elemAccess.Expression.Kind == ast.KindThisKeyword {
					allValid = false
					return true
				}

				// Check that this node is the ArgumentExpression (index)
				if elemAccess.ArgumentExpression == node {
					// Check that the array expression matches
					objectText := getNodeText(ctx.SourceFile, elemAccess.Expression)
					if objectText == arrayText {
						// Check that it's not being assigned to
						if !isAssignee(parent) {
							// Valid usage
							return false
						}
					}
				}
			}

			// If we reach here, it's an invalid usage of the index
			allValid = false
			return true
		}

		// Recursively check children
		node.ForEachChild(func(child *ast.Node) bool {
			return checkNode(child)
		})

		return false
	}

	checkNode(body)

	return allValid
}

// isAssignee checks if a node is the left-hand side of an assignment
func isAssignee(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}

	switch parent.Kind {
	case ast.KindBinaryExpression:
		binExpr := parent.AsBinaryExpression()
		// Check for compound assignment or regular assignment
		if isAssignmentOperator(binExpr.OperatorToken.Kind) {
			return binExpr.Left == node
		}
	case ast.KindPostfixUnaryExpression, ast.KindPrefixUnaryExpression:
		// ++, --, etc.
		return true
	case ast.KindDeleteExpression:
		return true
	case ast.KindArrayBindingPattern, ast.KindObjectBindingPattern:
		return true
	}

	return false
}

// isAssignmentOperator checks if the token is an assignment operator
func isAssignmentOperator(kind ast.Kind) bool {
	switch kind {
	case ast.KindEqualsToken,
		ast.KindPlusEqualsToken,
		ast.KindMinusEqualsToken,
		ast.KindAsteriskEqualsToken,
		ast.KindAsteriskAsteriskEqualsToken,
		ast.KindSlashEqualsToken,
		ast.KindPercentEqualsToken,
		ast.KindLessThanLessThanEqualsToken,
		ast.KindGreaterThanGreaterThanEqualsToken,
		ast.KindGreaterThanGreaterThanGreaterThanEqualsToken,
		ast.KindAmpersandEqualsToken,
		ast.KindBarEqualsToken,
		ast.KindCaretEqualsToken:
		return true
	}
	return false
}
