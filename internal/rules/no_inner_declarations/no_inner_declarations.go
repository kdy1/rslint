package no_inner_declarations

import (
	"github.com/microsoft/typescript-go/shim/ast"

	"github.com/web-infra-dev/rslint/internal/rule"
)

// Options mirrors ESLint no-inner-declarations options
type Options struct {
	// Mode can be "functions" (default) or "both"
	Mode                  string `json:"mode"`
	BlockScopedFunctions  string `json:"blockScopedFunctions"` // "allow" or "disallow"
}

func parseOptions(options any) Options {
	opts := Options{
		Mode:                 "functions",
		BlockScopedFunctions: "allow",
	}

	if options == nil {
		return opts
	}

	// Parse options with dual-format support
	var optsData interface{}

	// Handle array format: ["both"] or ["both", { blockScopedFunctions: "allow" }]
	if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
		// First element is the mode string
		if mode, ok := optArray[0].(string); ok {
			opts.Mode = mode
		}
		// Second element (if present) is the options object
		if len(optArray) > 1 {
			optsData = optArray[1]
		}
	} else {
		// Handle direct string format: "both"
		if mode, ok := options.(string); ok {
			opts.Mode = mode
		} else {
			// Handle direct object format
			optsData = options
		}
	}

	// Parse object options
	if optsMap, ok := optsData.(map[string]interface{}); ok {
		if v, ok := optsMap["blockScopedFunctions"].(string); ok {
			opts.BlockScopedFunctions = v
		}
	}

	return opts
}

func buildMoveDeclToRootMessage(declType string, body string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "moveDeclToRoot",
		Description: "Move " + declType + " declaration to " + body + " root.",
	}
}

// NoInnerDeclarationsRule implements the no-inner-declarations rule
var NoInnerDeclarationsRule = rule.CreateRule(rule.Rule{
	Name: "no-inner-declarations",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// Simplified implementation - just check for function declarations in blocks
		opts := parseOptions(options)
		_ = opts // TODO: Use options for more sophisticated checking

		return rule.RuleListeners{
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				// Simplified check: report if function declaration is inside a block
				// This is a basic implementation that catches the most common cases
				parent := node.Parent
				if parent != nil && parent.Kind == ast.KindBlock {
					// Check if the block is not directly in a source file or function
					grandparent := parent.Parent
					if grandparent != nil {
						switch grandparent.Kind {
						case ast.KindSourceFile, ast.KindFunctionDeclaration,
							 ast.KindFunctionExpression, ast.KindArrowFunction,
							 ast.KindMethodDeclaration:
							// OK - these are allowed contexts
							return
						}
						// In any other context, it's an error
						ctx.ReportNode(node, buildMoveDeclToRootMessage("function", "program"))
					}
				}
			},
		}
	},
})
