package constructor_super

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Message builders
func buildMissingAll() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingAll",
		Description: "Expected to call 'super()' in all paths.",
	}
}

func buildMissingSome() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingSome",
		Description: "Expected to call 'super()' in some paths.",
	}
}

func buildDuplicate() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "duplicate",
		Description: "Unexpected duplicate 'super()' call.",
	}
}

func buildBadSuper() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "badSuper",
		Description: "Unexpected 'super()' call because this class does not extend a valid constructor.",
	}
}

// isConstructor checks if a node is a constructor method
func isConstructor(node *ast.Node) bool {
	if node == nil {
		return false
	}

	// Check if it's a constructor method
	if node.Kind == ast.KindConstructor {
		return true
	}

	// For method declarations, check if it's named "constructor"
	if node.Kind == ast.KindMethodDeclaration {
		name := node.Name()
		if name != nil && name.Kind == ast.KindIdentifier && name.Text() == "constructor" {
			return true
		}
	}

	return false
}

// getClassNode gets the class declaration/expression for a constructor
func getClassNode(constructorNode *ast.Node) *ast.Node {
	if constructorNode == nil {
		return nil
	}

	// Walk up the parent chain to find the class
	current := constructorNode.Parent()
	for current != nil {
		if current.Kind == ast.KindClassDeclaration || current.Kind == ast.KindClassExpression {
			return current
		}
		current = current.Parent()
	}

	return nil
}

// hasValidExtends checks if a class extends a valid constructor
func hasValidExtends(classNode *ast.Node) bool {
	if classNode == nil {
		return false
	}

	// Get heritage clause (extends clause)
	heritageClauses := classNode.HeritageClauses()
	if heritageClauses == nil || len(heritageClauses.Nodes) == 0 {
		return false
	}

	// Look for extends clause (token = ExtendsKeyword)
	for _, clause := range heritageClauses.Nodes {
		if clause == nil {
			continue
		}
		// Check if this is an extends clause
		if clause.Token() == ast.KindExtendsKeyword {
			types := clause.Types()
			if types != nil && len(types) > 0 {
				// Check if the extended type is valid (not null, not a primitive)
				extendsExpr := types[0].Expression()
				if extendsExpr != nil && !isInvalidExtends(extendsExpr) {
					return true
				}
			}
		}
	}

	return false
}

// isInvalidExtends checks if an extends expression is invalid (null, literal, etc.)
func isInvalidExtends(node *ast.Node) bool {
	if node == nil {
		return true
	}

	switch node.Kind {
	case ast.KindNullKeyword:
		return true
	case ast.KindNumericLiteral, ast.KindStringLiteral, ast.KindTrueKeyword, ast.KindFalseKeyword:
		return true
	case ast.KindParenthesizedExpression:
		// Check inner expression
		expr := node.Expression()
		return isInvalidExtends(expr)
	case ast.KindBinaryExpression:
		// Binary expressions like B = 5, B += C are invalid
		// Only simple references or valid expressions are OK
		// We need to be conservative here
		operator := node.OperatorToken()
		if operator != nil {
			// Assignment operators are invalid
			switch operator.Kind {
			case ast.KindEqualsToken, ast.KindPlusEqualsToken, ast.KindMinusEqualsToken,
				ast.KindAsteriskEqualsToken, ast.KindSlashEqualsToken:
				return true
			}
		}
	}

	return false
}

