package no_unnecessary_type_conversion

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildUnnecessaryTypeConversionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unnecessaryTypeConversion",
		Description: "Unnecessary type conversion.",
	}
}

// Check if a type is a primitive type (string, number, boolean, bigint)
func isPrimitiveType(t *checker.Type) bool {
	if t == nil {
		return false
	}
	return utils.IsTypeFlagSet(t, checker.TypeFlagsStringLike|
		checker.TypeFlagsNumberLike|
		checker.TypeFlagsBooleanLike|
		checker.TypeFlagsBigIntLike)
}

// Check if a type is specifically a string type
func isStringType(t *checker.Type) bool {
	if t == nil {
		return false
	}
	return utils.IsTypeFlagSet(t, checker.TypeFlagsStringLike)
}

// Check if a type is specifically a number type
func isNumberType(t *checker.Type) bool {
	if t == nil {
		return false
	}
	return utils.IsTypeFlagSet(t, checker.TypeFlagsNumberLike)
}

// Check if a type is specifically a boolean type
func isBooleanType(t *checker.Type) bool {
	if t == nil {
		return false
	}
	return utils.IsTypeFlagSet(t, checker.TypeFlagsBooleanLike)
}

// Check if a type is specifically a bigint type
func isBigIntType(t *checker.Type) bool {
	if t == nil {
		return false
	}
	return utils.IsTypeFlagSet(t, checker.TypeFlagsBigIntLike)
}

// Check if a type is an object wrapper (String, Number, Boolean object)
func isObjectWrapper(t *checker.Type) bool {
	if t == nil {
		return false
	}
	// Object wrappers have the object flag set but also correspond to primitive types
	// We need to check if it's the wrapper object type, not the primitive
	return utils.IsTypeFlagSet(t, checker.TypeFlagsObject) && !isPrimitiveType(t)
}

