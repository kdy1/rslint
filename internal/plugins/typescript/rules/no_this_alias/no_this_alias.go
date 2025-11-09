package no_this_alias

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type NoThisAliasOptions struct {
	AllowDestructuring bool     `json:"allowDestructuring"`
	AllowedNames       []string `json:"allowedNames"`
}

var NoThisAliasRule = rule.CreateRule(rule.Rule{
	Name: "no-this-alias",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := NoThisAliasOptions{
			AllowDestructuring: true,
			AllowedNames:       []string{},
		}
		// Parse options with dual-format support (handles both array and object formats)
		if options != nil {
			var optsMap map[string]interface{}
			var ok bool

			// Handle array format: [{ option: value }]
			if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
				optsMap, ok = optArray[0].(map[string]interface{})
			} else {
				// Handle direct object format: { option: value }
				optsMap, ok = options.(map[string]interface{})
			}

			if ok {
				if allowDestructuring, ok := optsMap["allowDestructuring"].(bool); ok {
					opts.AllowDestructuring = allowDestructuring
				}
				if allowedNames, ok := optsMap["allowedNames"].([]interface{}); ok {
					opts.AllowedNames = make([]string, 0, len(allowedNames))
					for _, name := range allowedNames {
						if str, ok := name.(string); ok {
							opts.AllowedNames = append(opts.AllowedNames, str)
						}
					}
				}
			}
		}

		isNameAllowed := func(name string) bool {
			for _, allowedName := range opts.AllowedNames {
				if name == allowedName {
					return true
				}
			}
			return false
		}

		checkVariableDeclaration := func(node *ast.Node) {
			varStmt := node.AsVariableStatement()
			if varStmt == nil {
				return
			}

			if varStmt.DeclarationList == nil {
				return
			}

			declList := varStmt.DeclarationList.AsVariableDeclarationList()
			if declList == nil {
				return
			}

			for _, decl := range declList.Declarations.Nodes {
				varDecl := decl.AsVariableDeclaration()
				if varDecl == nil {
					continue
				}

				// Check if initializer is 'this'
				if varDecl.Initializer == nil {
					continue
				}

				init := varDecl.Initializer
				if init.Kind != ast.KindThisKeyword {
					continue
				}

				// Check the pattern to determine the type of assignment
				if varDecl.Name() == nil {
					continue
				}

				// Handle different binding patterns
				switch varDecl.Name().Kind {
				case ast.KindIdentifier:
					// Simple assignment: const foo = this
					ident := varDecl.Name().AsIdentifier()
					if ident == nil {
						continue
					}
					identName := ident.Text

					if !isNameAllowed(identName) {
						ctx.ReportNode(varDecl.Name(), rule.RuleMessage{
							Id:          "thisAssignment",
							Description: "Unexpected aliasing of 'this' to local variable.",
						})
					}

				case ast.KindObjectBindingPattern:
					// Destructuring: const { props, state } = this
					if !opts.AllowDestructuring {
						ctx.ReportNode(varDecl.Name(), rule.RuleMessage{
							Id:          "thisDestructure",
							Description: "Unexpected aliasing of members of 'this' to local variables.",
						})
					}

				case ast.KindArrayBindingPattern:
					// Array destructuring: const [foo, bar] = this
					if !opts.AllowDestructuring {
						ctx.ReportNode(varDecl.Name(), rule.RuleMessage{
							Id:          "thisDestructure",
							Description: "Unexpected aliasing of members of 'this' to local variables.",
						})
					}
				}
			}
		}

		checkAssignmentExpression := func(node *ast.Node) {
			binaryExpr := node.AsBinaryExpression()
			if binaryExpr == nil {
				return
			}

			// Check if this is an assignment (=)
			if binaryExpr.OperatorToken.Kind != ast.KindEqualsToken {
				return
			}

			// Check if right side is 'this'
			if binaryExpr.Right == nil || binaryExpr.Right.Kind != ast.KindThisKeyword {
				return
			}

			// Check if left side is an identifier
			if binaryExpr.Left == nil {
				return
			}

			// Handle different assignment patterns
			switch binaryExpr.Left.Kind {
			case ast.KindIdentifier:
				// Simple assignment: foo = this
				ident := binaryExpr.Left.AsIdentifier()
				if ident == nil {
					return
				}
				identName := ident.Text

				if !isNameAllowed(identName) {
					ctx.ReportNode(binaryExpr.Left, rule.RuleMessage{
						Id:          "thisAssignment",
						Description: "Unexpected aliasing of 'this' to local variable.",
					})
				}

			case ast.KindObjectBindingPattern:
				// Destructuring: ({ props, state } = this)
				if !opts.AllowDestructuring {
					ctx.ReportNode(binaryExpr.Left, rule.RuleMessage{
						Id:          "thisDestructure",
						Description: "Unexpected aliasing of members of 'this' to local variables.",
					})
				}

			case ast.KindArrayBindingPattern:
				// Array destructuring: ([foo, bar] = this)
				if !opts.AllowDestructuring {
					ctx.ReportNode(binaryExpr.Left, rule.RuleMessage{
						Id:          "thisDestructure",
						Description: "Unexpected aliasing of members of 'this' to local variables.",
					})
				}
			}
		}

		return rule.RuleListeners{
			ast.KindVariableStatement:  checkVariableDeclaration,
			ast.KindBinaryExpression:   checkAssignmentExpression,
		}
	},
})
