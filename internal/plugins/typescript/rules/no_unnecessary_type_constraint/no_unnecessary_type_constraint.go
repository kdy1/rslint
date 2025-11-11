package no_unnecessary_type_constraint

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// NoUnnecessaryTypeConstraintRule disallows unnecessary constraints on generic types
var NoUnnecessaryTypeConstraintRule = rule.CreateRule(rule.Rule{
	Name: "no-unnecessary-type-constraint",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	// Helper to check if a type node is 'any' or 'unknown'
	isAnyOrUnknown := func(typeNode *ast.Node) bool {
		if typeNode == nil {
			return false
		}
		return typeNode.Kind == ast.KindAnyKeyword || typeNode.Kind == ast.KindUnknownKeyword
	}

	// Helper to get the constraint keyword text
	getConstraintKeywordText := func(typeNode *ast.Node) string {
		if typeNode == nil {
			return ""
		}
		if typeNode.Kind == ast.KindAnyKeyword {
			return "any"
		}
		if typeNode.Kind == ast.KindUnknownKeyword {
			return "unknown"
		}
		return ""
	}

	// Check a type parameter node for unnecessary constraint
	checkTypeParameter := func(node *ast.Node) {
		if node == nil || node.Kind != ast.KindTypeParameter {
			return
		}

		typeParam := node.AsTypeParameter()
		if typeParam == nil {
			return
		}

		// Check if constraint exists and is 'any' or 'unknown'
		constraint := typeParam.Constraint
		if constraint != nil && isAnyOrUnknown(constraint) {
			constraintKeyword := getConstraintKeywordText(constraint)

			// Calculate the start of the constraint (after the type parameter name)
			// We need to remove " extends any" or " extends unknown"
			constraintStart := constraint.Pos

			// Find the start position by looking backwards for "extends" keyword
			// The constraint starts at the "extends" keyword position
			sourceText := utils.GetNodeSourceCode(node, ctx.SourceFile)

			// Get the full source text to calculate positions
			fullText := utils.GetSourceFileText(ctx.SourceFile)
			nodeStart := node.Pos
			nodeEnd := node.End

			// Find "extends" keyword between parameter name and constraint
			var extendsStart int
			var extendsEnd int
			foundExtends := false

			// Search for "extends" keyword
			for i := nodeStart; i < constraintStart; i++ {
				if i+7 <= len(fullText) {
					substr := fullText[i:i+7]
					if substr == "extends" {
						extendsStart = i
						extendsEnd = constraintStart + (constraint.End - constraint.Pos)
						foundExtends = true
						break
					}
				}
			}

			// Calculate fix range - from space before "extends" to end of constraint
			// Handle cases like "T extends any", "T extends any,", "T extends any = default"
			fixStart := extendsStart
			fixEnd := extendsEnd

			// Look for space before "extends"
			if fixStart > 0 && fullText[fixStart-1] == ' ' {
				fixStart--
			}

			// Check what comes after the constraint
			afterConstraintEnd := fixEnd
			if afterConstraintEnd < nodeEnd {
				// Check for trailing comma or default value
				remainingText := fullText[afterConstraintEnd:nodeEnd]

				// For arrow functions in JSX, we might need to keep a trailing comma
				if len(remainingText) > 0 {
					// Skip whitespace
					trimIdx := 0
					for trimIdx < len(remainingText) && (remainingText[trimIdx] == ' ' || remainingText[trimIdx] == '\t') {
						trimIdx++
					}

					// If we find a comma, check if we need to keep it
					if trimIdx < len(remainingText) && remainingText[trimIdx] == ',' {
						// Include the trailing whitespace in the fix
						fixEnd = afterConstraintEnd + trimIdx
					}
				}
			}

			// Create the fix
			var fix *rule.RuleFix
			if foundExtends {
				fix = &rule.RuleFix{
					Range: [2]int{fixStart, fixEnd},
					Text:  "",
				}
			}

			// Report the violation
			if fix != nil {
				ctx.ReportNodeWithFixes(constraint, rule.RuleMessage{
					Id:          "unnecessaryConstraint",
					Description: "Constraining a type parameter to `" + constraintKeyword + "` or `unknown` is redundant as it is the default constraint.",
				}, []rule.RuleFix{*fix})
			} else {
				ctx.ReportNode(constraint, rule.RuleMessage{
					Id:          "unnecessaryConstraint",
					Description: "Constraining a type parameter to `" + constraintKeyword + "` or `unknown` is redundant as it is the default constraint.",
				})
			}
		}
	}

	// Check all type parameters in a node list
	checkTypeParameters := func(typeParams *ast.NodeArray) {
		if typeParams == nil {
			return
		}

		for _, param := range typeParams.Nodes {
			checkTypeParameter(param)
		}
	}

	return rule.RuleListeners{
		ast.KindFunctionDeclaration: func(node *ast.Node) {
			funcDecl := node.AsFunctionDeclaration()
			if funcDecl != nil {
				checkTypeParameters(funcDecl.TypeParameters)
			}
		},
		ast.KindArrowFunction: func(node *ast.Node) {
			arrowFunc := node.AsArrowFunction()
			if arrowFunc != nil {
				checkTypeParameters(arrowFunc.TypeParameters)
			}
		},
		ast.KindFunctionExpression: func(node *ast.Node) {
			funcExpr := node.AsFunctionExpression()
			if funcExpr != nil {
				checkTypeParameters(funcExpr.TypeParameters)
			}
		},
		ast.KindMethodDeclaration: func(node *ast.Node) {
			methodDecl := node.AsMethodDeclaration()
			if methodDecl != nil {
				checkTypeParameters(methodDecl.TypeParameters)
			}
		},
		ast.KindClassDeclaration: func(node *ast.Node) {
			classDecl := node.AsClassDeclaration()
			if classDecl != nil {
				checkTypeParameters(classDecl.TypeParameters)
			}
		},
		ast.KindClassExpression: func(node *ast.Node) {
			classExpr := node.AsClassExpression()
			if classExpr != nil {
				checkTypeParameters(classExpr.TypeParameters)
			}
		},
		ast.KindInterfaceDeclaration: func(node *ast.Node) {
			interfaceDecl := node.AsInterfaceDeclaration()
			if interfaceDecl != nil {
				checkTypeParameters(interfaceDecl.TypeParameters)
			}
		},
		ast.KindTypeAliasDeclaration: func(node *ast.Node) {
			typeAlias := node.AsTypeAliasDeclaration()
			if typeAlias != nil {
				checkTypeParameters(typeAlias.TypeParameters)
			}
		},
	}
}
