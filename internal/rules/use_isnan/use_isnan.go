package use_isnan

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// UseIsnanOptions defines the configuration options for this rule
type UseIsnanOptions struct {
	EnforceForSwitchCase bool `json:"enforceForSwitchCase"`
	EnforceForIndexOf    bool `json:"enforceForIndexOf"`
}

// parseOptions parses and validates the rule options
func parseOptions(options any) UseIsnanOptions {
	opts := UseIsnanOptions{
		EnforceForSwitchCase: true,
		EnforceForIndexOf:    false,
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
		if v, ok := optsMap["enforceForSwitchCase"].(bool); ok {
			opts.EnforceForSwitchCase = v
		}
		if v, ok := optsMap["enforceForIndexOf"].(bool); ok {
			opts.EnforceForIndexOf = v
		}
	}

	return opts
}

// isNaNIdentifier checks if a node is NaN or Number.NaN
func isNaNIdentifier(node *ast.Node) bool {
	if node == nil {
		return false
	}

	// Handle comma/sequence expressions - check last element
	if node.Kind == ast.KindCommaListExpression {
		comma := node.AsCommaListExpression()
		if comma != nil && comma.Elements != nil && len(comma.Elements.Nodes) > 0 {
			return isNaNIdentifier(comma.Elements.Nodes[len(comma.Elements.Nodes)-1])
		}
		return false
	}

	// Check for simple NaN identifier
	if node.Kind == ast.KindIdentifier {
		ident := node.AsIdentifier()
		return ident != nil && ident.Text() == "NaN"
	}

	// Check for Number.NaN
	if node.Kind == ast.KindPropertyAccessExpression {
		propAccess := node.AsPropertyAccessExpression()
		if propAccess == nil || propAccess.Expression == nil || propAccess.Name() == nil {
			return false
		}

		// Check if expression is "Number" identifier
		if propAccess.Expression.Kind == ast.KindIdentifier {
			obj := propAccess.Expression.AsIdentifier()
			if obj != nil && obj.Text() == "Number" && propAccess.Name().Text() == "NaN" {
				return true
			}
		}
	}

	return false
}

// fixableOperators are operators that can have fix suggestions
var fixableOperators = map[ast.SyntaxKind]bool{
	ast.SyntaxKindEqualsEqualsToken:         true,
	ast.SyntaxKindEqualsEqualsEqualsToken:   true,
	ast.SyntaxKindExclamationEqualsToken:    true,
	ast.SyntaxKindExclamationEqualsEqualsToken: true,
}

// castableOperators support the casting suggestion
var castableOperators = map[ast.SyntaxKind]bool{
	ast.SyntaxKindEqualsEqualsToken:      true,
	ast.SyntaxKindExclamationEqualsToken: true,
}

