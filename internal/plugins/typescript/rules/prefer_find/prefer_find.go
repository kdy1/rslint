package prefer_find

import (
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildPreferFindMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferFind",
		Description: "Prefer .find() over .filter()[0] to search for a single matching element.",
	}
}

func buildPreferFindSuggestionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferFindSuggestion",
		Description: "Use .find() instead.",
	}
}

// isZero checks if the expression evaluates to zero (0, 0n, -0, -0n, NaN, or floats that round to 0)
func isZero(typeChecker *checker.Checker, node *ast.Node) bool {
	if node == nil {
		return false
	}

	// Check for numeric literal 0 or -0
	if ast.IsNumericLiteral(node) {
		val := node.AsNumericLiteral().Text
		if val == "0" || val == "-0" {
			return true
		}
		// Check if it's a float that rounds to 0 (like -0.12635678)
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			if int(f) == 0 {
				return true
			}
		}
	}

	// Check for bigint literal 0n or -0n
	if node.Kind == ast.KindBigIntLiteral {
		val := node.AsBigIntLiteral().Text
		// Remove 'n' suffix
		val = strings.TrimSuffix(val, "n")
		if val == "0" || val == "-0" {
			return true
		}
	}

	// Check for prefix unary expression with minus operator (like -0.12635678)
	if ast.IsPrefixUnaryExpression(node) {
		prefix := node.AsPrefixUnaryExpression()
		if prefix.Operator == ast.KindMinusToken {
			return isZero(typeChecker, prefix.Operand)
		}
	}

	// Check for NaN identifier
	if ast.IsIdentifier(node) && node.AsIdentifier().Text == "NaN" {
		return true
	}

	// Check for string literal "0"
	if ast.IsStringLiteral(node) {
		val := node.AsStringLiteral().Text
		if val == "0" {
			return true
		}
	}

	// Check for variable that has a constant value of 0
	t := typeChecker.GetTypeAtLocation(node)
	if t != nil && utils.IsTypeFlagSet(t, checker.TypeFlagsNumberLiteral) {
		// For number literal types, check if the value is 0
		if checker.Type_value(t) != nil {
			if v, ok := checker.Type_value(t).(float64); ok && int(v) == 0 {
				return true
			}
		}
	}
	if t != nil && utils.IsTypeFlagSet(t, checker.TypeFlagsBigIntLiteral) {
		// For bigint literal types, check if the value is 0n
		if checker.Type_value(t) != nil {
			if v, ok := checker.Type_value(t).(string); ok {
				v = strings.TrimSuffix(v, "n")
				if v == "0" || v == "-0" {
					return true
				}
			}
		}
	}

	return false
}

// isFilterCall checks if the node is a filter() method call on an array
func isFilterCall(ctx rule.RuleContext, node *ast.Node) bool {
	if !ast.IsCallExpression(node) {
		return false
	}

	callExpr := node.AsCallExpression()
	if !ast.IsAccessExpression(callExpr.Expression) {
		return false
	}

	// Check if it's a "filter" method
	propertyName, found := checker.Checker_getAccessedPropertyName(ctx.TypeChecker, callExpr.Expression)
	if !found || propertyName != "filter" {
		return false
	}

	// Check if the expression is on an array type
	t := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, callExpr.Expression.Expression())
	return utils.TypeRecurser(t, func(t *checker.Type) bool {
		return checker.Checker_isArrayOrTupleType(ctx.TypeChecker, t)
	})
}

