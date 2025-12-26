package strict_boolean_expressions

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Message builders for various condition error types
func buildConditionErrorAnyMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionErrorAny",
		Description: "Unexpected any value in {{context}}. An explicit comparison or type conversion is required.",
	}
}

func buildConditionErrorNullableBooleanMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionErrorNullableBoolean",
		Description: "Unexpected nullable boolean value in {{context}}. Please handle the nullish case explicitly.",
	}
}

func buildConditionErrorNullableEnumMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionErrorNullableEnum",
		Description: "Unexpected nullable enum value in {{context}}. Please handle the nullish/zero/NaN cases explicitly.",
	}
}

func buildConditionErrorNullableNumberMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionErrorNullableNumber",
		Description: "Unexpected nullable number value in {{context}}. Please handle the nullish/zero/NaN cases explicitly.",
	}
}

func buildConditionErrorNullableObjectMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionErrorNullableObject",
		Description: "Unexpected nullable object value in {{context}}. An explicit null check is required.",
	}
}

func buildConditionErrorNullableStringMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionErrorNullableString",
		Description: "Unexpected nullable string value in {{context}}. Please handle the nullish/empty cases explicitly.",
	}
}

func buildConditionErrorNullishMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionErrorNullish",
		Description: "Unexpected nullish value in conditional. The condition is always false.",
	}
}

func buildConditionErrorNumberMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionErrorNumber",
		Description: "Unexpected number value in {{context}}. An explicit zero/NaN check is required.",
	}
}

func buildConditionErrorObjectMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionErrorObject",
		Description: "Unexpected object value in {{context}}. The condition is always true.",
	}
}

func buildConditionErrorOtherMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionErrorOther",
		Description: "Unexpected value in conditional. A boolean expression is required.",
	}
}

func buildConditionErrorStringMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionErrorString",
		Description: "Unexpected string value in {{context}}. An explicit empty string check is required.",
	}
}

func buildNoStrictNullCheckMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noStrictNullCheck",
		Description: "This rule requires the `strictNullChecks` compiler option to be turned on to function correctly.",
	}
}

func buildPredicateCannotBeAsyncMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "predicateCannotBeAsync",
		Description: "Predicate function should not be 'async'; expected a boolean return type.",
	}
}

// Suggestion message builders
func buildConditionFixCastBooleanMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionFixCastBoolean",
		Description: "Explicitly convert value to a boolean (`Boolean(value)`)",
	}
}

func buildConditionFixCompareArrayLengthNonzeroMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionFixCompareArrayLengthNonzero",
		Description: "Change condition to check array's length (`value.length > 0`)",
	}
}

func buildConditionFixCompareArrayLengthZeroMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionFixCompareArrayLengthZero",
		Description: "Change condition to check array's length (`value.length === 0`)",
	}
}

func buildConditionFixCompareEmptyStringMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionFixCompareEmptyString",
		Description: "Change condition to check for empty string (`value !== \"\"`)",
	}
}

func buildConditionFixCompareFalseMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionFixCompareFalse",
		Description: "Change condition to check if false (`value === false`)",
	}
}

func buildConditionFixCompareNaNMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionFixCompareNaN",
		Description: "Change condition to check for NaN (`!Number.isNaN(value)`)",
	}
}

func buildConditionFixCompareNullishMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionFixCompareNullish",
		Description: "Change condition to check for null/undefined (`value != null`)",
	}
}

func buildConditionFixCompareStringLengthMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionFixCompareStringLength",
		Description: "Change condition to check string's length (`value.length !== 0`)",
	}
}

func buildConditionFixCompareTrueMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionFixCompareTrue",
		Description: "Change condition to check if true (`value === true`)",
	}
}

func buildConditionFixCompareZeroMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionFixCompareZero",
		Description: "Change condition to check for 0 (`value !== 0`)",
	}
}

func buildConditionFixDefaultEmptyStringMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionFixDefaultEmptyString",
		Description: "Explicitly treat nullish value the same as an empty string (`value ?? \"\"`)",
	}
}

func buildConditionFixDefaultFalseMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionFixDefaultFalse",
		Description: "Explicitly treat nullish value the same as false (`value ?? false`)",
	}
}

func buildConditionFixDefaultZeroMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionFixDefaultZero",
		Description: "Explicitly treat nullish value the same as 0 (`value ?? 0`)",
	}
}

func buildExplicitBooleanReturnTypeMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "explicitBooleanReturnType",
		Description: "Add an explicit `boolean` return type annotation.",
	}
}

// Options for the rule
type StrictBooleanExpressionsOptions struct {
	AllowAny                                               *bool
	AllowNullableBoolean                                   *bool
	AllowNullableEnum                                      *bool
	AllowNullableNumber                                    *bool
	AllowNullableObject                                    *bool
	AllowNullableString                                    *bool
	AllowNumber                                            *bool
	AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing *bool
	AllowString                                            *bool
}

// VariantType represents the types we care about
type VariantType string

const (
	VariantTypeAny           VariantType = "any"
	VariantTypeBoolean       VariantType = "boolean"
	VariantTypeEnum          VariantType = "enum"
	VariantTypeNever         VariantType = "never"
	VariantTypeNullish       VariantType = "nullish"
	VariantTypeNumber        VariantType = "number"
	VariantTypeObject        VariantType = "object"
	VariantTypeString        VariantType = "string"
	VariantTypeTruthyBoolean VariantType = "truthy boolean"
	VariantTypeTruthyNumber  VariantType = "truthy number"
	VariantTypeTruthyString  VariantType = "truthy string"
)

