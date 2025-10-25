package no_unmodified_loop_condition

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoUnmodifiedLoopConditionRule implements the no-unmodified-loop-condition rule
// Disallow unmodified loop conditions
var NoUnmodifiedLoopConditionRule = rule.Rule{
	Name: "no-unmodified-loop-condition",
	Run:  run,
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	typeChecker := ctx.TypeChecker
	if typeChecker == nil {
		return rule.RuleListeners{}
	}

	return rule.RuleListeners{
		ast.KindWhileStatement: func(node *ast.Node) {
			whileStmt := node.AsWhileStatement()
			if whileStmt == nil || whileStmt.Expression == nil {
				return
			}

			checkLoopCondition(ctx, typeChecker, whileStmt.Expression, whileStmt.Statement)
		},
		ast.KindDoStatement: func(node *ast.Node) {
			doStmt := node.AsDoStatement()
			if doStmt == nil || doStmt.Expression == nil {
				return
			}

			checkLoopCondition(ctx, typeChecker, doStmt.Expression, doStmt.Statement)
		},
		ast.KindForStatement: func(node *ast.Node) {
			forStmt := node.AsForStatement()
			if forStmt == nil || forStmt.Condition == nil {
				return
			}

			checkLoopCondition(ctx, typeChecker, forStmt.Condition, forStmt.Statement)
		},
	}
}

// checkLoopCondition checks if variables in the loop condition are modified in the loop body
func checkLoopCondition(ctx rule.RuleContext, typeChecker *checker.Checker, condition *ast.Node, body *ast.Node) {
	if condition == nil || body == nil {
		return
	}

	// Collect all identifiers used in the condition
	conditionVars := collectIdentifiers(condition)
	if len(conditionVars) == 0 {
		return
	}

	// Collect all identifiers that are modified in the loop body
	modifiedVars := collectModifiedIdentifiers(body)

	// Find symbols for condition variables
	conditionSymbols := make(map[*ast.Symbol]string)
	for _, varNode := range conditionVars {
		if varNode == nil {
			continue
		}

		// Skip if this is a property access (e.g., obj.prop)
		// We only care about direct variable references
		if isPartOfPropertyAccess(varNode) {
			continue
		}

		// Skip if this is a call expression (e.g., func())
		// Functions can have side effects
		if isPartOfCallExpression(varNode) {
			continue
		}

		symbol := typeChecker.GetSymbolAtLocation(varNode)
		if symbol != nil {
			identifier := varNode.AsIdentifier()
			if identifier != nil {
				conditionSymbols[symbol] = identifier.Text
			}
		}
	}

	// Find symbols for modified variables
	modifiedSymbols := make(map[*ast.Symbol]bool)
	for _, modNode := range modifiedVars {
		if modNode == nil {
			continue
		}

		symbol := typeChecker.GetSymbolAtLocation(modNode)
		if symbol != nil {
			modifiedSymbols[symbol] = true
		}
	}

	// Check which condition variables are not modified
	for symbol, varName := range conditionSymbols {
		if !modifiedSymbols[symbol] {
			// This variable is in the condition but not modified in the loop body
			ctx.ReportNode(condition, rule.RuleMessage{
				Id:          "loopConditionNotModified",
				Description: "'" + varName + "' is not modified in this loop.",
			})
		}
	}
}

// collectIdentifiers collects all identifier nodes in an expression
func collectIdentifiers(node *ast.Node) []*ast.Node {
	if node == nil {
		return nil
	}

	var identifiers []*ast.Node

	var visit func(*ast.Node)
	visit = func(n *ast.Node) {
		if n == nil {
			return
		}

		if n.Kind == ast.KindIdentifier {
			identifiers = append(identifiers, n)
		}

		// Visit children
		n.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}

	visit(node)
	return identifiers
}

// collectModifiedIdentifiers collects all identifier nodes that are modified (assigned to)
func collectModifiedIdentifiers(node *ast.Node) []*ast.Node {
	if node == nil {
		return nil
	}

	var modified []*ast.Node

	var visit func(*ast.Node)
	visit = func(n *ast.Node) {
		if n == nil {
			return
		}

		switch n.Kind {
		case ast.KindBinaryExpression:
			binExpr := n.AsBinaryExpression()
			if binExpr != nil && isAssignmentOperator(binExpr.OperatorToken) {
				// Collect identifiers on the left side of assignment
				leftIds := collectIdentifiers(binExpr.Left)
				modified = append(modified, leftIds...)
			}

		case ast.KindPrefixUnaryExpression:
			prefixExpr := n.AsPrefixUnaryExpression()
			if prefixExpr != nil && (prefixExpr.Operator == ast.KindPlusPlusToken || prefixExpr.Operator == ast.KindMinusMinusToken) {
				// ++ or -- operators modify the operand
				opIds := collectIdentifiers(prefixExpr.Operand)
				modified = append(modified, opIds...)
			}

		case ast.KindPostfixUnaryExpression:
			postfixExpr := n.AsPostfixUnaryExpression()
			if postfixExpr != nil && (postfixExpr.Operator == ast.KindPlusPlusToken || postfixExpr.Operator == ast.KindMinusMinusToken) {
				// ++ or -- operators modify the operand
				opIds := collectIdentifiers(postfixExpr.Operand)
				modified = append(modified, opIds...)
			}
		}

		// Visit children
		n.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}

	visit(node)
	return modified
}

// isAssignmentOperator checks if a token is an assignment operator
func isAssignmentOperator(token *ast.Node) bool {
	if token == nil {
		return false
	}

	switch token.Kind {
	case ast.KindEqualsToken, // =
		ast.KindPlusEqualsToken,              // +=
		ast.KindMinusEqualsToken,             // -=
		ast.KindAsteriskEqualsToken,          // *=
		ast.KindSlashEqualsToken,             // /=
		ast.KindPercentEqualsToken,           // %=
		ast.KindAsteriskAsteriskEqualsToken,  // **=
		ast.KindLessThanLessThanEqualsToken,  // <<=
		ast.KindGreaterThanGreaterThanEqualsToken, // >>=
		ast.KindGreaterThanGreaterThanGreaterThanEqualsToken, // >>>=
		ast.KindAmpersandEqualsToken,    // &=
		ast.KindBarEqualsToken,          // |=
		ast.KindCaretEqualsToken,        // ^=
		ast.KindBarBarEqualsToken,       // ||=
		ast.KindAmpersandAmpersandEqualsToken, // &&=
		ast.KindQuestionQuestionEqualsToken:   // ??=
		return true
	}
	return false
}

// isPartOfPropertyAccess checks if an identifier is part of a property access expression
func isPartOfPropertyAccess(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}

	parent := node.Parent

	// Check if this is the property name in a property access (e.g., obj.prop -> "prop")
	if parent.Kind == ast.KindPropertyAccessExpression {
		propAccess := parent.AsPropertyAccessExpression()
		if propAccess != nil && propAccess.Name() == node {
			return true
		}
	}

	// Check if the parent is a property access and this node is the entire expression
	// This handles cases like "foo.bar" where we want to treat it as dynamic
	if parent.Kind == ast.KindPropertyAccessExpression {
		return true
	}

	return false
}

// isPartOfCallExpression checks if an identifier is being called as a function
func isPartOfCallExpression(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}

	parent := node.Parent

	// Check if this identifier is the callee of a call expression
	if parent.Kind == ast.KindCallExpression {
		callExpr := parent.AsCallExpression()
		if callExpr != nil && callExpr.Expression == node {
			return true
		}
	}

	return false
}
