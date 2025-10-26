package no_dupe_else_if

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoDupeElseIfRule implements the no-dupe-else-if rule
// Disallow duplicate conditions in if-else-if chains
var NoDupeElseIfRule = rule.Rule{
	Name: "no-dupe-else-if",
	Run:  run,
}

// getConditionText extracts the normalized text of a condition
func getConditionText(node *ast.Node) string {
	if node == nil {
		return ""
	}

	src := node.GetSourceFile()
	if src == nil {
		return ""
	}

	// Get the text content of the condition
	text := src.Text()
	start := node.Pos()
	end := node.End()

	if start >= 0 && end <= len(text) && start < end {
		return text[start:end]
	}

	return ""
}

// checkIfStatement checks an if statement for duplicate conditions
func checkIfStatement(ctx rule.RuleContext, node *ast.Node) {
	if node == nil {
		return
	}

	// Track all conditions in the if-else-if chain
	var conditions []*ast.Node

	// Get the initial condition
	initialCondition := node.Expression()
	if initialCondition != nil {
		conditions = append(conditions, initialCondition)
	}

	// Walk through the else-if chain
	current := node
	for current != nil {
		elseStmt := current.ElseStatement()
		if elseStmt == nil {
			break
		}

		// Check if it's an else-if (not just an else block)
		if elseStmt.Kind == ast.KindIfStatement {
			condition := elseStmt.Expression()
			if condition != nil {
				// Check against all previous conditions
				conditionText := getConditionText(condition)
				if conditionText != "" {
					for _, prevCondition := range conditions {
						prevText := getConditionText(prevCondition)
						if prevText == conditionText {
							// Found a duplicate
							ctx.ReportNode(condition, rule.RuleMessage{
								Id:          "unexpected",
								Description: "This branch can never execute. Its condition is a duplicate or covered by previous conditions in the if-else-if chain.",
							})
							break
						}
					}
				}
				conditions = append(conditions, condition)
			}
			current = elseStmt
		} else {
			// It's just an else block, stop checking
			break
		}
	}
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	return rule.RuleListeners{
		ast.KindIfStatement: func(node *ast.Node) {
			checkIfStatement(ctx, node)
		},
	}
}