// isPossibleConstructor checks if an expression could be a constructor
// This handles cases like: class A extends (B = C) where we need to check if the result is a constructor
func isPossibleConstructor(node *ast.Node) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindClassExpression:
		return true
	case ast.KindIdentifier:
		// Could be a constructor reference
		return true
	case ast.KindParenthesizedExpression:
		// Check inner expression
		expr := node.Expression()
		return isPossibleConstructor(expr)
	case ast.KindBinaryExpression:
		// For assignments like (B = C), check the right side
		operator := node.OperatorToken()
		if operator != nil && operator.Kind == ast.KindEqualsToken {
			right := node.Right()
			return isPossibleConstructor(right)
		}
		// For other operators, it's not a constructor
		return false
	case ast.KindConditionalExpression:
		// For ternary, both branches must be constructors
		whenTrue := node.WhenTrue()
		whenFalse := node.WhenFalse()
		return isPossibleConstructor(whenTrue) && isPossibleConstructor(whenFalse)
	case ast.KindLogicalAndExpression, ast.KindLogicalOrExpression:
		// For logical expressions, check the right side (result)
		right := node.Right()
		return isPossibleConstructor(right)
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		// Could be a constructor reference like Class.Static
		return true
	case ast.KindCallExpression:
		// Could return a constructor
		return true
	}

	// For other node types, assume it could be a constructor
	return true
}

// analyzeSuperCalls analyzes super() calls in a constructor body
type superCallAnalysis struct {
	hasSuperCall       bool     // true if any super() call exists
	allPathsHaveSuper  bool     // true if all paths call super()
	superCallLocations []*ast.Node // locations of all super() calls
}

// analyzeSuperCallsInBody analyzes super() calls in a constructor body
func analyzeSuperCallsInBody(body *ast.Node) superCallAnalysis {
	result := superCallAnalysis{
		superCallLocations: make([]*ast.Node, 0),
	}

	if body == nil {
		result.allPathsHaveSuper = false
		return result
	}

	// Find all super() calls
	findSuperCalls(body, &result.superCallLocations)
	result.hasSuperCall = len(result.superCallLocations) > 0

	// Analyze if all paths have super
	result.allPathsHaveSuper = checkAllPathsHaveSuper(body)

	return result
}

// findSuperCalls recursively finds all super() call expressions
func findSuperCalls(node *ast.Node, locations *[]*ast.Node) {
	if node == nil {
		return
	}

	// Check if this is a super() call
	if node.Kind == ast.KindCallExpression {
		expr := node.Expression()
		if expr != nil && expr.Kind == ast.KindSuperKeyword {
			*locations = append(*locations, node)
		}
	}

	// Don't recurse into nested functions/classes
	if isFunctionOrClassNode(node) {
		return
	}

	// Recurse into children
	node.ForEachChild(func(child *ast.Node) bool {
		findSuperCalls(child, locations)
		return false // Continue iteration
	})
}

// isFunctionOrClassNode checks if a node is a function or class (boundary for super call search)
func isFunctionOrClassNode(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
		ast.KindMethodDeclaration, ast.KindClassDeclaration, ast.KindClassExpression:
		return true
	}
	return false
}

// checkAllPathsHaveSuper checks if all code paths in the body call super()
func checkAllPathsHaveSuper(body *ast.Node) bool {
	if body == nil {
		return false
	}

	if body.Kind != ast.KindBlock {
		return false
	}

	statements := body.Statements()
	if len(statements) == 0 {
		return false
	}

	// Use a simplified control flow analysis
	return analyzeStatements(statements)
}

// analyzeStatements checks if super() is called in all code paths
func analyzeStatements(statements []*ast.Node) bool {
	for _, stmt := range statements {
		if stmt == nil {
			continue
		}

		// If we find a super() call at this level, all paths up to here have it
		if hasSuperCall(stmt) {
			return true
		}

		// Check control flow statements
		switch stmt.Kind {
		case ast.KindIfStatement:
			// If-else with super in all branches
			ifStmt := stmt.AsIfStatement()
			if ifStmt != nil {
				thenHasSuper := statementHasSuper(ifStmt.ThenStatement)
				elseStmt := ifStmt.ElseStatement

				if elseStmt != nil {
					// Has else clause
					elseHasSuper := statementHasSuper(elseStmt)
					if thenHasSuper && elseHasSuper {
						return true
					}
				}
			}

		case ast.KindSwitchStatement:
			// Switch with super in all cases (including default)
			if switchHasSuper(stmt) {
				return true
			}

		case ast.KindReturnStatement:
			// Early return means this path doesn't need super
			return true

		case ast.KindThrowStatement:
			// Throw means this path doesn't need super
			return true
		}
	}

	return false
}

