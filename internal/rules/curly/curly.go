package curly

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Options for the curly rule
type Options struct {
	Mode       string // "all" (default), "multi", "multi-line", "multi-or-nest"
	Consistent bool   // requires consistent usage of braces in if-else chains
}

func parseOptions(options any) Options {
	opts := Options{
		Mode:       "all",
		Consistent: false,
	}

	if options == nil {
		return opts
	}

	// Handle array format: ["multi"] or ["multi-line", "consistent"]
	if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
		// First element is the mode
		if modeStr, isStr := optArray[0].(string); isStr {
			opts.Mode = modeStr
		}

		// Check for "consistent" option
		for i := 1; i < len(optArray); i++ {
			if str, ok := optArray[i].(string); ok && str == "consistent" {
				opts.Consistent = true
			}
		}
	} else if modeStr, ok := options.(string); ok {
		opts.Mode = modeStr
	}

	return opts
}

// Helper functions to determine if a statement is a block
func isBlock(stmt *ast.Node) bool {
	return stmt != nil && stmt.Kind == ast.KindBlock
}

// Helper to get node text
func getNodeText(sourceFile *ast.SourceFile, node *ast.Node) string {
	if node == nil {
		return ""
	}
	text := sourceFile.Text()
	rng := utils.TrimNodeTextRange(sourceFile, node)
	return text[rng.Pos():rng.End()]
}

// Check if statement is on a single line
func isSingleLine(sourceFile *ast.SourceFile, stmt *ast.Node) bool {
	if stmt == nil {
		return false
	}
	rng := utils.TrimNodeTextRange(sourceFile, stmt)
	// For now, just assume single line if the range is small
	// A proper implementation would need to count newlines
	return (rng.End() - rng.Pos()) < 80
}

// Check if statement contains nested control structures
func hasNestedControlStructure(stmt *ast.Node) bool {
	if stmt == nil || stmt.Kind == ast.KindBlock {
		return false
	}

	switch stmt.Kind {
	case ast.KindIfStatement, ast.KindWhileStatement, ast.KindDoStatement,
		ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement:
		return true
	}
	return false
}

// Build error messages
func buildMissingCurlyAfterMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingCurlyAfterCondition",
		Description: "Expected { after condition.",
	}
}

func buildMissingCurlyMessage(keyword string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingCurlyAfter",
		Description: "Expected { after '" + keyword + "'.",
	}
}

func buildUnexpectedCurlyAfterMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedCurlyAfterCondition",
		Description: "Unnecessary { after condition.",
	}
}

func buildUnexpectedCurlyMessage(keyword string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedCurlyAfter",
		Description: "Unnecessary { after '" + keyword + "'.",
	}
}

// Check a statement body
func checkBody(ctx rule.RuleContext, opts Options, stmt *ast.Node, body *ast.Node, keyword string, isCondition bool) {
	if body == nil {
		return
	}

	hasBlock := isBlock(body)

	// Mode "all": always require braces
	if opts.Mode == "all" {
		if !hasBlock {
			msg := buildMissingCurlyMessage(keyword)
			if isCondition {
				msg = buildMissingCurlyAfterMessage()
			}
			ctx.ReportNode(body, msg)
		}
		return
	}

	// Mode "multi": allow braceless for single statements
	if opts.Mode == "multi" {
		if hasBlock {
			// Check if block contains only one statement
			block := body.AsBlock()
			if block != nil && block.Statements != nil && len(block.Statements.Nodes) == 1 {
				msg := buildUnexpectedCurlyMessage(keyword)
				if isCondition {
					msg = buildUnexpectedCurlyAfterMessage()
				}
				ctx.ReportNode(body, msg)
			}
		}
		return
	}

	// Mode "multi-line": require braces only for multi-line bodies
	if opts.Mode == "multi-line" {
		isSingle := isSingleLine(ctx.SourceFile, body)
		if !hasBlock && !isSingle {
			msg := buildMissingCurlyMessage(keyword)
			if isCondition {
				msg = buildMissingCurlyAfterMessage()
			}
			ctx.ReportNode(body, msg)
		} else if hasBlock && isSingle {
			block := body.AsBlock()
			if block != nil && block.Statements != nil && len(block.Statements.Nodes) == 1 {
				msg := buildUnexpectedCurlyMessage(keyword)
				if isCondition {
					msg = buildUnexpectedCurlyAfterMessage()
				}
				ctx.ReportNode(body, msg)
			}
		}
		return
	}

	// Mode "multi-or-nest": allow braceless unless nested
	if opts.Mode == "multi-or-nest" {
		if !hasBlock && hasNestedControlStructure(body) {
			msg := buildMissingCurlyMessage(keyword)
			if isCondition {
				msg = buildMissingCurlyAfterMessage()
			}
			ctx.ReportNode(body, msg)
		} else if hasBlock && !hasNestedControlStructure(body) {
			block := body.AsBlock()
			if block != nil && block.Statements != nil && len(block.Statements.Nodes) == 1 {
				msg := buildUnexpectedCurlyMessage(keyword)
				if isCondition {
					msg = buildUnexpectedCurlyAfterMessage()
				}
				stmt := block.Statements.Nodes[0]
				if !hasNestedControlStructure(stmt) {
					ctx.ReportNode(body, msg)
				}
			}
		}
		return
	}
}

var CurlyRule = rule.CreateRule(rule.Rule{
	Name: "curly",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		listeners := rule.RuleListeners{}

		// Handle if statements
		listeners[ast.KindIfStatement] = func(node *ast.Node) {
			ifStmt := node.AsIfStatement()
			if ifStmt == nil {
				return
			}

			// Check then statement
			checkBody(ctx, opts, node, ifStmt.ThenStatement, "if", true)

			// Check else statement
			if ifStmt.ElseStatement != nil {
				// If else is another if statement, skip (it will be handled by its own listener)
				if ifStmt.ElseStatement.Kind != ast.KindIfStatement {
					checkBody(ctx, opts, node, ifStmt.ElseStatement, "else", false)
				}

				// Handle consistent option
				if opts.Consistent {
					thenHasBlock := isBlock(ifStmt.ThenStatement)
					elseHasBlock := isBlock(ifStmt.ElseStatement)
					if thenHasBlock != elseHasBlock {
						// They should be consistent
						// Report on whichever one doesn't have blocks
						if !thenHasBlock {
							msg := buildMissingCurlyAfterMessage()
							ctx.ReportNode(ifStmt.ThenStatement, msg)
						} else if !elseHasBlock && ifStmt.ElseStatement.Kind != ast.KindIfStatement {
							msg := buildMissingCurlyMessage("else")
							ctx.ReportNode(ifStmt.ElseStatement, msg)
						}
					}
				}
			}
		}

		// Handle while statements
		listeners[ast.KindWhileStatement] = func(node *ast.Node) {
			whileStmt := node.AsWhileStatement()
			if whileStmt == nil {
				return
			}
			checkBody(ctx, opts, node, whileStmt.Statement, "while", true)
		}

		// Handle do-while statements
		listeners[ast.KindDoStatement] = func(node *ast.Node) {
			doStmt := node.AsDoStatement()
			if doStmt == nil {
				return
			}
			checkBody(ctx, opts, node, doStmt.Statement, "do", false)
		}

		// Handle for statements
		listeners[ast.KindForStatement] = func(node *ast.Node) {
			forStmt := node.AsForStatement()
			if forStmt == nil {
				return
			}
			checkBody(ctx, opts, node, forStmt.Statement, "for", true)
		}

		// TODO: Add ForInStatement and ForOfStatement when the AST methods are available

		return listeners
	},
})