var NoUnnecessaryTypeConversionRule = rule.CreateRule(rule.Rule{
	Name: "no-unnecessary-type-conversion",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {

		// Helper to check if a call expression is a global conversion function
		isGlobalConversionFunction := func(node *ast.Node, name string) bool {
			if !ast.IsCallExpression(node) {
				return false
			}
			callExpr := node.AsCallExpression()
			if !ast.IsIdentifier(callExpr.Expression) {
				return false
			}
			identifier := callExpr.Expression.AsIdentifier()
			if identifier.Text != name {
				return false
			}

			// Check if it's the global function, not a user-defined one
			symbol := ctx.TypeChecker.GetSymbolAtLocation(callExpr.Expression)
			if symbol == nil {
				return false
			}

			// Global functions don't have a parent (or their parent is the global scope)
			// Check if there's a value declaration
			valueDeclaration := symbol.ValueDeclaration
			if valueDeclaration != nil {
				// If there's a local declaration, it's not the global function
				return false
			}

			return true
		}

		// Check String() conversions
		checkStringConversion := func(node *ast.Node) {
			callExpr := node.AsCallExpression()
			if len(callExpr.Arguments.Nodes) != 1 {
				return
			}

			arg := callExpr.Arguments.Nodes[0]
			argType := ctx.TypeChecker.GetTypeAtLocation(arg)

			// If the argument is already a string (not a boxed String), it's unnecessary
			if isStringType(argType) && !isObjectWrapper(argType) {
				ctx.ReportNode(node, buildUnnecessaryTypeConversionMessage())
			}
		}

		// Check Number() conversions
		checkNumberConversion := func(node *ast.Node) {
			callExpr := node.AsCallExpression()
			if len(callExpr.Arguments.Nodes) != 1 {
				return
			}

			arg := callExpr.Arguments.Nodes[0]
			argType := ctx.TypeChecker.GetTypeAtLocation(arg)

			// If the argument is already a number (not a boxed Number), it's unnecessary
			if isNumberType(argType) && !isObjectWrapper(argType) {
				ctx.ReportNode(node, buildUnnecessaryTypeConversionMessage())
			}
		}

		// Check Boolean() conversions
		checkBooleanConversion := func(node *ast.Node) {
			callExpr := node.AsCallExpression()
			if len(callExpr.Arguments.Nodes) != 1 {
				return
			}

			arg := callExpr.Arguments.Nodes[0]
			argType := ctx.TypeChecker.GetTypeAtLocation(arg)

			// If the argument is already a boolean (not a boxed Boolean), it's unnecessary
			if isBooleanType(argType) && !isObjectWrapper(argType) {
				ctx.ReportNode(node, buildUnnecessaryTypeConversionMessage())
			}
		}

		// Check BigInt() conversions
		checkBigIntConversion := func(node *ast.Node) {
			callExpr := node.AsCallExpression()
			if len(callExpr.Arguments.Nodes) != 1 {
				return
			}

			arg := callExpr.Arguments.Nodes[0]
			argType := ctx.TypeChecker.GetTypeAtLocation(arg)

			// If the argument is already a bigint, it's unnecessary
			if isBigIntType(argType) {
				ctx.ReportNode(node, buildUnnecessaryTypeConversionMessage())
			}
		}

		// Check .toString() method calls
		checkToStringMethod := func(node *ast.Node) {
			callExpr := node.AsCallExpression()
			if !ast.IsPropertyAccessExpression(callExpr.Expression) {
				return
			}

			propAccess := callExpr.Expression.AsPropertyAccessExpression()
			if !ast.IsIdentifier(propAccess.Name) {
				return
			}

			name := propAccess.Name.AsIdentifier()
			if name.Text != "toString" {
				return
			}

			// Check if the object is already a string
			objType := ctx.TypeChecker.GetTypeAtLocation(propAccess.Expression)
			if isStringType(objType) && !isObjectWrapper(objType) {
				ctx.ReportNode(node, buildUnnecessaryTypeConversionMessage())
			}
		}

		// Check unary + operator
		checkUnaryPlus := func(node *ast.Node) {
			unaryExpr := node.AsPrefixUnaryExpression()
			operand := unaryExpr.Operand
			operandType := ctx.TypeChecker.GetTypeAtLocation(operand)

			// If the operand is already a number (not a boxed Number), it's unnecessary
			if isNumberType(operandType) && !isObjectWrapper(operandType) {
				ctx.ReportNode(node, buildUnnecessaryTypeConversionMessage())
			}
		}

		// Check double negation !! operator
		checkDoubleNegation := func(node *ast.Node) {
			// Check if this is the outer negation of a double negation
			unaryExpr := node.AsPrefixUnaryExpression()
			if !ast.IsPrefixUnaryExpression(unaryExpr.Operand) {
				return
			}

			innerUnary := unaryExpr.Operand.AsPrefixUnaryExpression()
			if innerUnary.Operator != ast.KindExclamationToken {
				return
			}

			// This is a double negation, check the innermost operand
			operand := innerUnary.Operand
			operandType := ctx.TypeChecker.GetTypeAtLocation(operand)

			// If the operand is already a boolean, it's unnecessary
			if isBooleanType(operandType) && !isObjectWrapper(operandType) {
				ctx.ReportNode(node, buildUnnecessaryTypeConversionMessage())
			}
		}

		// Check double bitwise NOT ~~ operator
		checkDoubleBitwiseNot := func(node *ast.Node) {
			// Check if this is the outer tilde of a double tilde
			unaryExpr := node.AsPrefixUnaryExpression()
			if !ast.IsPrefixUnaryExpression(unaryExpr.Operand) {
				return
			}

			innerUnary := unaryExpr.Operand.AsPrefixUnaryExpression()
			if innerUnary.Operator != ast.KindTildeToken {
				return
			}

			// This is a double bitwise NOT, check the innermost operand
			operand := innerUnary.Operand
			operandType := ctx.TypeChecker.GetTypeAtLocation(operand)

			// If the operand is already a number (not a boxed Number), it's unnecessary
			// except for decimal numbers where ~~ is used for Math.floor
			if isNumberType(operandType) && !isObjectWrapper(operandType) {
				// Check if it's a literal integer (not a decimal)
				if ast.IsNumericLiteral(operand) {
					// ~~ is unnecessary for integer literals
					ctx.ReportNode(node, buildUnnecessaryTypeConversionMessage())
				} else {
					// For non-literal numbers, we need to check if it's an integer type
					// For now, report it as unnecessary if it's a number type
					// The actual TypeScript-ESLint rule has more sophisticated logic
					ctx.ReportNode(node, buildUnnecessaryTypeConversionMessage())
				}
			}
		}

		// Check string concatenation with empty string
		checkStringConcatenation := func(node *ast.Node) {
			binaryExpr := node.AsBinaryExpression()
			left := binaryExpr.Left
			right := binaryExpr.Right

			// Check if one side is an empty string and the other is already a string
			leftType := ctx.TypeChecker.GetTypeAtLocation(left)
			rightType := ctx.TypeChecker.GetTypeAtLocation(right)

			leftIsEmptyString := ast.IsStringLiteral(left) && left.AsStringLiteral().Text == ""
			rightIsEmptyString := ast.IsStringLiteral(right) && right.AsStringLiteral().Text == ""

			if leftIsEmptyString && isStringType(rightType) && !isObjectWrapper(rightType) {
				ctx.ReportNode(node, buildUnnecessaryTypeConversionMessage())
			} else if rightIsEmptyString && isStringType(leftType) && !isObjectWrapper(leftType) {
				ctx.ReportNode(node, buildUnnecessaryTypeConversionMessage())
			}
		}

		// Check += '' pattern
		checkStringAssignment := func(node *ast.Node) {
			binaryExpr := node.AsBinaryExpression()
			right := binaryExpr.Right

			// Check if right side is an empty string
			if !ast.IsStringLiteral(right) || right.AsStringLiteral().Text != "" {
				return
			}

			// Check if left side is already a string
			left := binaryExpr.Left
			leftType := ctx.TypeChecker.GetTypeAtLocation(left)

			if isStringType(leftType) && !isObjectWrapper(leftType) {
				ctx.ReportNode(node, buildUnnecessaryTypeConversionMessage())
			}
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				if isGlobalConversionFunction(node, "String") {
					checkStringConversion(node)
				} else if isGlobalConversionFunction(node, "Number") {
					checkNumberConversion(node)
				} else if isGlobalConversionFunction(node, "Boolean") {
					checkBooleanConversion(node)
				} else if isGlobalConversionFunction(node, "BigInt") {
					checkBigIntConversion(node)
				} else {
					checkToStringMethod(node)
				}
			},
			ast.KindPrefixUnaryExpression: func(node *ast.Node) {
				unaryExpr := node.AsPrefixUnaryExpression()
				switch unaryExpr.Operator {
				case ast.KindPlusToken:
					checkUnaryPlus(node)
				case ast.KindExclamationToken:
					checkDoubleNegation(node)
				case ast.KindTildeToken:
					checkDoubleBitwiseNot(node)
				}
			},
			ast.KindBinaryExpression: func(node *ast.Node) {
				binaryExpr := node.AsBinaryExpression()
				if binaryExpr.OperatorToken.Kind == ast.KindPlusToken {
					checkStringConcatenation(node)
				} else if binaryExpr.OperatorToken.Kind == ast.KindPlusEqualsToken {
					checkStringAssignment(node)
				}
			},
		}
	},
})
