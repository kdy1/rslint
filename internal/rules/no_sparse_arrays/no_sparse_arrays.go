package no_sparse_arrays

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoSparseArraysRule implements the no-sparse-arrays rule
// Disallow sparse arrays
var NoSparseArraysRule = rule.Rule{
	Name: "no-sparse-arrays",
	Run:  run,
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	return rule.RuleListeners{
		ast.KindArrayLiteralExpression: func(node *ast.Node) {
			arrayLit := node.AsArrayLiteralExpression()
			if arrayLit == nil || arrayLit.Elements == nil {
				return
			}

			// Check each element for missing/sparse positions
			for _, elem := range arrayLit.Elements.Nodes {
				if elem == nil {
					// This is a sparse/missing element
					continue
				}
				// Check if this is an OmittedExpression node
				if elem.Kind == ast.KindOmittedExpression {
					ctx.ReportNode(elem, rule.RuleMessage{
						Id:          "unexpectedSparseArray",
						Description: "Unexpected comma in middle of array.",
					})
				}
			}
		},
	}
}
