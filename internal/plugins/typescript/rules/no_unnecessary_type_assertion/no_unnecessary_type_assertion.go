package no_unnecessary_type_assertion

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// NoUnnecessaryTypeAssertionOptions defines the configuration options for this rule
type NoUnnecessaryTypeAssertionOptions struct {
	TypesToIgnore []string `json:"typesToIgnore"`
}

// parseOptions parses and validates the rule options
func parseOptions(options any) NoUnnecessaryTypeAssertionOptions {
	opts := NoUnnecessaryTypeAssertionOptions{
		TypesToIgnore: []string{},
	}

	if options == nil {
		return opts
	}

	var optsMap map[string]interface{}
	if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
		optsMap, _ = optArray[0].(map[string]interface{})
	} else {
		optsMap, _ = options.(map[string]interface{})
	}

	if optsMap != nil {
		if v, ok := optsMap["typesToIgnore"].([]interface{}); ok {
			for _, item := range v {
				if str, ok := item.(string); ok {
					opts.TypesToIgnore = append(opts.TypesToIgnore, str)
				}
			}
		}
	}

	return opts
}

func buildUnnecessaryAssertionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unnecessaryAssertion",
		Description: "This assertion is unnecessary since it does not change the type of the expression.",
	}
}

func buildContextuallyUnnecessaryMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "contextuallyUnnecessary",
		Description: "This assertion is unnecessary in this context.",
	}
}

// isTypeInIgnoreList checks if a type name is in the ignore list
func isTypeInIgnoreList(typeName string, ignoreList []string) bool {
	for _, ignored := range ignoreList {
		if typeName == ignored {
			return true
		}
	}
	return false
}

// getTypeName gets a simple type name from a type node
func getTypeName(typeNode *ast.Node) string {
	if typeNode == nil {
		return ""
	}

	switch typeNode.Kind {
	case ast.KindTypeReference:
		typeRef := typeNode.AsTypeReference()
		if typeRef != nil && ast.IsIdentifier(typeRef.TypeName) {
			return typeRef.TypeName.AsIdentifier().Text
		}
	}
	return ""
}

// typesAreEqual checks if two types are exactly the same
func typesAreEqual(tc *checker.Checker, type1, type2 *checker.Type) bool {
	if type1 == nil || type2 == nil {
		return false
	}

	// Check if they're the exact same type instance
	// This is stricter than assignability - we want to catch truly unnecessary assertions
	return type1 == type2
}

// NoUnnecessaryTypeAssertionRule implements the no-unnecessary-type-assertion rule
// Disallow type assertions that don't change type
var NoUnnecessaryTypeAssertionRule = rule.CreateRule(rule.Rule{
	Name: "no-unnecessary-type-assertion",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := parseOptions(options)
	_ = utils.Ref(opts) // Suppress unused warning for now

	checkAssertion := func(node *ast.Node) {
		// This rule requires type information
		if ctx.TypeChecker == nil {
			return
		}

		var exprNode *ast.Node
		var typeNode *ast.Node

		// Handle both <Type>expr and expr as Type syntax
		if node.Kind == ast.KindTypeAssertionExpression {
			typeAssertion := node.AsTypeAssertion()
			if typeAssertion == nil {
				return
			}
			exprNode = typeAssertion.Expression
			typeNode = typeAssertion.Type
		} else if node.Kind == ast.KindAsExpression {
			asExpr := node.AsAsExpression()
			if asExpr == nil {
				return
			}
			exprNode = asExpr.Expression
			typeNode = asExpr.Type
		} else {
			return
		}

		if exprNode == nil || typeNode == nil {
			return
		}

		// Check if the asserted type is in the ignore list
		assertedTypeName := getTypeName(typeNode)
		if isTypeInIgnoreList(assertedTypeName, opts.TypesToIgnore) {
			return
		}

		// Get the type of the expression
		exprType := ctx.TypeChecker.GetTypeAtLocation(exprNode)
		if exprType == nil {
			return
		}

		// Get the asserted type
		assertedType := ctx.TypeChecker.GetTypeFromTypeNode(typeNode)
		if assertedType == nil {
			return
		}

		// Check if types are equal
		if typesAreEqual(ctx.TypeChecker, exprType, assertedType) {
			// The assertion is unnecessary - suggest removal
			sourceText := ctx.SourceFile.Text()
			exprRange := utils.TrimNodeTextRange(ctx.SourceFile, exprNode)
			exprText := sourceText[exprRange.Pos():exprRange.End()]

			ctx.ReportNodeWithFixes(node, buildUnnecessaryAssertionMessage(),
				rule.RuleFixReplace(ctx.SourceFile, node, exprText))
		}
	}

	return rule.RuleListeners{
		ast.KindTypeAssertionExpression: checkAssertion,
		ast.KindAsExpression:            checkAssertion,
		// Note: Non-null assertion checking is complex due to TypeScript's control flow analysis
		// Disabling it for now to avoid false positives
		// TODO: Implement proper non-null assertion checking that accounts for control flow narrowing
	}
}
