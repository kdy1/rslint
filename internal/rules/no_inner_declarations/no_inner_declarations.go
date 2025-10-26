package no_inner_declarations

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoInnerDeclarationsOptions defines the configuration options for this rule
type NoInnerDeclarationsOptions struct {
	// Mode can be "functions" (default) or "both"
	Mode string
	// BlockScopedFunctions can be "allow" (default) or "disallow"
	BlockScopedFunctions string
}

// parseOptions parses and validates the rule options
func parseOptions(options any) NoInnerDeclarationsOptions {
	opts := NoInnerDeclarationsOptions{
		Mode:                 "functions", // default: check only functions
		BlockScopedFunctions: "allow",     // default: allow block-scoped functions in strict mode
	}

	if options == nil {
		return opts
	}

	// Handle array format: ["both"] or ["both", { blockScopedFunctions: "disallow" }]
	if optArray, isArray := options.([]interface{}); isArray {
		if len(optArray) > 0 {
			// First element is the mode string
			if mode, ok := optArray[0].(string); ok {
				opts.Mode = mode
			}
		}
		if len(optArray) > 1 {
			// Second element is an options object
			if optsMap, ok := optArray[1].(map[string]interface{}); ok {
				if v, ok := optsMap["blockScopedFunctions"].(string); ok {
					opts.BlockScopedFunctions = v
				}
			}
		}
		return opts
	}

	// Handle direct object format (shouldn't happen for this rule, but handle it anyway)
	if optsMap, ok := options.(map[string]interface{}); ok {
		if v, ok := optsMap["mode"].(string); ok {
			opts.Mode = v
		}
		if v, ok := optsMap["blockScopedFunctions"].(string); ok {
			opts.BlockScopedFunctions = v
		}
	}

	return opts
}

// isValidParent checks if the parent node is a valid context for declarations
func isValidParent(parent *ast.Node) bool {
	if parent == nil {
		return false
	}

	switch parent.Kind {
	case ast.KindSourceFile:
		return true
	case ast.KindClassStaticBlockDeclaration:
		return true
	case ast.KindExportDeclaration:
		return true
	case ast.KindExportAssignment:
		return true
	default:
		return false
	}
}

// isFunctionBody checks if the parent is a BlockStatement that is the body of a function
func isFunctionBody(parent *ast.Node) bool {
	if parent == nil || parent.Kind != ast.KindBlock {
		return false
	}

	grandparent := parent.Parent
	if grandparent == nil {
		return false
	}

	switch grandparent.Kind {
	case ast.KindFunctionDeclaration,
		ast.KindFunctionExpression,
		ast.KindArrowFunction,
		ast.KindMethodDeclaration,
		ast.KindConstructor,
		ast.KindGetAccessor,
		ast.KindSetAccessor:
		return true
	default:
		return false
	}
}

// getBodyDescription returns a description of the nearest valid enclosing context
func getBodyDescription(node *ast.Node) string {
	current := node.Parent

	for current != nil {
		switch current.Kind {
		case ast.KindSourceFile:
			return "program"
		case ast.KindFunctionDeclaration,
			ast.KindFunctionExpression,
			ast.KindArrowFunction,
			ast.KindMethodDeclaration,
			ast.KindConstructor,
			ast.KindGetAccessor,
			ast.KindSetAccessor:
			return "function body"
		case ast.KindClassStaticBlockDeclaration:
			return "class static block body"
		}
		current = current.Parent
	}

	return "program"
}

