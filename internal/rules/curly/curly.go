package curly

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// CurlyOptions defines the configuration options for this rule
type CurlyOptions struct {
	Mode       string // "all", "multi", "multi-line", "multi-or-nest"
	Consistent bool   // whether to require consistent bracing in if-else chains
}

// parseOptions parses and validates the rule options
func parseOptions(options any) CurlyOptions {
	opts := CurlyOptions{
		Mode:       "all",
		Consistent: false,
	}

	if options == nil {
		return opts
	}

	// Handle array format: ["multi", "consistent"] or [{ ... }]
	if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
		// First element can be a string mode or object
		if modeStr, ok := optArray[0].(string); ok {
			opts.Mode = modeStr
		} else if optsMap, ok := optArray[0].(map[string]interface{}); ok {
			if v, ok := optsMap["mode"].(string); ok {
				opts.Mode = v
			}
			if v, ok := optsMap["consistent"].(bool); ok {
				opts.Consistent = v
			}
		}
		// Check for "consistent" as second argument
		if len(optArray) > 1 {
			if consistentStr, ok := optArray[1].(string); ok && consistentStr == "consistent" {
				opts.Consistent = true
			}
		}
	} else if optsMap, ok := options.(map[string]interface{}); ok {
		if v, ok := optsMap["mode"].(string); ok {
			opts.Mode = v
		}
		if v, ok := optsMap["consistent"].(bool); ok {
			opts.Consistent = v
		}
	}

	return opts
}

func buildMissingCurlyMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingCurlyAfter",
		Description: fmt.Sprintf("Expected { after '%s' condition.", name),
	}
}

func buildMissingCurlyAfterElseMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingCurlyAfterElse",
		Description: "Expected { after 'else'.",
	}
}

func buildUnexpectedCurlyMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedCurlyAfter",
		Description: fmt.Sprintf("Unnecessary { after '%s' condition.", name),
	}
}

func buildUnexpectedCurlyAfterElseMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedCurlyAfterElse",
		Description: "Unnecessary { after 'else'.",
	}
}

func buildInconsistentMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "inconsistentCurly",
		Description: "Expected { after condition in if-else chain; all clauses should be wrapped in braces.",
	}
}

// CurlyRule implements the curly rule
// Enforce consistent brace style for all control statements
var CurlyRule = rule.Rule{
	Name: "curly",
	Run:  run,
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := parseOptions(options)

	checkStatement := func(stmt *ast.Node, stmtName string) {
		if stmt == nil {
			return
		}

		hasBlock := stmt.Kind == ast.KindBlock
		needsBraces := shouldHaveBraces(stmt, opts.Mode)

		if needsBraces && !hasBlock {
			// Missing braces
			msg := buildMissingCurlyMessage(stmtName)
			fix := wrapWithBraces(ctx, stmt)
			ctx.ReportNodeWithFixes(stmt, msg, fix)
		} else if !needsBraces && hasBlock {
			// Unnecessary braces (only for multi mode)
			if opts.Mode == "multi" {
				block := stmt.AsBlock()
				if block != nil && isSingleStatement(block) {
					msg := buildUnexpectedCurlyMessage(stmtName)
					fix := removeBraces(ctx, stmt, block)
					if fix.Edits != nil {
						ctx.ReportNodeWithFixes(stmt, msg, fix)
					}
				}
			}
		}
	}

	return rule.RuleListeners{
		ast.KindIfStatement: func(node *ast.Node) {
			ifStmt := node.AsIfStatement()
			if ifStmt == nil {
				return
			}

			checkStatement(ifStmt.ThenStatement, "if")

			// Check else clause
			if ifStmt.ElseStatement != nil {
				elseStmt := ifStmt.ElseStatement
				// If the else clause is another if statement, skip it
				if elseStmt.Kind != ast.KindIfStatement {
					hasBlock := elseStmt.Kind == ast.KindBlock
					needsBraces := shouldHaveBraces(elseStmt, opts.Mode)

					if needsBraces && !hasBlock {
						msg := buildMissingCurlyAfterElseMessage()
						fix := wrapWithBraces(ctx, elseStmt)
						ctx.ReportNodeWithFixes(elseStmt, msg, fix)
					} else if !needsBraces && hasBlock {
						if opts.Mode == "multi" {
							block := elseStmt.AsBlock()
							if block != nil && isSingleStatement(block) {
								msg := buildUnexpectedCurlyAfterElseMessage()
								fix := removeBraces(ctx, elseStmt, block)
								if fix.Edits != nil {
									ctx.ReportNodeWithFixes(elseStmt, msg, fix)
								}
							}
						}
					}
				}
			}

			// Check consistent option for if-else chains
			if opts.Consistent {
				checkConsistentBraces(ctx, node)
			}
		},
		ast.KindWhileStatement: func(node *ast.Node) {
			whileStmt := node.AsWhileStatement()
			if whileStmt == nil {
				return
			}
			checkStatement(whileStmt.Statement, "while")
		},
		ast.KindDoStatement: func(node *ast.Node) {
			doStmt := node.AsDoStatement()
			if doStmt == nil {
				return
			}
			checkStatement(doStmt.Statement, "do")
		},
		ast.KindForStatement: func(node *ast.Node) {
			forStmt := node.AsForStatement()
			if forStmt == nil {
				return
			}
			checkStatement(forStmt.Statement, "for")
		},
		ast.KindForInStatement: func(node *ast.Node) {
			forInStmt := node.AsForInStatement()
			if forInStmt == nil {
				return
			}
			checkStatement(forInStmt.Statement, "for-in")
		},
		ast.KindForOfStatement: func(node *ast.Node) {
			forOfStmt := node.AsForOfStatement()
			if forOfStmt == nil {
				return
			}
			checkStatement(forOfStmt.Statement, "for-of")
		},
	}
}

