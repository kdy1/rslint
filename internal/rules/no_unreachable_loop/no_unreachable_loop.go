package no_unreachable_loop

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Options defines the configuration options for the no-unreachable-loop rule
type Options struct {
	Ignore []string `json:"ignore"` // Loop types to ignore (e.g., ["WhileStatement", "DoStatement"])
}

func parseOptions(options any) Options {
	opts := Options{
		Ignore: []string{},
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
		if ignoreArray, ok := optsMap["ignore"].([]interface{}); ok {
			for _, v := range ignoreArray {
				if str, ok := v.(string); ok {
					opts.Ignore = append(opts.Ignore, str)
				}
			}
		}
	}

	return opts
}

// NoUnreachableLoopRule implements the no-unreachable-loop rule
// Disallow loops with a body that allows only one iteration
var NoUnreachableLoopRule = rule.Rule{
	Name: "no-unreachable-loop",
	Run:  run,
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := parseOptions(options)

	// Helper to check if a loop type should be ignored
	shouldIgnore := func(loopType string) bool {
		for _, ignore := range opts.Ignore {
			if ignore == loopType {
				return true
			}
		}
		return false
	}

	checkLoop := func(node *ast.Node, loopType string) {
		if shouldIgnore(loopType) {
			return
		}

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
		case ast.KindForInStatement, ast.KindForOfStatement:
			forInOfStmt := node.AsForInOrOfStatement()
			if forInOfStmt != nil {
				body = forInOfStmt.Statement
			}
		}

		if body == nil {
			return
		}

		// Check if the loop body always exits on first iteration
		if loopAlwaysExitsOnFirstIteration(body, node) {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "invalid",
				Description: "Invalid loop. Its body allows only one iteration.",
			})
		}
	}

	return rule.RuleListeners{
		ast.KindWhileStatement: func(node *ast.Node) {
			checkLoop(node, "WhileStatement")
		},
		ast.KindDoStatement: func(node *ast.Node) {
			checkLoop(node, "DoStatement")
		},
		ast.KindForStatement: func(node *ast.Node) {
			checkLoop(node, "ForStatement")
		},
		ast.KindForInStatement: func(node *ast.Node) {
			checkLoop(node, "ForInStatement")
		},
		ast.KindForOfStatement: func(node *ast.Node) {
			checkLoop(node, "ForOfStatement")
		},
	}
}

// loopAlwaysExitsOnFirstIteration checks if a loop body guarantees exit on first iteration
func loopAlwaysExitsOnFirstIteration(body *ast.Node, loopNode *ast.Node) bool {
	if body == nil {
		return false
	}

	// Check all paths through the body
	return allPathsExitLoop(body, loopNode)
}

// allPathsExitLoop checks if all code paths in a statement exit the loop
func allPathsExitLoop(stmt *ast.Node, loopNode *ast.Node) bool {
	if stmt == nil {
		return false
	}

	kind := stmt.Kind

	switch kind {
	case ast.KindReturnStatement, ast.KindThrowStatement:
		// These exit the entire function
		return true

	case ast.KindBreakStatement:
		// Break without label exits the current loop
		breakStmt := stmt.AsBreakStatement()
		if breakStmt == nil {
			return true
		}
		// If there's a label, we'd need to check if it targets this loop
		// For simplicity, we assume unlabeled breaks exit this loop
		return breakStmt.Label == nil

	case ast.KindContinueStatement:
		// Continue doesn't exit the loop, just skips to next iteration
		return false

	case ast.KindBlock:
		// A block exits if it contains any statement that exits
		block := stmt.AsBlock()
		if block == nil || block.Statements == nil {
			return false
		}

		for _, s := range block.Statements.Nodes {
			if allPathsExitLoop(&s, loopNode) {
				return true
			}
		}
		return false

	case ast.KindIfStatement:
		// If statement exits if both branches exist and both exit
		ifStmt := stmt.AsIfStatement()
		if ifStmt == nil {
			return false
		}

		thenExits := allPathsExitLoop(ifStmt.ThenStatement, loopNode)
		elseExits := ifStmt.ElseStatement != nil && allPathsExitLoop(ifStmt.ElseStatement, loopNode)

		return thenExits && elseExits

	case ast.KindSwitchStatement:
		// Switch exits if all cases exit (including default)
		switchStmt := stmt.AsSwitchStatement()
		if switchStmt == nil || switchStmt.CaseBlock == nil {
			return false
		}

		caseBlock := switchStmt.CaseBlock.AsCaseBlock()
		if caseBlock == nil || caseBlock.Clauses == nil {
			return false
		}

		hasDefault := false
		allCasesExit := true

		for _, clause := range caseBlock.Clauses.Nodes {
			clauseNode := clause.AsCaseOrDefaultClause()
			if clauseNode == nil {
				continue
			}

			if clauseNode.Kind == ast.KindDefaultClause {
				hasDefault = true
			}

			if clauseNode.Statements != nil {
				statements := clauseNode.Statements.Nodes
				clauseExits := false
				for _, s := range statements {
					if allPathsExitLoop(&s, loopNode) {
						clauseExits = true
						break
					}
				}
				if !clauseExits {
					allCasesExit = false
				}
			} else {
				allCasesExit = false
			}
		}

		return hasDefault && allCasesExit

	default:
		return false
	}
}
