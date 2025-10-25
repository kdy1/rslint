package no_unreachable

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoUnreachableRule implements the no-unreachable rule
// Disallow unreachable code after return, throw, continue, or break
var NoUnreachableRule = rule.Rule{
	Name: "no-unreachable",
	Run:  run,
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	return rule.RuleListeners{
		ast.KindBlock: func(node *ast.Node) {
			block := node.AsBlock()
			if block == nil || block.Statements == nil {
				return
			}

			statements := block.Statements.Nodes
			for i := 0; i < len(statements)-1; i++ {
				stmt := &statements[i]

				// Check if this statement causes control flow to exit
				if isControlFlowExit(stmt) {
					// Check if there are statements after this one
					nextStmt := &statements[i+1]

					// Report unreachable code on the next statement
					ctx.ReportNode(nextStmt, rule.RuleMessage{
						Id:          "unreachableCode",
						Description: "Unreachable code.",
					})

					// Only report once per unreachable segment
					break
				}
			}
		},
		ast.KindSwitchStatement: func(node *ast.Node) {
			switchStmt := node.AsSwitchStatement()
			if switchStmt == nil || switchStmt.CaseBlock == nil {
				return
			}

			caseBlock := switchStmt.CaseBlock.AsCaseBlock()
			if caseBlock == nil || caseBlock.Clauses == nil {
				return
			}

			// Check each case/default clause for unreachable code
			for _, clause := range caseBlock.Clauses.Nodes {
				clauseNode := clause.AsCaseOrDefaultClause()
				if clauseNode == nil || clauseNode.Statements == nil {
					continue
				}

				statements := clauseNode.Statements.Nodes
				for i := 0; i < len(statements)-1; i++ {
					stmt := &statements[i]

					if isControlFlowExit(stmt) {
						nextStmt := &statements[i+1]
						ctx.ReportNode(nextStmt, rule.RuleMessage{
							Id:          "unreachableCode",
							Description: "Unreachable code.",
						})
						break
					}
				}
			}
		},
	}
}

// isControlFlowExit checks if a statement causes control flow to exit
// (return, throw, break, continue)
func isControlFlowExit(stmt *ast.Node) bool {
	kind := stmt.Kind

	switch kind {
	case ast.KindReturnStatement, ast.KindThrowStatement,
		 ast.KindBreakStatement, ast.KindContinueStatement:
		return true

	case ast.KindIfStatement:
		// If statement exits control flow if both branches exit
		ifStmt := stmt.AsIfStatement()
		if ifStmt == nil {
			return false
		}

		// Check if both then and else branches exist and both exit
		thenExits := ifStmt.ThenStatement != nil && statementExits(ifStmt.ThenStatement)
		elseExits := ifStmt.ElseStatement != nil && statementExits(ifStmt.ElseStatement)

		return thenExits && elseExits

	case ast.KindBlock:
		// A block exits if it contains an exiting statement
		block := stmt.AsBlock()
		if block == nil || block.Statements == nil {
			return false
		}

		for _, s := range block.Statements.Nodes {
			if isControlFlowExit(&s) {
				return true
			}
		}
		return false

	default:
		return false
	}
}

// statementExits checks if a statement guarantees control flow exit
func statementExits(stmt *ast.Node) bool {
	if stmt == nil {
		return false
	}
	return isControlFlowExit(stmt)
}