func shouldHaveBraces(stmt *ast.Node, mode string) bool {
	if stmt == nil {
		return false
	}

	switch mode {
	case "all":
		return true
	case "multi":
		// Needs braces if it's a block with multiple statements
		if stmt.Kind == ast.KindBlock {
			block := stmt.AsBlock()
			return block != nil && !isSingleStatement(block)
		}
		return false
	case "multi-line":
		// Needs braces if statement spans multiple lines
		return isMultiLine(stmt)
	case "multi-or-nest":
		// Needs braces if multi-line or contains nested control structures
		if isMultiLine(stmt) {
			return true
		}
		return hasNestedControlStructure(stmt)
	default:
		return true
	}
}

func isSingleStatement(block *ast.Block) bool {
	if block == nil || block.Statements == nil {
		return false
	}
	return len(*block.Statements) == 1
}

func isMultiLine(node *ast.Node) bool {
	if node == nil {
		return false
	}
	startLine := utils.GetStartLine(node)
	endLine := utils.GetEndLine(node)
	return endLine > startLine
}

func hasNestedControlStructure(node *ast.Node) bool {
	if node == nil {
		return false
	}
	kind := node.Kind
	return kind == ast.KindIfStatement ||
		kind == ast.KindWhileStatement ||
		kind == ast.KindDoStatement ||
		kind == ast.KindForStatement ||
		kind == ast.KindForInStatement ||
		kind == ast.KindForOfStatement
}

func wrapWithBraces(ctx rule.RuleContext, stmt *ast.Node) rule.RuleFix {
	stmtText := utils.GetNodeText(stmt)
	wrapped := fmt.Sprintf("{ %s }", stmtText)
	return rule.RuleFix{
		Message: "Add braces around statement",
		Edits: []rule.TextEdit{
			rule.RuleFixReplace(ctx.SourceFile, stmt, wrapped),
		},
	}
}

func removeBraces(ctx rule.RuleContext, blockNode *ast.Node, block *ast.Block) rule.RuleFix {
	if block == nil || block.Statements == nil || len(*block.Statements) != 1 {
		return rule.RuleFix{}
	}

	innerStmt := (*block.Statements)[0]
	innerText := utils.GetNodeText(&innerStmt)

	return rule.RuleFix{
		Message: "Remove unnecessary braces",
		Edits: []rule.TextEdit{
			rule.RuleFixReplace(ctx.SourceFile, blockNode, innerText),
		},
	}
}

func checkConsistentBraces(ctx rule.RuleContext, ifNode *ast.Node) {
	ifStmt := ifNode.AsIfStatement()
	if ifStmt == nil || ifStmt.ElseStatement == nil {
		return
	}

	thenHasBraces := ifStmt.ThenStatement != nil && ifStmt.ThenStatement.Kind == ast.KindBlock
	elseHasBraces := ifStmt.ElseStatement.Kind == ast.KindBlock

	// If else is another if statement, recurse
	if ifStmt.ElseStatement.Kind == ast.KindIfStatement {
		return
	}

	// Check for inconsistency
	if thenHasBraces != elseHasBraces {
		// Report on whichever doesn't have braces
		if !thenHasBraces {
			ctx.ReportNodeWithFixes(ifStmt.ThenStatement, buildInconsistentMessage(),
				wrapWithBraces(ctx, ifStmt.ThenStatement))
		} else {
			ctx.ReportNodeWithFixes(ifStmt.ElseStatement, buildInconsistentMessage(),
				wrapWithBraces(ctx, ifStmt.ElseStatement))
		}
	}
}