// hasSuperCall checks if a statement contains a direct super() call
func hasSuperCall(stmt *ast.Node) bool {
	if stmt == nil {
		return false
	}

	// Direct call
	if stmt.Kind == ast.KindExpressionStatement {
		expr := stmt.Expression()
		if expr != nil && expr.Kind == ast.KindCallExpression {
			callExpr := expr.Expression()
			if callExpr != nil && callExpr.Kind == ast.KindSuperKeyword {
				return true
			}
		}
	}

	return false
}

// statementHasSuper checks if a statement (or block) has super call
func statementHasSuper(stmt *ast.Node) bool {
	if stmt == nil {
		return false
	}

	if stmt.Kind == ast.KindBlock {
		return analyzeStatements(stmt.Statements())
	}

	return hasSuperCall(stmt)
}

// switchHasSuper checks if a switch statement has super in all branches
func switchHasSuper(switchStmt *ast.Node) bool {
	if switchStmt == nil || switchStmt.Kind != ast.KindSwitchStatement {
		return false
	}

	caseBlock := switchStmt.CaseBlock()
	if caseBlock == nil {
		return false
	}

	clauses := caseBlock.Clauses()
	if len(clauses) == 0 {
		return false
	}

	hasDefault := false
	allClausesHaveSuper := true

	for _, clause := range clauses {
		if clause == nil {
			continue
		}

		if clause.Kind == ast.KindDefaultClause {
			hasDefault = true
		}

		// Check if this clause has super
		statements := clause.Statements()
		if !analyzeStatements(statements) {
			allClausesHaveSuper = false
		}
	}

	// All cases must have super AND there must be a default case
	return hasDefault && allClausesHaveSuper
}

// ConstructorSuperRule enforces proper super() calls in constructors
var ConstructorSuperRule = rule.CreateRule(rule.Rule{
	Name: "constructor-super",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindConstructor: func(node *ast.Node) {
				// Check if this is a constructor
				if !isConstructor(node) {
					return
				}

				// Get the class this constructor belongs to
				classNode := getClassNode(node)
				if classNode == nil {
					return
				}

				// Check if the class extends something
				hasExtends := hasValidExtends(classNode)

				// Get the constructor body
				body := node.Body()

				// Analyze super() calls
				analysis := analyzeSuperCallsInBody(body)

				if hasExtends {
					// Derived class: must call super()
					if !analysis.hasSuperCall {
						// No super() call at all
						ctx.ReportNode(node, buildMissingAll())
					} else if !analysis.allPathsHaveSuper {
						// super() called in some paths but not all
						ctx.ReportNode(node, buildMissingSome())
					} else if len(analysis.superCallLocations) > 1 {
						// Multiple super() calls - report duplicates
						for i := 1; i < len(analysis.superCallLocations); i++ {
							ctx.ReportNode(analysis.superCallLocations[i], buildDuplicate())
						}
					}
				} else {
					// Non-derived class or extends null: must NOT call super()
					if analysis.hasSuperCall {
						// Report each super() call as invalid
						for _, superCall := range analysis.superCallLocations {
							ctx.ReportNode(superCall, buildBadSuper())
						}
					}
				}

				// Special check for extends with invalid expressions
				// Check if class extends something but it's invalid (like null, literals, etc.)
				heritageClauses := classNode.HeritageClauses()
				if heritageClauses != nil && len(heritageClauses.Nodes) > 0 {
					for _, clause := range heritageClauses.Nodes {
						if clause == nil {
							continue
						}
						if clause.Token() == ast.KindExtendsKeyword {
							types := clause.Types()
							if types != nil && len(types) > 0 {
								extendsExpr := types[0].Expression()
								// If extends is invalid AND we have super() calls, report them
								if isInvalidExtends(extendsExpr) && analysis.hasSuperCall {
									for _, superCall := range analysis.superCallLocations {
										ctx.ReportNode(superCall, buildBadSuper())
									}
								}
							}
						}
					}
				}
			},
		}
	},
})
