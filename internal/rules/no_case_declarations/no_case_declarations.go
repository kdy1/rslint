package no_case_declarations

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoCaseDeclarationsRule implements the no-case-declarations rule
// Disallow lexical declarations in case clauses
var NoCaseDeclarationsRule = rule.CreateRule(rule.Rule{
	Name: "no-case-declarations",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	checkCaseOrDefaultClause := func(node *ast.Node) {
		clause := node.AsCaseOrDefaultClause()
		if clause == nil || clause.Statements == nil {
			return
		}

		// Check each statement
		for _, stmt := range clause.Statements.Nodes {
			if stmt == nil {
				continue
			}

			// Check for lexical declarations that are NOT wrapped in a block
			shouldReport := false

			switch stmt.Kind {
			case ast.KindVariableStatement:
				shouldReport = true
			case ast.KindFunctionDeclaration:
				shouldReport = true
			case ast.KindClassDeclaration:
				shouldReport = true
			}

			if shouldReport {
				ctx.ReportNode(stmt, rule.RuleMessage{
					Id:          "unexpected",
					Description: "Unexpected lexical declaration in case block.",
				})
			}
		}
	}

	return rule.RuleListeners{
		ast.KindCaseClause:    checkCaseOrDefaultClause,
		ast.KindDefaultClause: checkCaseOrDefaultClause,
	}
}