// checkElementAccessExpression checks arr.filter()[0]
func checkElementAccessExpression(ctx rule.RuleContext, node *ast.Node) {
	elemAccess := node.AsElementAccessExpression()

	// Skip optional chaining on the element access itself (arr.filter()?.[0])
	if elemAccess.QuestionDotToken != nil {
		return
	}

	// Check if the expression is a filter call
	filterExpr := ast.SkipParentheses(elemAccess.Expression)
	if !isFilterCall(ctx, filterExpr) {
		return
	}

	// Check if we're accessing index 0
	if !isZero(ctx.TypeChecker, elemAccess.ArgumentExpression) {
		return
	}

	// Get the call expression to check for optional chaining
	filterCallExpr := filterExpr.AsCallExpression()

	// Skip if filter itself is optionally called (arr.filter?.(x => true))
	if filterCallExpr.QuestionDotToken != nil {
		return
	}

	// Check if the filter's target might not be an array
	// This handles cases like: notNecessarilyAnArray?.filter(item => true)[0]
	accessExpr := filterCallExpr.Expression
	if ast.IsPropertyAccessExpression(accessExpr) || ast.IsElementAccessExpression(accessExpr) {
		// Check for optional chaining in the property access chain
		if ast.IsPropertyAccessExpression(accessExpr) && accessExpr.AsPropertyAccessExpression().QuestionDotToken != nil {
			// The array itself is optional, so we need to check the type more carefully
			baseExpr := accessExpr.AsPropertyAccessExpression().Expression
			baseType := ctx.TypeChecker.GetTypeAtLocation(baseExpr)

			// If the base type includes undefined/null, skip
			if utils.TypeRecurser(baseType, func(t *checker.Type) bool {
				return utils.IsTypeNullableType(t)
			}) {
				return
			}
		} else if ast.IsElementAccessExpression(accessExpr) && accessExpr.AsElementAccessExpression().QuestionDotToken != nil {
			baseExpr := accessExpr.AsElementAccessExpression().Expression
			baseType := ctx.TypeChecker.GetTypeAtLocation(baseExpr)

			if utils.TypeRecurser(baseType, func(t *checker.Type) bool {
				return utils.IsTypeNullableType(t)
			}) {
				return
			}
		}
	}

	reportViolation(ctx, node, filterExpr)
}

// checkCallExpression checks arr.filter().at(0)
func checkCallExpression(ctx rule.RuleContext, node *ast.Node) {
	callExpr := node.AsCallExpression()

	// Skip optional chaining on the at() call
	if callExpr.QuestionDotToken != nil {
		return
	}

	// Check if it's calling "at"
	if !ast.IsAccessExpression(callExpr.Expression) {
		return
	}

	propertyName, found := checker.Checker_getAccessedPropertyName(ctx.TypeChecker, callExpr.Expression)
	if !found || propertyName != "at" {
		return
	}

	// Check if we have an argument and it's 0
	if len(callExpr.Arguments()) != 1 {
		return
	}

	if !isZero(ctx.TypeChecker, callExpr.Arguments()[0]) {
		return
	}

	// Get the expression being called with .at()
	atTarget := callExpr.Expression
	var atTargetExpr *ast.Node
	if ast.IsPropertyAccessExpression(atTarget) {
		atTargetExpr = atTarget.AsPropertyAccessExpression().Expression
	} else if ast.IsElementAccessExpression(atTarget) {
		atTargetExpr = atTarget.AsElementAccessExpression().Expression
	}

	if atTargetExpr == nil {
		return
	}

	// Check if the target is a filter call
	filterExpr := ast.SkipParentheses(atTargetExpr)
	if !isFilterCall(ctx, filterExpr) {
		return
	}

	// Get the call expression to check for optional chaining
	filterCallExpr := filterExpr.AsCallExpression()

	// Skip if filter itself is optionally called
	if filterCallExpr.QuestionDotToken != nil {
		return
	}

	// Check for optional chaining in the filter's access expression
	accessExpr := filterCallExpr.Expression
	if ast.IsPropertyAccessExpression(accessExpr) || ast.IsElementAccessExpression(accessExpr) {
		if ast.IsPropertyAccessExpression(accessExpr) && accessExpr.AsPropertyAccessExpression().QuestionDotToken != nil {
			baseExpr := accessExpr.AsPropertyAccessExpression().Expression
			baseType := ctx.TypeChecker.GetTypeAtLocation(baseExpr)

			if utils.TypeRecurser(baseType, func(t *checker.Type) bool {
				return utils.IsTypeNullableType(t)
			}) {
				return
			}
		} else if ast.IsElementAccessExpression(accessExpr) && accessExpr.AsElementAccessExpression().QuestionDotToken != nil {
			baseExpr := accessExpr.AsElementAccessExpression().Expression
			baseType := ctx.TypeChecker.GetTypeAtLocation(baseExpr)

			if utils.TypeRecurser(baseType, func(t *checker.Type) bool {
				return utils.IsTypeNullableType(t)
			}) {
				return
			}
		}
	}

	reportViolation(ctx, node, filterExpr)
}

