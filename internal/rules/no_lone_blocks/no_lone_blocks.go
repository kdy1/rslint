package no_lone_blocks

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoLoneBlocksRule implements the no-lone-blocks rule
// Disallow unnecessary nested blocks
var NoLoneBlocksRule = rule.CreateRule(rule.Rule{
	Name: "no-lone-blocks",
	Run:  run,
})

func buildRedundantBlockMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "redundantBlock",
		Description: "Block is redundant.",
	}
}

func buildRedundantNestedBlockMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "redundantNestedBlock",
		Description: "Nested block is redundant.",
	}
}

// isLoneBlock checks if a block statement is a "lone block" that might be redundant
func isLoneBlock(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindBlock {
		return false
	}

	parent := node.Parent()
	if parent == nil {
		return false
	}

	switch parent.Kind {
	case ast.KindBlock, ast.KindSourceFile:
		return true
	case ast.KindCaseClause, ast.KindDefaultClause:
		// Check if this is the only statement in the case
		statements := parent.Statements()
		return statements != nil && len(statements.Nodes) > 1
	}

	return false
}

// hasBlockLevelBindings checks if a block contains let/const/class declarations or functions in strict mode
func hasBlockLevelBindings(node *ast.Node, isStrict bool) bool {
	if node == nil || node.Kind != ast.KindBlock {
		return false
	}

	statements := node.Statements()
	if statements == nil {
		return false
	}

	for _, stmt := range statements.Nodes {
		if stmt == nil {
			continue
		}

		switch stmt.Kind {
		case ast.KindVariableStatement:
			// Check if it's let or const
			declList := stmt.DeclarationList()
			if declList != nil {
				flags := declList.Flags()
				// NodeFlags: Let = 1, Const = 2
				if flags&1 != 0 || flags&2 != 0 {
					return true
				}
			}
		case ast.KindClassDeclaration:
			return true
		case ast.KindFunctionDeclaration:
			// Function declarations create block-level bindings in strict mode
			if isStrict {
				return true
			}
		}
	}

	return false
}

// isInStaticBlock checks if a node is within a static class block
func isInStaticBlock(node *ast.Node) bool {
	current := node.Parent()
	for current != nil {
		if current.Kind == ast.KindClassStaticBlockDeclaration {
			return true
		}
		current = current.Parent()
	}
	return false
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	// Track blocks that might be redundant
	var blockStack []*ast.Node

	return rule.RuleListeners{
		ast.KindBlock: func(node *ast.Node) {
			if node == nil {
				return
			}

			// Check if this is a lone block
			if !isLoneBlock(node) {
				return
			}

			// In static blocks, only report if it's nested
			if isInStaticBlock(node) {
				parent := node.Parent()
				if parent != nil && parent.Kind == ast.KindClassStaticBlockDeclaration {
					return
				}
			}

			// Add to stack for later checking
			blockStack = append(blockStack, node)
		},

		// Check for block-level bindings
		ast.KindVariableStatement: func(node *ast.Node) {
			if node == nil || len(blockStack) == 0 {
				return
			}

			// Check if this is a let or const declaration
			declList := node.DeclarationList()
			if declList != nil {
				flags := declList.Flags()
				// NodeFlags: Let = 1, Const = 2
				if flags&1 != 0 || flags&2 != 0 {
					// Remove the containing block from the stack
					for i := len(blockStack) - 1; i >= 0; i-- {
						block := blockStack[i]
						if nodeContains(block, node) {
							blockStack = append(blockStack[:i], blockStack[i+1:]...)
							break
						}
					}
				}
			}
		},

		ast.KindClassDeclaration: func(node *ast.Node) {
			if node == nil || len(blockStack) == 0 {
				return
			}

			// Remove the containing block from the stack
			for i := len(blockStack) - 1; i >= 0; i-- {
				block := blockStack[i]
				if nodeContains(block, node) {
					blockStack = append(blockStack[:i], blockStack[i+1:]...)
					break
				}
			}
		},

		ast.KindFunctionDeclaration: func(node *ast.Node) {
			if node == nil || len(blockStack) == 0 {
				return
			}

			// Check if we're in strict mode
			// For now, assume strict mode for function declarations in blocks
			isStrict := true // This could be enhanced to check actual strict mode

			if isStrict {
				// Remove the containing block from the stack
				for i := len(blockStack) - 1; i >= 0; i-- {
					block := blockStack[i]
					if nodeContains(block, node) {
						blockStack = append(blockStack[:i], blockStack[i+1:]...)
						break
					}
				}
			}
		},

		// Report at the end of file processing
		ast.KindEndOfFileToken: func(node *ast.Node) {
			// Report all remaining blocks in the stack
			for _, block := range blockStack {
				parent := block.Parent()
				if parent != nil && (parent.Kind == ast.KindBlock || parent.Kind == ast.KindClassStaticBlockDeclaration) {
					ctx.ReportNode(block, buildRedundantNestedBlockMessage())
				} else {
					ctx.ReportNode(block, buildRedundantBlockMessage())
				}
			}
			blockStack = nil
		},
	}
}

// nodeContains checks if parent node contains child node
func nodeContains(parent, child *ast.Node) bool {
	if parent == nil || child == nil {
		return false
	}

	current := child.Parent()
	for current != nil {
		if current == parent {
			return true
		}
		current = current.Parent()
	}
	return false
}
