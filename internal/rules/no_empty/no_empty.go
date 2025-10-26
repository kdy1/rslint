package no_empty

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoEmptyOptions defines the configuration options for this rule
type NoEmptyOptions struct {
	AllowEmptyCatch bool `json:"allowEmptyCatch"`
}

// parseOptions parses and validates the rule options
func parseOptions(options any) NoEmptyOptions {
	opts := NoEmptyOptions{
		AllowEmptyCatch: false, // Default value
	}

	if options == nil {
		return opts
	}

	// Handle both array format [{ option: value }] and object format { option: value }
	var optsMap map[string]interface{}
	if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
		optsMap, _ = optArray[0].(map[string]interface{})
	} else {
		optsMap, _ = options.(map[string]interface{})
	}

	if optsMap != nil {
		if v, ok := optsMap["allowEmptyCatch"].(bool); ok {
			opts.AllowEmptyCatch = v
		}
	}

	return opts
}

func buildUnexpectedMessage(blockType string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpected",
		Description: "Empty " + blockType + " statement.",
	}
}

func buildSuggestionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "suggestComment",
		Description: "Add comment inside empty block",
	}
}

// NoEmptyRule implements the no-empty rule
// Disallow empty block statements
var NoEmptyRule = rule.CreateRule(rule.Rule{
	Name: "no-empty",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		// Helper function to check if a block has comments
		hasComments := func(block *ast.Node) bool {
			if block == nil {
				return false
			}

			// Check for leading trivia (comments)
			fullText := ctx.SourceFile.Text()
			start := block.Pos()
			end := block.End()

			// Look for comments within the block
			blockText := fullText[start:end]
			// Check if there are any comment markers in the block
			for i := 0; i < len(blockText)-1; i++ {
				if blockText[i] == '/' && (blockText[i+1] == '/' || blockText[i+1] == '*') {
					return true
				}
			}

			return false
		}

		// Helper function to check if a block is empty
		isEmptyBlock := func(block *ast.Node) bool {
			if block == nil || block.Kind != ast.KindBlock {
				return false
			}

			statements := block.AsBlock().Statements
			if statements == nil || len(statements.Nodes) == 0 {
				return !hasComments(block)
			}

			return false
		}

		return rule.RuleListeners{
			ast.KindBlock: func(node *ast.Node) {
				if !isEmptyBlock(node) {
					return
				}

				// Check if this is a catch clause block and allowEmptyCatch is enabled
				parent := node.Parent()
				if parent != nil && parent.Kind == ast.KindCatchClause && opts.AllowEmptyCatch {
					return
				}

				// Determine the block type for the error message
				blockType := "block"
				if parent != nil {
					switch parent.Kind {
					case ast.KindIfStatement:
						blockType = "if"
					case ast.KindWhileStatement:
						blockType = "while"
					case ast.KindDoStatement:
						blockType = "do...while"
					case ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement:
						blockType = "for"
					case ast.KindSwitchStatement:
						blockType = "switch"
					case ast.KindTryStatement:
						blockType = "try"
					case ast.KindCatchClause:
						blockType = "catch"
					case ast.KindFinallyClause:
						blockType = "finally"
					}
				}

				// Report with suggestion to add a comment
				ctx.ReportNodeWithSuggestions(node, buildUnexpectedMessage(blockType),
					rule.RuleSuggestion{
						Message: buildSuggestionMessage(),
						FixesArr: []rule.RuleFix{
							rule.RuleFixInsertTextAfterRange(
								rule.TextRange{Pos: node.Pos() + 1, End: node.Pos() + 1},
								"/* empty */",
							),
						},
					},
				)
			},
		}
	},
})
