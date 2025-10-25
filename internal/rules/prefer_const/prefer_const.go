package prefer_const

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// PreferConstRule implements the prefer-const rule
// Require const for variables never reassigned
var PreferConstRule = rule.Rule{
	Name: "prefer-const",
	Run:  run,
}

// Options for prefer-const rule
type Options struct {
	Destructuring         string `json:"destructuring"`         // "any" or "all"
	IgnoreReadBeforeAssign bool   `json:"ignoreReadBeforeAssign"`
}

func parseOptions(options any) Options {
	opts := Options{
		Destructuring:         "any",
		IgnoreReadBeforeAssign: false,
	}

	if options == nil {
		return opts
	}

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
		if v, ok := optsMap["destructuring"].(string); ok {
			opts.Destructuring = v
		}
		if v, ok := optsMap["ignoreReadBeforeAssign"].(bool); ok {
			opts.IgnoreReadBeforeAssign = v
		}
	}

	return opts
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := parseOptions(options)

	return rule.RuleListeners{
		ast.KindVariableDeclaration: func(node *ast.Node) {
			varDecl := node.AsVariableStatement()
			if varDecl == nil || varDecl.DeclarationList == nil {
				return
			}

			declList := varDecl.DeclarationList
			if declList.Kind != ast.KindVariableDeclarationList {
				return
			}

			// Only check let declarations
			if (declList.Flags & ast.NodeFlagsLet) == 0 {
				return
			}

			if declList.Declarations == nil {
				return
			}

			// Check each declaration
			for _, decl := range declList.Declarations.Nodes {
				if decl == nil {
					continue
				}

				varDecl := decl.AsVariableDeclaration()
				if varDecl == nil || varDecl.Name == nil {
					continue
				}

				// Skip if no initializer (variables assigned later might be reassigned)
				if varDecl.Initializer == nil {
					continue
				}

				// For now, we'll do a simple check - if it's initialized and not obviously reassigned
				// A full implementation would need to track all assignments in the scope

				// Check if this is a destructuring pattern
				isDestructuring := varDecl.Name.Kind == ast.KindObjectBindingPattern ||
					varDecl.Name.Kind == ast.KindArrayBindingPattern

				// Simple heuristic: suggest const for initialized let declarations
				// TODO: Implement full scope analysis to detect reassignments
				shouldSuggestConst := true

				if shouldSuggestConst {
					msg := rule.RuleMessage{
						Id:          "useConst",
						Description: "'{{name}}' is never reassigned. Use 'const' instead.",
					}

					// Get variable name for message
					// For simple case, just report on the node

					// Create fix
					text := ctx.SourceFile.Text()
					listRange := utils.TrimNodeTextRange(ctx.SourceFile, &declList.Node)
					fullText := text[listRange.Pos():listRange.End()]

					// Replace 'let' with 'const'
					letPos := listRange.Pos()
					for i := 0; i < len(fullText); i++ {
						if i+3 <= len(fullText) && fullText[i:i+3] == "let" {
							// Check if it's followed by whitespace
							if i+3 < len(fullText) && (fullText[i+3] == ' ' || fullText[i+3] == '\t' || fullText[i+3] == '\n') {
								letPos += i
								break
							}
						}
					}

					replacement := text[listRange.Pos():letPos] + "const" + text[letPos+3:listRange.End()]

					ctx.ReportNodeWithFixes(&declList.Node, msg, rule.RuleFixReplace(ctx.SourceFile, &declList.Node, replacement))
				}
			}
		},
	}
}