// reportViolation creates a suggestion to replace .filter() with .find()
func reportViolation(ctx rule.RuleContext, accessNode *ast.Node, filterNode *ast.Node) {
	filterCallExpr := filterNode.AsCallExpression()
	filterAccessExpr := filterCallExpr.Expression

	// Build the fix
	fixes := []rule.RuleFix{}

	// Step 1: Replace "filter" with "find"
	if ast.IsPropertyAccessExpression(filterAccessExpr) {
		propAccess := filterAccessExpr.AsPropertyAccessExpression()
		nameNode := propAccess.Name
		nameRange := utils.TrimNodeTextRange(ctx.SourceFile, nameNode)
		fixes = append(fixes, rule.RuleFixReplaceRange(nameRange, "find"))
	} else if ast.IsElementAccessExpression(filterAccessExpr) {
		elemAccess := filterAccessExpr.AsElementAccessExpression()
		argExpr := elemAccess.ArgumentExpression

		// Check if the argument is a string literal
		var newText string
		if ast.IsStringLiteral(argExpr) {
			newText = `"find"`
		} else {
			// For computed property names, try to keep the same format
			newText = `"find"`
		}

		argRange := utils.TrimNodeTextRange(ctx.SourceFile, argExpr)
		fixes = append(fixes, rule.RuleFixReplaceRange(argRange, newText))
	}

	// Step 2: Remove the [0] or .at(0) part
	// Find the range from the end of the filter call to the end of the access
	filterEnd := utils.TrimNodeTextRange(ctx.SourceFile, filterNode).End()
	accessEnd := utils.TrimNodeTextRange(ctx.SourceFile, accessNode).End()

	if filterEnd < accessEnd {
		fixes = append(fixes, rule.RuleFixRemoveRange(core.NewTextRange(filterEnd, accessEnd)))
	}

	// Create the suggestion
	suggestion := rule.RuleSuggestion{
		Message:  buildPreferFindSuggestionMessage(),
		FixesArr: fixes,
	}

	ctx.ReportNodeWithSuggestions(accessNode, buildPreferFindMessage(), suggestion)
}

// Helper function to recursively check conditional expressions
func checkConditionalExpression(ctx rule.RuleContext, node *ast.Node, visitedNodes *utils.Set[*ast.Node]) {
	if visitedNodes.Has(node) {
		return
	}
	visitedNodes.Add(node)

	if !ast.IsConditionalExpression(node) {
		return
	}

	condExpr := node.AsConditionalExpression()

	// Check both branches
	checkConditionalBranch(ctx, condExpr.WhenTrue, visitedNodes)
	checkConditionalBranch(ctx, condExpr.WhenFalse, visitedNodes)
}

func checkConditionalBranch(ctx rule.RuleContext, branch *ast.Node, visitedNodes *utils.Set[*ast.Node]) {
	branch = ast.SkipParentheses(branch)

	// If it's another conditional, recurse
	if ast.IsConditionalExpression(branch) {
		checkConditionalExpression(ctx, branch, visitedNodes)
	} else if ast.IsBinaryExpression(branch) {
		// Handle sequence expressions (comma operator)
		binary := branch.AsBinaryExpression()
		if binary.OperatorToken.Kind == ast.KindCommaToken {
			// The rightmost expression is what matters
			checkConditionalBranch(ctx, binary.Right, visitedNodes)
		}
	}
}

