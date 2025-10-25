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

func buildDuplicateConditionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "duplicateCondition",
		Description: "This branch can never execute. Its condition is a duplicate or covered by previous conditions in the if-else-if chain.",
	}
}

// getConditionText returns the source text of a condition node
func getConditionText(ctx rule.RuleContext, condition *ast.Node) string {
	if condition == nil {
		return ""
	}

	text := ctx.SourceFile.Text()
	start := condition.Pos()
	end := condition.End()

	if start >= 0 && end <= len(text) && start < end {
		return text[start:end]
	}
	return ""
}

// checkDuplicateConditions checks for duplicate conditions in an if-else-if chain
func checkDuplicateConditions(ctx rule.RuleContext, ifStmt *ast.Node) {
	if ifStmt == nil || ifStmt.Kind != ast.KindIfStatement {
		return
	}

	// Track seen conditions in this if-else-if chain
	seenConditions := make(map[string]*ast.Node)

	currentIf := ifStmt

	// Traverse the if-else-if chain
	for currentIf != nil && currentIf.Kind == ast.KindIfStatement {
		stmt := currentIf.AsIfStatement()
		if stmt == nil || stmt.Expression == nil {
			break
		}

		conditionText := getConditionText(ctx, stmt.Expression)
		if conditionText == "" {
			// Move to next else-if
			if stmt.ElseStatement != nil {
				currentIf = stmt.ElseStatement
			} else {
				break
			}
			continue
		}

		// Check if this condition was already seen in the chain
		if _, exists := seenConditions[conditionText]; exists {
			// Report duplicate condition
			ctx.ReportNode(stmt.Expression, buildDuplicateConditionMessage())
		} else {
			// Record this condition
			seenConditions[conditionText] = stmt.Expression
		}

		// Move to the next else-if in the chain
		if stmt.ElseStatement != nil && stmt.ElseStatement.Kind == ast.KindIfStatement {
			currentIf = stmt.ElseStatement
		} else {
			// No more else-if statements
			break
		}
	}
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	// Track which if statements have already been processed as part of a chain
	// to avoid duplicate checking
	processedNodes := make(map[*ast.Node]bool)

	return rule.RuleListeners{
		ast.KindIfStatement: func(node *ast.Node) {
			// Skip if this node was already processed as part of another chain
			if processedNodes[node] {
				return
			}

			// Mark this node and all nodes in its chain as processed
			currentIf := node
			for currentIf != nil && currentIf.Kind == ast.KindIfStatement {
				processedNodes[currentIf] = true

				stmt := currentIf.AsIfStatement()
				if stmt == nil || stmt.ElseStatement == nil {
					break
				}

				// Move to next else-if
				if stmt.ElseStatement.Kind == ast.KindIfStatement {
					currentIf = stmt.ElseStatement
				} else {
					break
				}
			}

			// Now check for duplicates in the chain starting from this node
			checkDuplicateConditions(ctx, node)
		},
	}
}
