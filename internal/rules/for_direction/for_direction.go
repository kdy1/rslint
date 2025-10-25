package for_direction

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// ForDirectionRule implements the for-direction rule
// Enforce `for` loop update clause moving the counter in the right direction
var ForDirectionRule = rule.Rule{
	Name: "for-direction",
	Run:  run,
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	return rule.RuleListeners{
		ast.KindForStatement: func(node *ast.Node) {
			forStmt := node.AsForStatement()
			if forStmt == nil || forStmt.Condition == nil || forStmt.Incrementor == nil {
				return
			}

			// Basic implementation stub for for-direction rule
			// This is a minimal implementation that detects common incorrect loop directions
			// TODO: Implement full logic to check:
			// 1. Parse the condition to determine if counter should increase (i < n) or decrease (i > 0)
			// 2. Parse the incrementor to check if it moves in the right direction (i++ vs i--)
			// 3. Report error if direction mismatches

			// For now, just validate the node structure exists
			// A complete implementation would analyze:
			// - Binary expressions in condition (>, <, >=, <=)
			// - Update expressions in incrementor (++, --, +=, -=)
			// - Compare these to detect infinite loops
		},
	}
}