// UseIsnanRule implements the use-isnan rule
// Require calls to `isNaN()` when checking for `NaN`
var UseIsnanRule = rule.Rule{
	Name: "use-isnan",
	Run:  run,
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := parseOptions(options)

	// Check binary expressions for NaN comparisons
	checkBinaryExpression := func(node *ast.Node) {
		binary := node.AsBinaryExpression()
		if binary == nil {
			return
		}

		op := binary.OperatorToken.Kind

		// Check for comparison operators: ==, ===, !=, !==, <, >, <=, >=
		isComparison := op == ast.SyntaxKindEqualsEqualsToken ||
			op == ast.SyntaxKindEqualsEqualsEqualsToken ||
			op == ast.SyntaxKindExclamationEqualsToken ||
			op == ast.SyntaxKindExclamationEqualsEqualsToken ||
			op == ast.SyntaxKindLessThanToken ||
			op == ast.SyntaxKindGreaterThanToken ||
			op == ast.SyntaxKindLessThanEqualsToken ||
			op == ast.SyntaxKindGreaterThanEqualsToken

		if !isComparison {
			return
		}

		// Check if either operand is NaN
		if !isNaNIdentifier(binary.Left) && !isNaNIdentifier(binary.Right) {
			return
		}

		// Determine which operand is the value being compared
		nanNode := binary.Left
		if isNaNIdentifier(binary.Left) {
			nanNode = binary.Left
		} else {
			nanNode = binary.Right
		}

		isSequenceExpression := nanNode.Kind == ast.KindCommaListExpression
		isSuggestable := fixableOperators[op] && !isSequenceExpression
		isCastable := castableOperators[op]

		suggestions := []rule.RuleSuggestion{}

		if isSuggestable {
			// Get the compared value (non-NaN operand)
			comparedValue := binary.Right
			if isNaNIdentifier(binary.Right) {
				comparedValue = binary.Left
			}

			text := ctx.SourceFile.Text()
			comparedRange := utils.TrimNodeTextRange(ctx.SourceFile, comparedValue)
			comparedText := text[comparedRange.Pos():comparedRange.End()]

			// Wrap in parentheses if needed (for sequence expressions, though we filter those out)
			shouldWrap := comparedValue.Kind == ast.KindCommaListExpression
			if shouldWrap {
				comparedText = fmt.Sprintf("(%s)", comparedText)
			}

			// Determine if we need negation (for != and !==)
			shouldNegate := op == ast.SyntaxKindExclamationEqualsToken || op == ast.SyntaxKindExclamationEqualsEqualsToken
			negation := ""
			if shouldNegate {
				negation = "!"
			}

			// Suggestion 1: Replace with Number.isNaN
			fix1 := rule.RuleFixReplace(ctx.SourceFile, node, fmt.Sprintf("%sNumber.isNaN(%s)", negation, comparedText))
			suggestions = append(suggestions, rule.RuleSuggestion{
				Description: "Replace with Number.isNaN.",
				Fix:         fix1,
			})

			// Suggestion 2: Replace with casting and Number.isNaN (only for == and !=)
			if isCastable {
				fix2 := rule.RuleFixReplace(ctx.SourceFile, node, fmt.Sprintf("%sNumber.isNaN(Number(%s))", negation, comparedText))
				suggestions = append(suggestions, rule.RuleSuggestion{
					Description: "Replace with Number.isNaN and cast to a Number.",
					Fix:         fix2,
				})
			}
		}

		ctx.ReportNodeWithSuggestions(node, rule.RuleMessage{
			Id:          "comparisonWithNaN",
			Description: "Use the isNaN function to compare with NaN.",
		}, suggestions)
	}

	// Check switch statements for NaN
	checkSwitchStatement := func(node *ast.Node) {
		switchStmt := node.AsSwitchStatement()
		if switchStmt == nil {
			return
		}

		// Check discriminant (the value being switched on)
		if isNaNIdentifier(switchStmt.Expression) {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "switchNaN",
				Description: "'switch(NaN)' can never match a case clause. Use Number.isNaN instead of the switch.",
			})
		}

		// Check case clauses
		if switchStmt.CaseBlock != nil && switchStmt.CaseBlock.Clauses != nil {
			for _, clause := range switchStmt.CaseBlock.Clauses.Nodes {
				if clause == nil || clause.Kind != ast.KindCaseClause {
					continue
				}
				caseClause := clause.AsCaseClause()
				if caseClause != nil && isNaNIdentifier(caseClause.Expression) {
					ctx.ReportNode(clause, rule.RuleMessage{
						Id:          "caseNaN",
						Description: "'case NaN' can never match. Use Number.isNaN before the switch.",
					})
				}
			}
		}
	}

	// Check for indexOf/lastIndexOf with NaN
	checkCallExpression := func(node *ast.Node) {
		call := node.AsCallExpression()
		if call == nil || call.Expression == nil {
			return
		}

		// Skip chain expressions to get actual callee
		callee := call.Expression
		for callee != nil && callee.Kind == ast.KindNonNullExpression {
			nonNull := callee.AsNonNullExpression()
			if nonNull == nil || nonNull.Expression == nil {
				break
			}
			callee = nonNull.Expression
		}

		// Check if callee is a member expression
		if callee.Kind != ast.KindPropertyAccessExpression && callee.Kind != ast.KindElementAccessExpression {
			return
		}

		var methodName string
		if callee.Kind == ast.KindPropertyAccessExpression {
			propAccess := callee.AsPropertyAccessExpression()
			if propAccess != nil && propAccess.Name() != nil {
				methodName = propAccess.Name().Text()
			}
		} else if callee.Kind == ast.KindElementAccessExpression {
			elemAccess := callee.AsElementAccessExpression()
			if elemAccess != nil && elemAccess.ArgumentExpression != nil {
				// Try to get static property name from bracket notation
				if elemAccess.ArgumentExpression.Kind == ast.KindStringLiteral {
					str := elemAccess.ArgumentExpression.AsStringLiteral()
					if str != nil {
						text := str.Text()
						methodName = strings.Trim(text, "\"'`")
					}
				}
			}
		}

		// Check if method is indexOf or lastIndexOf
		if methodName != "indexOf" && methodName != "lastIndexOf" {
			return
		}

		// Check arguments
		if call.Arguments == nil || len(call.Arguments.Nodes) == 0 || len(call.Arguments.Nodes) > 2 {
			return
		}

		// Check if first argument is NaN
		if !isNaNIdentifier(call.Arguments.Nodes[0]) {
			return
		}

		// Check if we can suggest a fix
		isSuggestable := call.Arguments.Nodes[0].Kind != ast.KindCommaListExpression && len(call.Arguments.Nodes) == 1

		suggestions := []rule.RuleSuggestion{}

		if isSuggestable {
			// Determine replacement method name
			findMethod := "findIndex"
			if methodName == "lastIndexOf" {
				findMethod = "findLastIndex"
			}

			text := ctx.SourceFile.Text()

			// Determine if we need to wrap the method name in quotes (for computed property access)
			shouldWrap := callee.Kind == ast.KindElementAccessExpression
			propertyName := findMethod
			if shouldWrap {
				propertyName = fmt.Sprintf("\"%s\"", findMethod)
			}

			// Build the fix
			var propRange utils.TextRange
			if callee.Kind == ast.KindPropertyAccessExpression {
				propAccess := callee.AsPropertyAccessExpression()
				if propAccess != nil && propAccess.Name() != nil {
					propRange = utils.TrimNodeTextRange(ctx.SourceFile, propAccess.Name())
				}
			} else if callee.Kind == ast.KindElementAccessExpression {
				elemAccess := callee.AsElementAccessExpression()
				if elemAccess != nil && elemAccess.ArgumentExpression != nil {
					propRange = utils.TrimNodeTextRange(ctx.SourceFile, elemAccess.ArgumentExpression)
				}
			}

			argRange := utils.TrimNodeTextRange(ctx.SourceFile, call.Arguments.Nodes[0])

			// Create composite fix: replace property name and replace argument
			fixParts := []rule.RuleFix{
				rule.RuleFixReplaceRange(ctx.SourceFile, propRange, propertyName),
				rule.RuleFixReplaceRange(ctx.SourceFile, argRange, "Number.isNaN"),
			}

			suggestions = append(suggestions, rule.RuleSuggestion{
				Description: fmt.Sprintf("Replace with Array.prototype.%s.", findMethod),
				Fix:         rule.RuleFixComposite(fixParts...),
			})
		}

		ctx.ReportNodeWithSuggestions(node, rule.RuleMessage{
			Id:          "indexOfNaN",
			Description: fmt.Sprintf("Array prototype method '%s' cannot find NaN.", methodName),
		}, suggestions)
	}

	listeners := rule.RuleListeners{
		ast.KindBinaryExpression: checkBinaryExpression,
	}

	if opts.EnforceForSwitchCase {
		listeners[ast.KindSwitchStatement] = checkSwitchStatement
	}

	if opts.EnforceForIndexOf {
		listeners[ast.KindCallExpression] = checkCallExpression
	}

	return listeners
}
