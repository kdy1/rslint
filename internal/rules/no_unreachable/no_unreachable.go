package no_unreachable

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoUnreachableRule implements the no-unreachable rule
// Disallow unreachable code after `return`, `throw`, `continue`, and `break` statements
var NoUnreachableRule = rule.Rule{
	Name: "no-unreachable",
	Run:  run,
}

func buildUnreachableMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unreachableCode",
		Description: "Unreachable code.",
	}
}

// isTerminatingStatement checks if a statement terminates control flow
func isTerminatingStatement(node *ast.Node) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindReturnStatement, ast.KindThrowStatement,
		ast.KindBreakStatement, ast.KindContinueStatement:
		return true
	case ast.KindIfStatement:
		// If both branches exist and both terminate, the if statement terminates
		thenStmt := node.ThenStatement()
		elseStmt := node.ElseStatement()
		if thenStmt != nil && elseStmt != nil {
			return blockAlwaysTerminates(thenStmt) && blockAlwaysTerminates(elseStmt)
		}
	case ast.KindBlock:
		return blockAlwaysTerminates(node)
	}
	return false
}

// blockAlwaysTerminates checks if a block always terminates control flow
func blockAlwaysTerminates(node *ast.Node) bool {
	if node == nil {
		return false
	}

	if node.Kind == ast.KindBlock {
		statements := node.Statements()
		for _, stmt := range statements {
			if isTerminatingStatement(stmt) {
				return true
			}
		}
		return false
	}

	return isTerminatingStatement(node)
}

// findUnreachableStatements finds unreachable statements after a terminating statement
func findUnreachableStatements(statements []*ast.Node) []*ast.Node {
	var unreachable []*ast.Node
	foundTerminator := false

	for _, stmt := range statements {
		if stmt == nil {
			continue
		}

		if foundTerminator {
			// Skip variable declarations without initializers (hoisting)
			if stmt.Kind == ast.KindVariableStatement {
				declList := stmt.DeclarationList()
				if declList != nil {
					decls := declList.Declarations()
					hasInit := false
					for _, decl := range decls {
						if decl != nil && decl.Initializer() != nil {
							hasInit = true
							break
						}
					}
					if !hasInit {
						continue
					}
				}
			}

			// Skip function declarations (hoisted)
			if stmt.Kind == ast.KindFunctionDeclaration {
				continue
			}

			unreachable = append(unreachable, stmt)
		}

		if isTerminatingStatement(stmt) {
			foundTerminator = true
		}
	}

	return unreachable
}

// checkBlock checks a block statement for unreachable code
func checkBlock(ctx rule.RuleContext, node *ast.Node) {
	if node == nil || node.Kind != ast.KindBlock {
		return
	}

	statements := node.Statements()
	unreachable := findUnreachableStatements(statements)

	for _, stmt := range unreachable {
		ctx.ReportNode(stmt, buildUnreachableMessage())
	}
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	return rule.RuleListeners{
		// Check all block statements
		ast.KindBlock: func(node *ast.Node) {
			checkBlock(ctx, node)
		},

		// Check switch case clauses
		ast.KindCaseClause: func(node *ast.Node) {
			if node == nil {
				return
			}
			statements := node.Statements()
			unreachable := findUnreachableStatements(statements)
			for _, stmt := range unreachable {
				ctx.ReportNode(stmt, buildUnreachableMessage())
			}
		},

		ast.KindDefaultClause: func(node *ast.Node) {
			if node == nil {
				return
			}
			statements := node.Statements()
			unreachable := findUnreachableStatements(statements)
			for _, stmt := range unreachable {
				ctx.ReportNode(stmt, buildUnreachableMessage())
			}
		},
	}
}