// isInStrictMode checks if a node is in strict mode
func isInStrictMode(node *ast.Node) bool {
	// Find the nearest enclosing function or source file
	current := node
	for current != nil {
		// Check if this is a source file or function with strict mode
		var statementsToCheck []*ast.Node

		switch current.Kind {
		case ast.KindSourceFile:
			statementsToCheck = current.Statements()
		case ast.KindFunctionDeclaration,
			ast.KindFunctionExpression,
			ast.KindArrowFunction,
			ast.KindMethodDeclaration,
			ast.KindConstructor:
			body := current.Body()
			if body != nil && body.Kind == ast.KindBlock {
				statementsToCheck = body.Statements()
			}
		}

		// Check for "use strict" directive
		if len(statementsToCheck) > 0 {
			for _, stmt := range statementsToCheck {
				if stmt == nil {
					continue
				}
				// Check if this is an expression statement with a string literal
				if stmt.Kind == ast.KindExpressionStatement {
					expr := stmt.Expression()
					if expr != nil && expr.Kind == ast.KindStringLiteral {
						text := expr.Text()
						if text == "'use strict'" || text == "\"use strict\"" {
							return true
						}
					}
				}
				// If we hit a non-directive statement, stop looking in this scope
				if stmt.Kind != ast.KindExpressionStatement {
					break
				}
			}
		}

		// Module files are automatically strict
		if current.Kind == ast.KindSourceFile {
			// Check if this is a module (has import/export statements)
			statements := current.Statements()
			for _, stmt := range statements {
				if stmt == nil {
					continue
				}
				if stmt.Kind == ast.KindImportDeclaration || stmt.Kind == ast.KindExportDeclaration || stmt.Kind == ast.KindExportAssignment {
					return true
				}
			}
			break
		}

		current = current.Parent
	}

	return false
}

// checkDeclaration checks if a declaration is in a valid position
func checkDeclaration(ctx rule.RuleContext, node *ast.Node, declType string, opts NoInnerDeclarationsOptions) {
	if node == nil {
		return
	}

	parent := node.Parent
	if parent == nil {
		return
	}

	// Check if parent is a valid location for declarations
	if isValidParent(parent) {
		return
	}

	// Check if parent is a function body (BlockStatement inside a function)
	if isFunctionBody(parent) {
		return
	}

	// Special handling for function declarations with blockScopedFunctions option
	if declType == "function" && opts.BlockScopedFunctions == "allow" {
		// In ES2015+ strict mode, block-scoped function declarations are allowed
		// Check if we're in strict mode
		if isInStrictMode(node) {
			// In strict mode with ES2015+, function declarations in blocks are block-scoped
			// We allow this when blockScopedFunctions is "allow"
			// However, we need to check the ECMAScript version
			// For now, we'll assume ES2015+ and allow it in strict mode
			return
		}
	}

	// If we reach here, the declaration is in an invalid location
	bodyDesc := getBodyDescription(node)

	ctx.ReportNode(node, rule.RuleMessage{
		Id:          "moveDeclToRoot",
		Description: "Move " + declType + " declaration to " + bodyDesc + " root.",
	})
}

// NoInnerDeclarationsRule implements the no-inner-declarations rule
// Disallow variable or `function` declarations in nested blocks
var NoInnerDeclarationsRule = rule.Rule{
	Name: "no-inner-declarations",
	Run:  run,
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := parseOptions(options)

	return rule.RuleListeners{
		ast.KindFunctionDeclaration: func(node *ast.Node) {
			checkDeclaration(ctx, node, "function", opts)
		},
		ast.KindVariableStatement: func(node *ast.Node) {
			// Only check variable declarations if mode is "both"
			if opts.Mode != "both" {
				return
			}

			// Only check 'var' declarations (not let/const/using)
			// VariableStatement contains a VariableDeclarationList
			declList := node.DeclarationList()
			if declList == nil {
				return
			}

			// Check the flags to see if this is a 'var' declaration
			// In TypeScript AST, the flags tell us the declaration kind
			// We need to check if this is a 'var' (not 'let', 'const', or other)
			flags := declList.Flags
			isVar := (flags&ast.FlagLet) == 0 && (flags&ast.FlagConst) == 0

			if isVar {
				checkDeclaration(ctx, node, "variable", opts)
			}
		},
	}
}