var StrictBooleanExpressionsRule = rule.CreateRule(rule.Rule{
	Name: "strict-boolean-expressions",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts, ok := options.(StrictBooleanExpressionsOptions)
		if !ok {
			opts = StrictBooleanExpressionsOptions{}
		}

		// Set defaults
		if opts.AllowAny == nil {
			opts.AllowAny = utils.Ref(false)
		}
		if opts.AllowNullableBoolean == nil {
			opts.AllowNullableBoolean = utils.Ref(false)
		}
		if opts.AllowNullableEnum == nil {
			opts.AllowNullableEnum = utils.Ref(false)
		}
		if opts.AllowNullableNumber == nil {
			opts.AllowNullableNumber = utils.Ref(false)
		}
		if opts.AllowNullableObject == nil {
			opts.AllowNullableObject = utils.Ref(true)
		}
		if opts.AllowNullableString == nil {
			opts.AllowNullableString = utils.Ref(false)
		}
		if opts.AllowNumber == nil {
			opts.AllowNumber = utils.Ref(true)
		}
		if opts.AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing == nil {
			opts.AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing = utils.Ref(false)
		}
		if opts.AllowString == nil {
			opts.AllowString = utils.Ref(true)
		}

		// Check strictNullChecks
		compilerOptions := ctx.Program.GetCompilerOptions()
		isStrictNullChecks := compilerOptions.StrictNullChecks != nil && *compilerOptions.StrictNullChecks
		if !isStrictNullChecks && compilerOptions.Strict != nil && *compilerOptions.Strict {
			isStrictNullChecks = true
		}

		if !isStrictNullChecks && !*opts.AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing {
			ctx.ReportRange(
				scanner.TextRange{Pos: 0, End: 0},
				buildNoStrictNullCheckMessage(),
			)
			return rule.RuleListeners{}
		}

		traversedNodes := make(map[*ast.Node]bool)

		// Helper: Inspect variant types
		inspectVariantTypes := func(types []*checker.Type) map[VariantType]bool {
			variantTypes := make(map[VariantType]bool)

			// Check for nullish
			for _, t := range types {
				if utils.IsTypeFlagSet(t, checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) {
					variantTypes[VariantTypeNullish] = true
					break
				}
			}

			// Check for booleans
			var booleans []*checker.Type
			for _, t := range types {
				if utils.IsTypeFlagSet(t, checker.TypeFlagsBooleanLike) {
					booleans = append(booleans, t)
				}
			}

			if len(booleans) == 1 {
				// Check if it's true literal
				if utils.IsTypeFlagSet(booleans[0], checker.TypeFlagsBooleanLiteral) {
					intrinsicName := checker.Type_getIntrinsicName(booleans[0])
					if intrinsicName == "true" {
						variantTypes[VariantTypeTruthyBoolean] = true
					} else {
						variantTypes[VariantTypeBoolean] = true
					}
				} else {
					variantTypes[VariantTypeBoolean] = true
				}
			} else if len(booleans) == 2 {
				variantTypes[VariantTypeBoolean] = true
			}

			// Check for strings
			var strings []*checker.Type
			for _, t := range types {
				if utils.IsTypeFlagSet(t, checker.TypeFlagsStringLike) {
					strings = append(strings, t)
				}
			}

			if len(strings) > 0 {
				allTruthy := true
				for _, t := range strings {
					if utils.IsTypeFlagSet(t, checker.TypeFlagsStringLiteral) {
						value := checker.Type_getStringLiteralValue(t)
						if value == "" {
							allTruthy = false
							break
						}
					} else {
						allTruthy = false
						break
					}
				}
				if allTruthy {
					variantTypes[VariantTypeTruthyString] = true
				} else {
					variantTypes[VariantTypeString] = true
				}
			}

			// Check for numbers
			var numbers []*checker.Type
			for _, t := range types {
				if utils.IsTypeFlagSet(t, checker.TypeFlagsNumberLike|checker.TypeFlagsBigIntLike) {
					numbers = append(numbers, t)
				}
			}

			if len(numbers) > 0 {
				allTruthy := true
				for _, t := range numbers {
					if utils.IsTypeFlagSet(t, checker.TypeFlagsNumberLiteral) {
						value := checker.Type_getNumberLiteralValue(t)
						if value == 0 {
							allTruthy = false
							break
						}
					} else {
						allTruthy = false
						break
					}
				}
				if allTruthy {
					variantTypes[VariantTypeTruthyNumber] = true
				} else {
					variantTypes[VariantTypeNumber] = true
				}
			}

			// Check for enums
			for _, t := range types {
				if utils.IsTypeFlagSet(t, checker.TypeFlagsEnumLike) {
					variantTypes[VariantTypeEnum] = true
					break
				}
			}

			// Check for objects
			for _, t := range types {
				flags := checker.TypeFlagsNull | checker.TypeFlagsUndefined | checker.TypeFlagsVoid |
					checker.TypeFlagsBooleanLike | checker.TypeFlagsStringLike |
					checker.TypeFlagsNumberLike | checker.TypeFlagsBigIntLike |
					checker.TypeFlagsTypeParameter | checker.TypeFlagsAny |
					checker.TypeFlagsUnknown | checker.TypeFlagsNever

				if !utils.IsTypeFlagSet(t, flags) {
					// Check for branded boolean
					isBranded := false
					if utils.IsTypeFlagSet(t, checker.TypeFlagsIntersection) {
						for _, part := range utils.IntersectionTypeParts(t) {
							if utils.IsTypeFlagSet(part, checker.TypeFlagsBoolean|checker.TypeFlagsBooleanLiteral) {
								isBranded = true
								break
							}
						}
					}
					if isBranded {
						variantTypes[VariantTypeBoolean] = true
					} else {
						variantTypes[VariantTypeObject] = true
					}
					break
				}
			}

			// Check for any/unknown/type parameters
			for _, t := range types {
				if utils.IsTypeFlagSet(t, checker.TypeFlagsTypeParameter|checker.TypeFlagsAny|checker.TypeFlagsUnknown) {
					variantTypes[VariantTypeAny] = true
					break
				}
			}

			// Check for never
			for _, t := range types {
				if utils.IsTypeFlagSet(t, checker.TypeFlagsNever) {
					variantTypes[VariantTypeNever] = true
					break
				}
			}

			return variantTypes
		}

		// Helper: Determine report type
		determineReportType := func(types map[VariantType]bool) *rule.RuleMessage {
			is := func(wantedTypes ...VariantType) bool {
				if len(types) != len(wantedTypes) {
					return false
				}
				for _, wt := range wantedTypes {
					if !types[wt] {
						return false
					}
				}
				return true
			}

			// boolean
			if is(VariantTypeBoolean) || is(VariantTypeTruthyBoolean) {
				return nil
			}

			// never
			if is(VariantTypeNever) {
				return nil
			}

			// nullish
			if is(VariantTypeNullish) {
				msg := buildConditionErrorNullishMessage()
				return &msg
			}

			// Known edge case: boolean `true` and nullish values
			if is(VariantTypeNullish, VariantTypeTruthyBoolean) {
				return nil
			}

			// nullable boolean
			if is(VariantTypeNullish, VariantTypeBoolean) {
				if !*opts.AllowNullableBoolean {
					msg := buildConditionErrorNullableBooleanMessage()
					return &msg
				}
				return nil
			}

			// Known edge cases: truthy primitives and nullish
			if *opts.AllowNumber && is(VariantTypeNullish, VariantTypeTruthyNumber) {
				return nil
			}
			if *opts.AllowString && is(VariantTypeNullish, VariantTypeTruthyString) {
				return nil
			}

			// string
			if is(VariantTypeString) || is(VariantTypeTruthyString) {
				if !*opts.AllowString {
					msg := buildConditionErrorStringMessage()
					return &msg
				}
				return nil
			}

			// nullable string
			if is(VariantTypeNullish, VariantTypeString) {
				if !*opts.AllowNullableString {
					msg := buildConditionErrorNullableStringMessage()
					return &msg
				}
				return nil
			}

			// number
			if is(VariantTypeNumber) || is(VariantTypeTruthyNumber) {
				if !*opts.AllowNumber {
					msg := buildConditionErrorNumberMessage()
					return &msg
				}
				return nil
			}

			// nullable number
			if is(VariantTypeNullish, VariantTypeNumber) {
				if !*opts.AllowNullableNumber {
					msg := buildConditionErrorNullableNumberMessage()
					return &msg
				}
				return nil
			}

			// object
			if is(VariantTypeObject) {
				msg := buildConditionErrorObjectMessage()
				return &msg
			}

			// nullable object
			if is(VariantTypeNullish, VariantTypeObject) {
				if !*opts.AllowNullableObject {
					msg := buildConditionErrorNullableObjectMessage()
					return &msg
				}
				return nil
			}

			// nullable enum
			if is(VariantTypeNullish, VariantTypeNumber, VariantTypeEnum) ||
				is(VariantTypeNullish, VariantTypeString, VariantTypeEnum) ||
				is(VariantTypeNullish, VariantTypeTruthyNumber, VariantTypeEnum) ||
				is(VariantTypeNullish, VariantTypeTruthyString, VariantTypeEnum) ||
				is(VariantTypeNullish, VariantTypeTruthyNumber, VariantTypeTruthyString, VariantTypeEnum) ||
				is(VariantTypeNullish, VariantTypeTruthyNumber, VariantTypeString, VariantTypeEnum) ||
				is(VariantTypeNullish, VariantTypeTruthyString, VariantTypeNumber, VariantTypeEnum) ||
				is(VariantTypeNullish, VariantTypeNumber, VariantTypeString, VariantTypeEnum) {
				if !*opts.AllowNullableEnum {
					msg := buildConditionErrorNullableEnumMessage()
					return &msg
				}
				return nil
			}

			// any
			if is(VariantTypeAny) {
				if !*opts.AllowAny {
					msg := buildConditionErrorAnyMessage()
					return &msg
				}
				return nil
			}

			msg := buildConditionErrorOtherMessage()
			return &msg
		}

		// Helper: Get suggestions for condition error
		getSuggestionsForConditionError := func(node *ast.Node, reportMsg rule.RuleMessage, isNegated bool) []rule.RuleSuggestion {
			var suggestions []rule.RuleSuggestion

			switch reportMsg.Id {
			case "conditionErrorString":
				text := utils.GetSourceText(ctx.SourceFile, node)
				if isNegated {
					suggestions = append(suggestions,
						rule.RuleSuggestion{
							Message: buildConditionFixCompareStringLengthMessage(),
							FixesArr: []rule.RuleFix{
								rule.RuleFixReplaceRange(
									scanner.TextRange{Pos: node.Parent.Pos(), End: node.Parent.End()},
									text+".length === 0",
								),
							},
						},
						rule.RuleSuggestion{
							Message: buildConditionFixCompareEmptyStringMessage(),
							FixesArr: []rule.RuleFix{
								rule.RuleFixReplaceRange(
									scanner.TextRange{Pos: node.Parent.Pos(), End: node.Parent.End()},
									text+` === ""`,
								),
							},
						},
						rule.RuleSuggestion{
							Message: buildConditionFixCastBooleanMessage(),
							FixesArr: []rule.RuleFix{
								rule.RuleFixReplaceRange(
									scanner.TextRange{Pos: node.Parent.Pos(), End: node.Parent.End()},
									"!Boolean("+text+")",
								),
							},
						},
					)
				} else {
					suggestions = append(suggestions,
						rule.RuleSuggestion{
							Message: buildConditionFixCompareStringLengthMessage(),
							FixesArr: []rule.RuleFix{
								rule.RuleFixReplaceRange(
									scanner.TextRange{Pos: node.Pos(), End: node.End()},
									text+".length > 0",
								),
							},
						},
						rule.RuleSuggestion{
							Message: buildConditionFixCompareEmptyStringMessage(),
							FixesArr: []rule.RuleFix{
								rule.RuleFixReplaceRange(
									scanner.TextRange{Pos: node.Pos(), End: node.End()},
									text+` !== ""`,
								),
							},
						},
						rule.RuleSuggestion{
							Message: buildConditionFixCastBooleanMessage(),
							FixesArr: []rule.RuleFix{
								rule.RuleFixReplaceRange(
									scanner.TextRange{Pos: node.Pos(), End: node.End()},
									"Boolean("+text+")",
								),
							},
						},
					)
				}

			case "conditionErrorNumber":
				text := utils.GetSourceText(ctx.SourceFile, node)
				if isNegated {
					suggestions = append(suggestions,
						rule.RuleSuggestion{
							Message: buildConditionFixCompareZeroMessage(),
							FixesArr: []rule.RuleFix{
								rule.RuleFixReplaceRange(
									scanner.TextRange{Pos: node.Parent.Pos(), End: node.Parent.End()},
									text+" === 0",
								),
							},
						},
						rule.RuleSuggestion{
							Message: buildConditionFixCompareNaNMessage(),
							FixesArr: []rule.RuleFix{
								rule.RuleFixReplaceRange(
									scanner.TextRange{Pos: node.Parent.Pos(), End: node.Parent.End()},
									"Number.isNaN("+text+")",
								),
							},
						},
						rule.RuleSuggestion{
							Message: buildConditionFixCastBooleanMessage(),
							FixesArr: []rule.RuleFix{
								rule.RuleFixReplaceRange(
									scanner.TextRange{Pos: node.Parent.Pos(), End: node.Parent.End()},
									"!Boolean("+text+")",
								),
							},
						},
					)
				} else {
					suggestions = append(suggestions,
						rule.RuleSuggestion{
							Message: buildConditionFixCompareZeroMessage(),
							FixesArr: []rule.RuleFix{
								rule.RuleFixReplaceRange(
									scanner.TextRange{Pos: node.Pos(), End: node.End()},
									text+" !== 0",
								),
							},
						},
						rule.RuleSuggestion{
							Message: buildConditionFixCompareNaNMessage(),
							FixesArr: []rule.RuleFix{
								rule.RuleFixReplaceRange(
									scanner.TextRange{Pos: node.Pos(), End: node.End()},
									"!Number.isNaN("+text+")",
								),
							},
						},
						rule.RuleSuggestion{
							Message: buildConditionFixCastBooleanMessage(),
							FixesArr: []rule.RuleFix{
								rule.RuleFixReplaceRange(
									scanner.TextRange{Pos: node.Pos(), End: node.End()},
									"Boolean("+text+")",
								),
							},
						},
					)
				}

			case "conditionErrorNullableObject", "conditionErrorNullableString", "conditionErrorNullableNumber", "conditionErrorNullableBoolean", "conditionErrorNullableEnum":
				text := utils.GetSourceText(ctx.SourceFile, node)
				if isNegated {
					suggestions = append(suggestions,
						rule.RuleSuggestion{
							Message: buildConditionFixCompareNullishMessage(),
							FixesArr: []rule.RuleFix{
								rule.RuleFixReplaceRange(
									scanner.TextRange{Pos: node.Parent.Pos(), End: node.Parent.End()},
									text+" == null",
								),
							},
						},
					)
				} else {
					suggestions = append(suggestions,
						rule.RuleSuggestion{
							Message: buildConditionFixCompareNullishMessage(),
							FixesArr: []rule.RuleFix{
								rule.RuleFixReplaceRange(
									scanner.TextRange{Pos: node.Pos(), End: node.End()},
									text+" != null",
								),
							},
						},
					)
				}

			case "conditionErrorAny":
				text := utils.GetSourceText(ctx.SourceFile, node)
				suggestions = append(suggestions,
					rule.RuleSuggestion{
						Message: buildConditionFixCastBooleanMessage(),
						FixesArr: []rule.RuleFix{
							rule.RuleFixReplaceRange(
								scanner.TextRange{Pos: node.Pos(), End: node.End()},
								"Boolean("+text+")",
							),
						},
					},
				)
			}

			return suggestions
		}

		// Helper: Check node type
		checkNode := func(node *ast.Node, context string) {
			t := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, node)
			types := inspectVariantTypes(utils.UnionTypeParts(t))
			reportMsg := determineReportType(types)

			if reportMsg != nil {
				// Check if parent is logical negation
				isNegated := false
				if node.Parent != nil && ast.IsPrefix(node.Parent) {
					prefix := node.Parent.AsPrefixUnaryExpression()
					if prefix.Operator == scanner.ExclamationToken {
						isNegated = true
					}
				}

				suggestions := getSuggestionsForConditionError(node, *reportMsg, isNegated)

				msg := *reportMsg
				msg.Data = map[string]interface{}{
					"context": context,
				}

				ctx.ReportNodeWithSuggestions(node, msg, suggestions...)
			}
		}

		// Helper: Traverse node
		var traverseNode func(node *ast.Node, isCondition bool)
		traverseNode = func(node *ast.Node, isCondition bool) {
			if node == nil {
				return
			}

			// Prevent checking the same node multiple times
			if traversedNodes[node] {
				return
			}
			traversedNodes[node] = true

			// For logical operators, check operands
			if ast.IsBinary(node) {
				binary := node.AsBinaryExpression()
				if binary.OperatorToken.Kind == scanner.AmpersandAmpersandToken ||
					binary.OperatorToken.Kind == scanner.BarBarToken {
					// Left is always a condition
					traverseNode(binary.Left, true)
					// Right is a condition if parent expression is a condition
					traverseNode(binary.Right, isCondition)
					return
				}
			}

			// Skip if not a condition
			if !isCondition {
				return
			}

			checkNode(node, "conditional")
		}

		// Helper: Check array predicate
		checkArrayPredicate := func(node *ast.Node) {
			if !ast.IsCallExpression(node) {
				return
			}

			callExpr := node.AsCallExpression()
			if len(callExpr.Arguments.Nodes) == 0 {
				return
			}

			// Check if it's an array method with predicate
			if !ast.IsMemberExpression(callExpr.Expression) {
				return
			}

			memberExpr := callExpr.Expression.AsPropertyAccessExpression()
			if memberExpr.Name == nil {
				return
			}

			methodName := utils.GetSourceText(ctx.SourceFile, memberExpr.Name)
			predicateMethods := map[string]bool{
				"filter": true, "find": true, "findIndex": true,
				"some": true, "every": true, "findLast": true, "findLastIndex": true,
			}

			if !predicateMethods[methodName] {
				return
			}

			predicateNode := callExpr.Arguments.Nodes[0]

			// Check if predicate is async
			if ast.IsArrowFunction(predicateNode) && predicateNode.AsArrowFunction().AsteriskToken != nil {
				ctx.ReportNode(predicateNode, buildPredicateCannotBeAsyncMessage())
				return
			}
			if ast.IsFunctionExpression(predicateNode) && predicateNode.AsFunctionExpression().AsteriskToken != nil {
				ctx.ReportNode(predicateNode, buildPredicateCannotBeAsyncMessage())
				return
			}

			// Get return types
			predicateType := ctx.TypeChecker.GetTypeAtLocation(predicateNode)
			signatures := utils.GetCallSignatures(ctx.TypeChecker, predicateType)

			var returnTypes []*checker.Type
			for _, sig := range signatures {
				returnType := checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, sig)

				// Handle type parameters
				if utils.IsTypeFlagSet(returnType, checker.TypeFlagsTypeParameter) {
					constraint := checker.TypeChecker_getBaseConstraintOfType(ctx.TypeChecker, returnType)
					if constraint != nil {
						returnType = constraint
					}
				}

				returnTypes = append(returnTypes, returnType)
			}

			// Flatten union types
			var flattenedTypes []*checker.Type
			seenTypes := make(map[*checker.Type]bool)
			for _, rt := range returnTypes {
				for _, part := range utils.UnionTypeParts(rt) {
					if !seenTypes[part] {
						seenTypes[part] = true
						flattenedTypes = append(flattenedTypes, part)
					}
				}
			}

			types := inspectVariantTypes(flattenedTypes)
			reportMsg := determineReportType(types)

			if reportMsg != nil {
				msg := *reportMsg
				msg.Data = map[string]interface{}{
					"context": "array predicate return type",
				}

				suggestions := []rule.RuleSuggestion{
					{
						Message:  buildExplicitBooleanReturnTypeMessage(),
						FixesArr: []rule.RuleFix{},
					},
				}

				ctx.ReportNodeWithSuggestions(predicateNode, msg, suggestions...)
			}
		}

		return rule.RuleListeners{
			ast.KindIfStatement: func(node *ast.Node) {
				ifStmt := node.AsIfStatement()
				if ifStmt.Expression != nil {
					traverseNode(ifStmt.Expression, true)
				}
			},
			ast.KindWhileStatement: func(node *ast.Node) {
				whileStmt := node.AsWhileStatement()
				if whileStmt.Expression != nil {
					traverseNode(whileStmt.Expression, true)
				}
			},
			ast.KindDoStatement: func(node *ast.Node) {
				doStmt := node.AsDoStatement()
				if doStmt.Expression != nil {
					traverseNode(doStmt.Expression, true)
				}
			},
			ast.KindForStatement: func(node *ast.Node) {
				forStmt := node.AsForStatement()
				if forStmt.Condition != nil {
					traverseNode(forStmt.Condition, true)
				}
			},
			ast.KindConditionalExpression: func(node *ast.Node) {
				condExpr := node.AsConditionalExpression()
				if condExpr.Condition != nil {
					traverseNode(condExpr.Condition, true)
				}
			},
			ast.KindPrefixUnaryExpression: func(node *ast.Node) {
				prefix := node.AsPrefixUnaryExpression()
				if prefix.Operator == scanner.ExclamationToken {
					traverseNode(prefix.Operand, true)
				}
			},
			ast.KindCallExpression: func(node *ast.Node) {
				checkArrayPredicate(node)
			},
		}
	},
})
