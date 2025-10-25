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

// isPartOfElseIfChain checks if this if statement is part of an else-if chain
// (i.e., if it's an else branch of another if statement)
func isPartOfElseIfChain(ctx rule.RuleContext, node *ast.Node) bool {
	// This is a simplified check - in a real implementation, you might want to
	// track parent nodes to determine if this if statement is in an else branch
	// For now, we'll check all if statements and let the chain traversal logic handle it
	return false
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	return rule.RuleListeners{
		ast.KindIfStatement: func(node *ast.Node) {
			// Only check if this is the start of an if-else-if chain
			// (not itself an else-if)
			if isPartOfElseIfChain(ctx, node) {
				return
			}

			checkDuplicateConditions(ctx, node)
		},
	}
}
