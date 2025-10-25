package prefer_destructuring

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// PreferDestructuringOptions defines the configuration options for this rule
type PreferDestructuringOptions struct {
	ArrayDestructuring            bool `json:"array"`
	ObjectDestructuring           bool `json:"object"`
	EnforceForRenamedProperties   bool `json:"enforceForRenamedProperties"`
}

// parseOptions parses and validates the rule options
func parseOptions(options any) PreferDestructuringOptions {
	// Default: both array and object destructuring enabled
	opts := PreferDestructuringOptions{
		ArrayDestructuring:          true,
		ObjectDestructuring:         true,
		EnforceForRenamedProperties: false,
	}

	if options == nil {
		return opts
	}

	// Handle both array format and object format
	var optsMap map[string]interface{}
	if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
		optsMap, _ = optArray[0].(map[string]interface{})
	} else {
		optsMap, _ = options.(map[string]interface{})
	}

	if optsMap != nil {
		// Check for nested VariableDeclarator config
		if varDecl, ok := optsMap["VariableDeclarator"].(map[string]interface{}); ok {
			if v, ok := varDecl["array"].(bool); ok {
				opts.ArrayDestructuring = v
			}
			if v, ok := varDecl["object"].(bool); ok {
				opts.ObjectDestructuring = v
			}
		}

		// Direct settings
		if v, ok := optsMap["array"].(bool); ok {
			opts.ArrayDestructuring = v
		}
		if v, ok := optsMap["object"].(bool); ok {
			opts.ObjectDestructuring = v
		}
		if v, ok := optsMap["enforceForRenamedProperties"].(bool); ok {
			opts.EnforceForRenamedProperties = v
		}
	}

	return opts
}

func buildArrayMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferDestructuring",
		Description: "Use array destructuring.",
	}
}

func buildObjectMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferDestructuring",
		Description: "Use object destructuring.",
	}
}

// getNodeText returns the trimmed text of a node
func getNodeText(srcFile *ast.SourceFile, node *ast.Node) string {
	rng := utils.TrimNodeTextRange(srcFile, node)
	return srcFile.Text()[rng.Pos():rng.End()]
}

// isOptionalChaining checks if the expression uses optional chaining
func isOptionalChaining(node *ast.Node) bool {
	if node == nil {
		return false
	}
	// Check for PropertyAccessExpression with question dot token
	if node.Kind == ast.KindPropertyAccessExpression {
		if propAccess := node.AsPropertyAccessExpression(); propAccess != nil {
			return propAccess.QuestionDotToken != nil
		}
	}
	// Check for ElementAccessExpression with question dot token
	if node.Kind == ast.KindElementAccessExpression {
		if elemAccess := node.AsElementAccessExpression(); elemAccess != nil {
			return elemAccess.QuestionDotToken != nil
		}
	}
	return false
}

// PreferDestructuringRule implements the prefer-destructuring rule
// Require destructuring from arrays/objects
var PreferDestructuringRule = rule.Rule{
	Name: "prefer-destructuring",
	Run:  run,
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := parseOptions(options)

	checkVariableDeclarator := func(node *ast.Node) {
		varDecl := node.AsVariableDeclaration()
		if varDecl == nil || varDecl.Initializer == nil || varDecl.Name == nil {
			return
		}

		// Only check simple identifiers as the target
		if varDecl.Name.Kind != ast.KindIdentifier {
			return
		}

		init := varDecl.Initializer

		// Skip optional chaining
		if isOptionalChaining(init) {
			return
		}

		// Check for array[0] pattern
		if opts.ArrayDestructuring && init.Kind == ast.KindElementAccessExpression {
			elemAccess := init.AsElementAccessExpression()
			if elemAccess != nil && elemAccess.ArgumentExpression != nil {
				// Check if it's a numeric literal 0, 1, 2, etc.
				if elemAccess.ArgumentExpression.Kind == ast.KindNumericLiteral {
					// Report without auto-fix (complex transformation)
					ctx.ReportNode(node, buildArrayMessage())
					return
				}
			}
		}

		// Check for object.property pattern
		if opts.ObjectDestructuring && init.Kind == ast.KindPropertyAccessExpression {
			propAccess := init.AsPropertyAccessExpression()
			if propAccess != nil && propAccess.Name() != nil {
				varName := varDecl.Name.AsIdentifier().Text()
				propName := propAccess.Name().Text()

				// If property name matches variable name, suggest destructuring
				if varName == propName || opts.EnforceForRenamedProperties {
					objText := getNodeText(ctx.SourceFile, propAccess.Expression)

					// Only auto-fix if names match (simple case)
					if varName == propName && !utils.HasCommentsInRange(ctx.SourceFile, utils.TrimNodeTextRange(ctx.SourceFile, node)) {
						replacement := "{" + varName + "} = " + objText
						ctx.ReportNodeWithFixes(varDecl.Name, buildObjectMessage(),
							rule.RuleFixReplace(ctx.SourceFile, varDecl.Name.Parent, replacement))
					} else {
						ctx.ReportNode(varDecl.Name, buildObjectMessage())
					}
					return
				}
			}
		}

		// Check for object['property'] pattern
		if opts.ObjectDestructuring && init.Kind == ast.KindElementAccessExpression {
			elemAccess := init.AsElementAccessExpression()
			if elemAccess != nil && elemAccess.ArgumentExpression != nil {
				// Only for string literal property access
				if elemAccess.ArgumentExpression.Kind == ast.KindStringLiteral {
					// Report without auto-fix for now
					ctx.ReportNode(node, buildObjectMessage())
					return
				}
			}
		}
	}

	return rule.RuleListeners{
		ast.KindVariableDeclaration: func(node *ast.Node) {
			checkVariableDeclarator(node)
		},
		ast.KindAssignmentExpression: func(node *ast.Node) {
			assignExpr := node.AsAssignmentExpression()
			if assignExpr == nil || assignExpr.Right == nil || assignExpr.Left == nil {
				return
			}

			// Only check simple identifier assignments
			if assignExpr.Left.Kind != ast.KindIdentifier {
				return
			}

			init := assignExpr.Right

			// Skip optional chaining
			if isOptionalChaining(init) {
				return
			}

			// Check for array[0] pattern
			if opts.ArrayDestructuring && init.Kind == ast.KindElementAccessExpression {
				elemAccess := init.AsElementAccessExpression()
				if elemAccess != nil && elemAccess.ArgumentExpression != nil {
					if elemAccess.ArgumentExpression.Kind == ast.KindNumericLiteral {
						ctx.ReportNode(assignExpr.Left, buildArrayMessage())
						return
					}
				}
			}

			// Check for object.property pattern
			if opts.ObjectDestructuring && init.Kind == ast.KindPropertyAccessExpression {
				propAccess := init.AsPropertyAccessExpression()
				if propAccess != nil && propAccess.Name() != nil {
					leftIdent := assignExpr.Left.AsIdentifier()
					if leftIdent != nil {
						varName := leftIdent.Text()
						propName := propAccess.Name().Text()

						if varName == propName || opts.EnforceForRenamedProperties {
							ctx.ReportNode(assignExpr.Left, buildObjectMessage())
							return
						}
					}
				}
			}

			// Check for object['property'] pattern
			if opts.ObjectDestructuring && init.Kind == ast.KindElementAccessExpression {
				elemAccess := init.AsElementAccessExpression()
				if elemAccess != nil && elemAccess.ArgumentExpression != nil {
					if elemAccess.ArgumentExpression.Kind == ast.KindStringLiteral {
						ctx.ReportNode(assignExpr.Left, buildObjectMessage())
						return
					}
				}
			}
		},
	}
}
