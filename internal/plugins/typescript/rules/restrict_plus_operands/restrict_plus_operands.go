package restrict_plus_operands

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildBigintAndNumberMessage(left, right string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "bigintAndNumber",
		Description: fmt.Sprintf("Numeric '+' operations must either be both bigints or both numbers. Got `%s` + `%s`.", left, right),
	}
}

func buildInvalidMessage(t, stringLike string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "invalid",
		Description: fmt.Sprintf("Invalid operand for a '+' operation. Operands must each be a number or %s. Got `%s`.", stringLike, t),
	}
}

func buildMismatchedMessage(left, right, stringLike string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "mismatched",
		Description: fmt.Sprintf("Operands of '+' operations must be a number or %s. Got `%s` + `%s`.", stringLike, left, right),
	}
}

func getDefaultOptions() map[string]bool {
	return map[string]bool{
		"allowAny":               true,
		"allowBoolean":           true,
		"allowNullish":           true,
		"allowNumberAndString":   true,
		"allowRegExp":            true,
		"skipCompoundAssignments": false,
	}
}

func mergeOptions(options any) map[string]bool {
	defaults := getDefaultOptions()
	if options == nil {
		return defaults
	}

	if opts, ok := options.(map[string]any); ok {
		if v, ok := opts["allowAny"].(bool); ok {
			defaults["allowAny"] = v
		}
		if v, ok := opts["allowBoolean"].(bool); ok {
			defaults["allowBoolean"] = v
		}
		if v, ok := opts["allowNullish"].(bool); ok {
			defaults["allowNullish"] = v
		}
		if v, ok := opts["allowNumberAndString"].(bool); ok {
			defaults["allowNumberAndString"] = v
		}
		if v, ok := opts["allowRegExp"].(bool); ok {
			defaults["allowRegExp"] = v
		}
		if v, ok := opts["skipCompoundAssignments"].(bool); ok {
			defaults["skipCompoundAssignments"] = v
		}
	}

	return defaults
}

func isTypeFlagSetInUnion(t *checker.Type, flag checker.TypeFlags) bool {
	for _, subType := range utils.UnionTypeParts(t) {
		if utils.IsTypeFlagSet(subType, flag) {
			return true
		}
	}
	return false
}

func isDeeplyObjectType(t *checker.Type) bool {
	if utils.IsIntersectionType(t) {
		for _, constituent := range utils.IntersectionTypeParts(t) {
			if !utils.IsObjectType(constituent) {
				return false
			}
		}
		return true
	}

	for _, constituent := range utils.UnionTypeParts(t) {
		if !utils.IsObjectType(constituent) {
			return false
		}
	}
	return true
}