var PreferFindRule = rule.CreateRule(rule.Rule{
	Name: "prefer-find",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		visitedConditionals := utils.NewSetFromItems[*ast.Node]()

		return rule.RuleListeners{
			// Check for arr.filter()[0]
			ast.KindElementAccessExpression: func(node *ast.Node) {
				// Check if this is part of a conditional expression that accesses [0]
				parent := node.Parent
				if ast.IsConditionalExpression(parent) {
					// This will be handled specially
					return
				}

				// Also check parent chains for conditionals
				current := parent
				for current != nil {
					if ast.IsConditionalExpression(current) {
						return
					}
					// Stop at expression boundaries
					if ast.IsExpressionStatement(current) || ast.IsVariableDeclaration(current) {
						break
					}
					current = current.Parent
				}

				checkElementAccessExpression(ctx, node)
			},

			// Check for arr.filter().at(0)
			ast.KindCallExpression: func(node *ast.Node) {
				// Check if this is part of a conditional expression that calls .at(0)
				parent := node.Parent
				if ast.IsConditionalExpression(parent) {
					return
				}

				// Also check parent chains for conditionals
				current := parent
				for current != nil {
					if ast.IsConditionalExpression(current) {
						return
					}
					if ast.IsExpressionStatement(current) || ast.IsVariableDeclaration(current) {
						break
					}
					current = current.Parent
				}

				checkCallExpression(ctx, node)
			},

			// Special handling for conditional expressions
			ast.KindConditionalExpression: func(node *ast.Node) {
				if visitedConditionals.Has(node) {
					return
				}

				// Find the outermost conditional in the chain
				outermost := node
				for outermost.Parent != nil && ast.IsConditionalExpression(outermost.Parent) {
					outermost = outermost.Parent
				}

				// Now check if the whole conditional is being accessed with [0] or .at(0)
				parent := outermost.Parent

				var accessNode *ast.Node
				var isElementAccess bool
				var isAtCall bool

				// Check for (cond ? a : b)[0]
				if ast.IsElementAccessExpression(parent) {
					elemAccess := parent.AsElementAccessExpression()
					if elemAccess.Expression == outermost && isZero(ctx.TypeChecker, elemAccess.ArgumentExpression) {
						accessNode = parent
						isElementAccess = true
					}
				} else if ast.IsCallExpression(parent) {
					// Check for (cond ? a : b).at(0)
					callExpr := parent.AsCallExpression()
					if ast.IsAccessExpression(callExpr.Expression) {
						var atTarget *ast.Node
						if ast.IsPropertyAccessExpression(callExpr.Expression) {
							atTarget = callExpr.Expression.AsPropertyAccessExpression().Expression
						} else if ast.IsElementAccessExpression(callExpr.Expression) {
							atTarget = callExpr.Expression.AsElementAccessExpression().Expression
						}

						if atTarget == outermost {
							propertyName, found := checker.Checker_getAccessedPropertyName(ctx.TypeChecker, callExpr.Expression)
							if found && propertyName == "at" && len(callExpr.Arguments()) == 1 && isZero(ctx.TypeChecker, callExpr.Arguments()[0]) {
								accessNode = parent
								isAtCall = true
							}
						}
					}
				}

				if accessNode == nil {
					return
				}

				// Now check if all branches use .filter()
				allBranchesAreFilter := true
				var filterNodes []*ast.Node

				var collectFilterNodes func(*ast.Node)
				collectFilterNodes = func(branch *ast.Node) {
					branch = ast.SkipParentheses(branch)

					if ast.IsConditionalExpression(branch) {
						condExpr := branch.AsConditionalExpression()
						collectFilterNodes(condExpr.WhenTrue)
						collectFilterNodes(condExpr.WhenFalse)
					} else if ast.IsBinaryExpression(branch) {
						// Handle sequence expressions
						binary := branch.AsBinaryExpression()
						if binary.OperatorToken.Kind == ast.KindCommaToken {
							collectFilterNodes(binary.Right)
						} else {
							allBranchesAreFilter = false
						}
					} else {
						// Check if this branch is a filter call
						if isFilterCall(ctx, branch) {
							filterNodes = append(filterNodes, branch)
						} else {
							allBranchesAreFilter = false
						}
					}
				}

				condExpr := outermost.AsConditionalExpression()
				collectFilterNodes(condExpr.WhenTrue)
				collectFilterNodes(condExpr.WhenFalse)

				if !allBranchesAreFilter || len(filterNodes) == 0 {
					return
				}

				// Build the fix: replace all filter calls with find calls and remove the accessor
				fixes := []rule.RuleFix{}

				for _, filterNode := range filterNodes {
					filterCallExpr := filterNode.AsCallExpression()
					filterAccessExpr := filterCallExpr.Expression

					if ast.IsPropertyAccessExpression(filterAccessExpr) {
						propAccess := filterAccessExpr.AsPropertyAccessExpression()
						nameNode := propAccess.Name
						nameRange := utils.TrimNodeTextRange(ctx.SourceFile, nameNode)
						fixes = append(fixes, rule.RuleFixReplaceRange(nameRange, "find"))
					} else if ast.IsElementAccessExpression(filterAccessExpr) {
						elemAccess := filterAccessExpr.AsElementAccessExpression()
						argExpr := elemAccess.ArgumentExpression
						argRange := utils.TrimNodeTextRange(ctx.SourceFile, argExpr)
						fixes = append(fixes, rule.RuleFixReplaceRange(argRange, `"find"`))
					}
				}

				// Remove the [0] or .at(0) part
				outermostEnd := utils.TrimNodeTextRange(ctx.SourceFile, outermost).End()
				accessEnd := utils.TrimNodeTextRange(ctx.SourceFile, accessNode).End()

				if outermostEnd < accessEnd {
					fixes = append(fixes, rule.RuleFixRemoveRange(core.NewTextRange(outermostEnd, accessEnd)))
				}

				suggestion := rule.RuleSuggestion{
					Message:  buildPreferFindSuggestionMessage(),
					FixesArr: fixes,
				}

				ctx.ReportNodeWithSuggestions(accessNode, buildPreferFindMessage(), suggestion)

				visitedConditionals.Add(node)
			},
		}
	},
})
