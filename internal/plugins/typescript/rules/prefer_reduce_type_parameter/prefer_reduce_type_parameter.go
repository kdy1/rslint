package prefer_reduce_type_parameter

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildPreferTypeParameterMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferTypeParameter",
		Description: "Unnecessary assertion: Array#reduce accepts a type parameter for the default value.",
	}
}

var PreferReduceTypeParameterRule = rule.CreateRule(rule.Rule{
	Name: "prefer-reduce-type-parameter",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// Helper to check if all parts of a type are arrays or tuples
		isArrayType := func(t *checker.Type) bool {
			return utils.Every(utils.UnionTypeParts(t), func(unionPart *checker.Type) bool {
				return utils.Every(utils.IntersectionTypeParts(unionPart), func(intersectionPart *checker.Type) bool {
					return checker.Checker_isArrayOrTupleType(ctx.TypeChecker, intersectionPart)
				})
			})
		}

		// Helper to extract type and expression from type assertions
		getAssertionInfo := func(node *ast.Node) (typeNode *ast.Node, expression *ast.Node) {
			switch node.Kind {
			case ast.KindAsExpression:
				asExpr := node.AsAsExpression()
				if asExpr != nil {
					return asExpr.Type, asExpr.Expression
				}
			case ast.KindTypeAssertionExpression:
				typeAssertion := node.AsTypeAssertion()
				if typeAssertion != nil {
					return typeAssertion.Type, typeAssertion.Expression
				}
			}
			return nil, nil
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				expr := node.AsCallExpression()

				// Check if there are at least 2 arguments
				if len(expr.Arguments.Nodes) < 2 {
					return
				}

				// Get the callee (should be a property access like array.reduce)
				callee := expr.Expression
				if !ast.IsAccessExpression(callee) {
					return
				}

				// Check if the property name is 'reduce'
				propertyName, found := checker.Checker_getAccessedPropertyName(ctx.TypeChecker, callee)
				if !found || propertyName != "reduce" {
					return
				}

				// Get the second argument (initializer)
				secondArg := expr.Arguments.Nodes[1]

				// Check if it's a type assertion
				if secondArg.Kind != ast.KindAsExpression && secondArg.Kind != ast.KindTypeAssertionExpression {
					return
				}

				// Extract type and expression from the assertion
				assertionType, assertionExpr := getAssertionInfo(secondArg)
				if assertionType == nil || assertionExpr == nil {
					return
				}

				// Get the type of the expression inside the assertion
				initializerType := ctx.TypeChecker.GetTypeAtLocation(assertionExpr)
				assertedType := ctx.TypeChecker.GetTypeAtLocation(assertionType)

				// Check if the assertion is necessary
				// Don't report if the resulting fix will be a type error
				if !checker.Checker_isTypeAssignableTo(ctx.TypeChecker, initializerType, assertedType) {
					return
				}

				// Get the type of the object the reduce is being called on
				calleeObjType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, callee.Expression())

				// Check if the object is an array type
				if !isArrayType(calleeObjType) {
					return
				}

				// Build the fix
				var fixes []rule.RuleFix

				// Get the text of the type
				typeText := utils.GetNodeText(ctx.SourceFile, assertionType)

				// Remove the type assertion wrapper
				// For 'as' syntax: remove from the end of expression to the end of the assertion
				// For angle bracket syntax: remove from the start to the start of the expression
				if secondArg.Kind == ast.KindAsExpression {
					// Remove " as Type" part, keep the expression
					exprEnd := assertionExpr.End()
					assertionEnd := secondArg.End()
					fixes = append(fixes, rule.RuleFixRemoveRange(utils.NewTextRange(exprEnd, assertionEnd)))
				} else {
					// Remove "<Type>" part at the beginning, keep the expression
					assertionStart := secondArg.Pos()
					exprStart := assertionExpr.Pos()
					fixes = append(fixes, rule.RuleFixRemoveRange(utils.NewTextRange(assertionStart, exprStart)))
				}

				// Add type parameter if not already present
				if expr.TypeArguments == nil || len(expr.TypeArguments.Nodes) == 0 {
					// Insert the type parameter after the callee
					calleeEnd := callee.End()
					fixes = append(fixes, rule.RuleFixInsertTextAt(calleeEnd, "<"+typeText+">"))
				}

				ctx.ReportNodeWithFix(secondArg, buildPreferTypeParameterMessage(), fixes...)
			},
		}
	},
})
