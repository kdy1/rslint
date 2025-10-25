package no_duplicate_case

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// NoDuplicateCaseRule implements the no-duplicate-case rule
// Disallow duplicate case labels
var NoDuplicateCaseRule = rule.Rule{
	Name: "no-duplicate-case",
	Run:  run,
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	return rule.RuleListeners{
		ast.KindSwitchStatement: func(node *ast.Node) {
			switchStmt := node.AsSwitchStatement()
			if switchStmt == nil || switchStmt.CaseBlock == nil {
				return
			}

			caseBlock := switchStmt.CaseBlock.AsCaseBlock()
			if caseBlock == nil || caseBlock.Clauses == nil {
				return
			}

			// Track case expressions we've seen
			// Map from normalized expression text to the expression node
			seenCases := make(map[string]*ast.Node)

			for _, clause := range caseBlock.Clauses.Nodes {
				if clause == nil {
					continue
				}

				var caseExpr *ast.Node

				switch clause.Kind {
				case ast.KindCaseClause:
					caseClause := clause.AsCaseClause()
					if caseClause == nil || caseClause.Expression == nil {
						continue
					}
					caseExpr = caseClause.Expression

				case ast.KindDefaultClause:
					// Default clauses don't have expressions to compare
					continue

				default:
					continue
				}

				// Normalize the expression text for comparison
				normalized := normalizeExpression(ctx.SourceFile, caseExpr)

				if firstCase, exists := seenCases[normalized]; exists {
					ctx.ReportNode(caseExpr, rule.RuleMessage{
						Id:          "unexpected",
						Description: "Duplicate case label.",
					})
					_ = firstCase
				} else {
					seenCases[normalized] = caseExpr
				}
			}
		},
	}
}

// normalizeExpression converts an expression to a normalized string for comparison
// This removes whitespace and comments to detect duplicates
func normalizeExpression(sourceFile *ast.SourceFile, expr *ast.Node) string {
	if expr == nil {
		return ""
	}

	// Get the text of the expression
	textRange := utils.TrimNodeTextRange(sourceFile, expr)
	text := sourceFile.Text()[textRange.Pos():textRange.End()]

	// Normalize whitespace - replace all sequences of whitespace with single space
	// This helps match expressions that differ only in formatting
	normalized := strings.Join(strings.Fields(text), " ")

	return normalized
}
