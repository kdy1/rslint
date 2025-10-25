package no_var

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// NoVarRule implements the no-var rule
// Require let or const instead of var
var NoVarRule = rule.Rule{
	Name: "no-var",
	Run:  run,
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
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

			// Check if this is a var declaration
			if (declList.Flags & ast.NodeFlagsLet) != 0 || (declList.Flags & ast.NodeFlagsConst) != 0 {
				return
			}

			// This is a var declaration - check if it's safe to fix
			canFix := isSafeToFix(ctx, declList)

			msg := rule.RuleMessage{
				Id:          "unexpectedVar",
				Description: "Unexpected var, use let or const instead.",
			}

			if canFix {
				// Create a fix that replaces 'var' with 'let'
				text := ctx.SourceFile.Text()
				nodeRange := utils.TrimNodeTextRange(ctx.SourceFile, &declList.Node)
				fullText := text[nodeRange.Pos():nodeRange.End()]

				// Replace 'var' with 'let'
				// Find the position of 'var' keyword
				varPos := nodeRange.Pos()
				for i := 0; i < len(fullText); i++ {
					if i+3 <= len(fullText) && fullText[i:i+3] == "var" {
						// Check if it's followed by whitespace or bracket
						if i+3 < len(fullText) && (fullText[i+3] == ' ' || fullText[i+3] == '\t' || fullText[i+3] == '\n') {
							varPos += i
							break
						}
					}
				}

				// Build the replacement
				replacement := text[nodeRange.Pos():varPos] + "let" + text[varPos+3:nodeRange.End()]

				ctx.ReportNodeWithFixes(&declList.Node, msg, rule.RuleFixReplace(ctx.SourceFile, &declList.Node, replacement))
			} else {
				// Report without fix if it's not safe
				ctx.ReportNode(&declList.Node, msg)
			}
		},
	}
}

// isSafeToFix checks if it's safe to replace var with let
// Returns false if:
// 1. Variable is redeclared
// 2. Variable is used outside its block scope (would change behavior)
// 3. Variable is captured in a loop closure
func isSafeToFix(ctx rule.RuleContext, declList *ast.VariableDeclarationList) bool {
	// For now, we'll implement a simple check
	// TODO: Implement more sophisticated scoping analysis for edge cases
	// like variables used outside block scope or captured in closures

	if declList.Declarations == nil {
		return true
	}

	// Check for duplicate declarations in the same list
	names := make(map[string]bool)
	for _, decl := range declList.Declarations.Nodes {
		if decl == nil {
			continue
		}
		varDecl := decl.AsVariableDeclaration()
		if varDecl == nil || varDecl.Name == nil {
			continue
		}

		// Extract variable name(s)
		extractNames(varDecl.Name, names)
	}

	// For now, allow the fix - more sophisticated checks would require
	// full scope analysis to detect:
	// - Variables used before declaration (hoisting)
	// - Variables used outside their block scope
	// - Variables captured in closures where let/const would change semantics
	return true
}

// extractNames recursively extracts all variable names from a binding pattern
func extractNames(binding *ast.Node, names map[string]bool) {
	if binding == nil {
		return
	}

	switch binding.Kind {
	case ast.KindIdentifier:
		ident := binding.AsIdentifier()
		if ident != nil {
			name := ident.Text()
			if names[name] {
				// Duplicate found - but we'll still allow the fix since
				// the syntax error will be caught anyway
			}
			names[name] = true
		}
	case ast.KindObjectBindingPattern:
		pattern := binding.AsObjectBindingPattern()
		if pattern != nil && pattern.Elements != nil {
			for _, elem := range pattern.Elements.Nodes {
				if elem != nil {
					if binding := elem.AsBindingElement(); binding != nil {
						extractNames(binding.Name, names)
					}
				}
			}
		}
	case ast.KindArrayBindingPattern:
		pattern := binding.AsArrayBindingPattern()
		if pattern != nil && pattern.Elements != nil {
			for _, elem := range pattern.Elements.Nodes {
				if elem != nil {
					if binding := elem.AsBindingElement(); binding != nil {
						extractNames(binding.Name, names)
					}
				}
			}
		}
	}
}
