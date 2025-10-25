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
		opts := parseOptions(options)

		listeners := rule.RuleListeners{}

		// Always check for function declarations
		listeners[ast.KindFunctionDeclaration] = func(node *ast.Node) {
			// Check if function declaration is inside a block
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
		}

		// When mode is "both", also check for var declarations
		if opts.Mode == "both" {
			listeners[ast.KindVariableStatement] = func(node *ast.Node) {
				// Check if this is a var declaration
				if varDecl, ok := node.AsVariableStatement(); ok {
					if (varDecl.DeclarationList.Flags & uint32(ast.NodeFlagsLet|ast.NodeFlagsConst)) == 0 {
						// It's a var declaration (not let/const)
						// Check if it's inside a nested block
						parent := node.Parent
						for parent != nil {
							if parent.Kind == ast.KindBlock {
								// Check the context of the block
								grandparent := parent.Parent
								if grandparent != nil {
									switch grandparent.Kind {
									case ast.KindSourceFile:
										// OK - at program root
										return
									case ast.KindFunctionDeclaration, ast.KindFunctionExpression,
										ast.KindArrowFunction, ast.KindMethodDeclaration,
										ast.KindConstructor:
										// Check if this is the immediate function body
										// If we're directly in the function's block, it's OK
										// If we're in a nested block, it's an error
										funcParent := grandparent
										varAncestor := node.Parent

										// Walk up from the var declaration to see if we hit the function
										// before hitting another block
										for varAncestor != nil && varAncestor != funcParent {
											if varAncestor.Kind == ast.KindBlock && varAncestor.Parent != nil {
												blockContext := varAncestor.Parent
												switch blockContext.Kind {
												case ast.KindIfStatement, ast.KindForStatement,
													ast.KindWhileStatement, ast.KindDoStatement,
													ast.KindSwitchStatement, ast.KindWithStatement,
													ast.KindForInStatement, ast.KindForOfStatement,
													ast.KindBlock:
													// Found a nested block - this is an error
													ctx.ReportNode(node, buildMoveDeclToRootMessage("variable", "function"))
													return
												}
											}
											varAncestor = varAncestor.Parent
										}
										return
									default:
										// In other contexts like if/for/while, report error
										ctx.ReportNode(node, buildMoveDeclToRootMessage("variable", "program"))
										return
									}
								}
							}
							parent = parent.Parent
						}
					}
				}
			}
		}

		return listeners
	},
})