var RestrictPlusOperandsRule = rule.CreateRule(rule.Rule{
	Name: "restrict-plus-operands",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := mergeOptions(options)

		// Build the stringLike message part based on options
		var stringLikes []string
		if opts["allowAny"] {
			stringLikes = append(stringLikes, "`any`")
		}
		if opts["allowBoolean"] {
			stringLikes = append(stringLikes, "`boolean`")
		}
		if opts["allowNullish"] {
			stringLikes = append(stringLikes, "`null`")
		}
		if opts["allowRegExp"] {
			stringLikes = append(stringLikes, "`RegExp`")
		}
		if opts["allowNullish"] {
			stringLikes = append(stringLikes, "`undefined`")
		}

		var stringLike string
		if len(stringLikes) == 0 {
			stringLike = "string"
		} else if len(stringLikes) == 1 {
			stringLike = fmt.Sprintf("string, allowing a string + %s", stringLikes[0])
		} else {
			stringLike = fmt.Sprintf("string, allowing a string + any of: %s", strings.Join(stringLikes, ", "))
		}

		getTypeConstrained := func(node *ast.Node) *checker.Type {
			return checker.Checker_getBaseTypeOfLiteralType(
				ctx.TypeChecker,
				utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, node),
			)
		}

		checkPlusOperands := func(node *ast.Node, left *ast.Node, right *ast.Node) {
			leftType := getTypeConstrained(left)
			rightType := getTypeConstrained(right)

			// If both types are the same and are bigint, number, or string, allow it
			if leftType == rightType &&
				utils.IsTypeFlagSet(leftType, checker.TypeFlagsBigIntLike|checker.TypeFlagsNumberLike|checker.TypeFlagsStringLike) {
				return
			}

			hadIndividualComplaint := false

			// Check each operand individually for invalid types
			for _, info := range []struct {
				baseNode  *ast.Node
				baseType  *checker.Type
				otherType *checker.Type
			}{
				{left, leftType, rightType},
				{right, rightType, leftType},
			} {
				baseNode := info.baseNode
				baseType := info.baseType
				otherType := info.otherType

				// Check for disallowed type flags in union
				if isTypeFlagSetInUnion(baseType, checker.TypeFlagsESSymbolLike|checker.TypeFlagsNever|checker.TypeFlagsUnknown) ||
					(!opts["allowAny"] && isTypeFlagSetInUnion(baseType, checker.TypeFlagsAny)) ||
					(!opts["allowBoolean"] && isTypeFlagSetInUnion(baseType, checker.TypeFlagsBooleanLike)) ||
					(!opts["allowNullish"] && utils.IsTypeFlagSet(baseType, checker.TypeFlagsNull|checker.TypeFlagsUndefined)) {
					ctx.ReportNode(baseNode, buildInvalidMessage(ctx.TypeChecker.TypeToString(baseType), stringLike))
					hadIndividualComplaint = true
					continue
				}

				// Check for RegExp and deeply object types
				for _, subBaseType := range utils.UnionTypeParts(baseType) {
					typeName := utils.GetTypeName(ctx.TypeChecker, subBaseType)
					shouldReport := false

					if typeName == "RegExp" {
						// Report if RegExp is not allowed, or if the other type is number
						if !opts["allowRegExp"] || utils.IsTypeFlagSet(otherType, checker.TypeFlagsNumberLike) {
							shouldReport = true
						}
					} else {
						// Report if it's any (when not allowed) or deeply object type
						if (!opts["allowAny"] && utils.IsTypeAnyType(subBaseType)) || isDeeplyObjectType(subBaseType) {
							shouldReport = true
						}
					}

					if shouldReport {
						ctx.ReportNode(baseNode, buildInvalidMessage(ctx.TypeChecker.TypeToString(subBaseType), stringLike))
						hadIndividualComplaint = true
						break
					}
				}
			}

			if hadIndividualComplaint {
				return
			}

			// Check for type mismatches
			for _, info := range []struct {
				baseType  *checker.Type
				otherType *checker.Type
			}{
				{leftType, rightType},
				{rightType, leftType},
			} {
				baseType := info.baseType
				otherType := info.otherType

				// Check for string + number/bigint when not allowed
				if !opts["allowNumberAndString"] &&
					isTypeFlagSetInUnion(baseType, checker.TypeFlagsStringLike) &&
					isTypeFlagSetInUnion(otherType, checker.TypeFlagsNumberLike|checker.TypeFlagsBigIntLike) {
					ctx.ReportNode(node, buildMismatchedMessage(
						ctx.TypeChecker.TypeToString(leftType),
						ctx.TypeChecker.TypeToString(rightType),
						stringLike,
					))
					return
				}

				// Check for number + bigint (never allowed)
				if isTypeFlagSetInUnion(baseType, checker.TypeFlagsNumberLike) &&
					isTypeFlagSetInUnion(otherType, checker.TypeFlagsBigIntLike) {
					ctx.ReportNode(node, buildBigintAndNumberMessage(
						ctx.TypeChecker.TypeToString(leftType),
						ctx.TypeChecker.TypeToString(rightType),
					))
					return
				}
			}
		}

		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				binExpr := node.AsBinaryExpression()
				if binExpr.OperatorToken.Kind == ast.KindPlusToken {
					checkPlusOperands(node, binExpr.Left, binExpr.Right)
				} else if !opts["skipCompoundAssignments"] && binExpr.OperatorToken.Kind == ast.KindPlusEqualsToken {
					checkPlusOperands(node, binExpr.Left, binExpr.Right)
				}
			},
		}
	},
})
