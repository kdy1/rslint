package prefer_nullish_coalescing

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildPreferNullishOverOrMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferNullishOverOr",
		Description: "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
	}
}

func buildPreferNullishOverTernaryMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferNullishOverTernary",
		Description: "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
	}
}

func buildNoStrictNullCheckMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noStrictNullCheck",
		Description: "This rule requires the `strictNullChecks` compiler option to be turned on to function correctly.",
	}
}

func buildSuggestNullishMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "suggestNullish",
		Description: "Fix to nullish coalescing operator (`??`).",
	}
}

type IgnorePrimitivesConfig struct {
	String  *bool
	Number  *bool
	Bigint  *bool
	Boolean *bool
}

type PreferNullishCoalescingOptions struct {
	IgnoreTernaryTests            *bool
	IgnoreConditionalTests        *bool
	IgnoreMixedLogicalExpressions *bool
	AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing *bool
	IgnorePrimitives              *IgnorePrimitivesConfig
	IgnoreBooleanCoercion         *bool
}

var PreferNullishCoalescingRule = rule.CreateRule(rule.Rule{
	Name: "prefer-nullish-coalescing",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts, ok := options.(PreferNullishCoalescingOptions)
		if !ok {
			opts = PreferNullishCoalescingOptions{}
		}

		// Set default values
		if opts.IgnoreTernaryTests == nil {
			opts.IgnoreTernaryTests = utils.Ref(false)
		}
		if opts.IgnoreConditionalTests == nil {
			opts.IgnoreConditionalTests = utils.Ref(true)
		}
		if opts.IgnoreMixedLogicalExpressions == nil {
			opts.IgnoreMixedLogicalExpressions = utils.Ref(false)
		}
		if opts.AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing == nil {
			opts.AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing = utils.Ref(false)
		}
		if opts.IgnorePrimitives == nil {
			opts.IgnorePrimitives = &IgnorePrimitivesConfig{}
		}
		if opts.IgnoreBooleanCoercion == nil {
			opts.IgnoreBooleanCoercion = utils.Ref(false)
		}

		// Check if strictNullChecks is enabled
		compilerOptions := ctx.Program.Options()
		strictNullChecksEnabled := utils.IsStrictCompilerOptionEnabled(
			compilerOptions,
			compilerOptions.StrictNullChecks,
		)

		if !strictNullChecksEnabled && !*opts.AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing {
			ctx.ReportNode(&ctx.SourceFile.Node, buildNoStrictNullCheckMessage())
			return rule.RuleListeners{}
		}

		// Helper function to check if a type is nullable (includes null or undefined)
		isNullableType := func(t *checker.Type) bool {
			for _, part := range utils.UnionTypeParts(t) {
				if utils.IsTypeFlagSet(part, checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) {
					return true
				}
			}
			return false
		}

		// Helper to check if we should ignore based on primitive types
		shouldIgnorePrimitive := func(t *checker.Type) bool {
			for _, part := range utils.UnionTypeParts(t) {
				// Skip nullable parts
				if utils.IsTypeFlagSet(part, checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) {
					continue
				}

				if opts.IgnorePrimitives.String != nil && *opts.IgnorePrimitives.String {
					if utils.IsTypeFlagSet(part, checker.TypeFlagsStringLike) {
						return true
					}
				}
				if opts.IgnorePrimitives.Number != nil && *opts.IgnorePrimitives.Number {
					if utils.IsTypeFlagSet(part, checker.TypeFlagsNumberLike) {
						return true
					}
				}
				if opts.IgnorePrimitives.Bigint != nil && *opts.IgnorePrimitives.Bigint {
					if utils.IsTypeFlagSet(part, checker.TypeFlagsBigIntLike) {
						return true
					}
				}
				if opts.IgnorePrimitives.Boolean != nil && *opts.IgnorePrimitives.Boolean {
					if utils.IsTypeFlagSet(part, checker.TypeFlagsBooleanLike) {
						return true
					}
				}
			}
			return false
		}

		// Check if node is in a conditional context (if, while, for, do-while, conditional of ternary)
		isInConditionalContext := func(node *ast.Node) bool {
			parent := node.Parent
			if parent == nil {
				return false
			}

			switch {
			case ast.IsIfStatement(parent):
				return parent.AsIfStatement().Expression == node
			case parent.Kind == ast.KindWhileStatement:
				return parent.AsWhileStatement().Expression == node
			case parent.Kind == ast.KindDoStatement:
				return parent.AsDoStatement().Expression == node
			case ast.IsForStatement(parent):
				return parent.AsForStatement().Condition == node
			case ast.IsConditionalExpression(parent):
				return parent.AsConditionalExpression().Condition == node
			}
			return false
		}

		// Check if expression contains mixed logical operators
		isMixedLogicalExpression := func(node *ast.Node) bool {
			if !ast.IsBinaryExpression(node) {
				return false
			}

			hasOr := false
			hasAnd := false

			var checkMixed func(*ast.Node)
			checkMixed = func(n *ast.Node) {
				if !ast.IsBinaryExpression(n) {
					return
				}
				be := n.AsBinaryExpression()
				if be.OperatorToken.Kind == ast.KindBarBarToken {
					hasOr = true
				} else if be.OperatorToken.Kind == ast.KindAmpersandAmpersandToken {
					hasAnd = true
				}
				checkMixed(be.Left)
				checkMixed(be.Right)
			}

			checkMixed(node)
			return hasOr && hasAnd
		}

		// Check if the node is inside a Boolean() call
		isInBooleanCall := func(node *ast.Node) bool {
			parent := node.Parent
			if parent == nil || !ast.IsCallExpression(parent) {
				return false
			}
			call := parent.AsCallExpression()
			if !ast.IsIdentifier(call.Expression) {
				return false
			}
			return call.Expression.AsIdentifier().Text == "Boolean"
		}

		checkBinaryExpression := func(node *ast.Node) {
			expr := node.AsBinaryExpression()

			// Only check || and ||= operators
			if expr.OperatorToken.Kind != ast.KindBarBarToken &&
			   expr.OperatorToken.Kind != ast.KindBarBarEqualsToken {
				return
			}

			// Check if in conditional context and should ignore
			if *opts.IgnoreConditionalTests && isInConditionalContext(node) {
				return
			}

			// Check if mixed logical expression and should ignore
			if *opts.IgnoreMixedLogicalExpressions && isMixedLogicalExpression(node) {
				return
			}

			// Check if in Boolean call and should ignore
			if *opts.IgnoreBooleanCoercion && isInBooleanCall(node) {
				return
			}

			// Get type of left operand
			leftType := ctx.TypeChecker.GetTypeAtLocation(expr.Left)

			// Check if left operand is nullable
			if !isNullableType(leftType) {
				return
			}

			// Check if we should ignore based on primitive type
			if shouldIgnorePrimitive(leftType) {
				return
			}

			// Don't report if type is any or unknown
			if utils.IsTypeFlagSet(leftType, checker.TypeFlagsAny|checker.TypeFlagsUnknown) {
				return
			}

			// Report the error
			ctx.ReportNode(node, buildPreferNullishOverOrMessage())
		}

		checkConditionalExpression := func(node *ast.Node) {
			if *opts.IgnoreTernaryTests {
				return
			}

			cond := node.AsConditionalExpression()

			// Check if condition is a binary expression checking for null/undefined
			if !ast.IsBinaryExpression(cond.Condition) {
				return
			}

			condExpr := cond.Condition.AsBinaryExpression()

			// Check for patterns like: x !== null && x !== undefined ? x : y
			// or: x === null || x === undefined ? y : x
			// These are more complex and we'll handle the simple cases first

			// Simple case: x != null ? x : y  or  x !== undefined ? x : y
			isNullCheck := false
			isUndefinedCheck := false
			var checkedNode *ast.Node
			var isTruthyCheck bool // true if checking for non-null, false if checking for null

			switch condExpr.OperatorToken.Kind {
			case ast.KindExclamationEqualsToken, ast.KindExclamationEqualsEqualsToken:
				// x != null or x !== undefined
				isTruthyCheck = true
				if condExpr.Right.Kind == ast.KindNullKeyword {
					isNullCheck = true
					checkedNode = condExpr.Left
				} else if ast.IsIdentifier(condExpr.Right) &&
					condExpr.Right.AsIdentifier().Text == "undefined" {
					isUndefinedCheck = true
					checkedNode = condExpr.Left
				}
			case ast.KindEqualsEqualsToken, ast.KindEqualsEqualsEqualsToken:
				// x == null or x === undefined
				isTruthyCheck = false
				if condExpr.Right.Kind == ast.KindNullKeyword {
					isNullCheck = true
					checkedNode = condExpr.Left
				} else if ast.IsIdentifier(condExpr.Right) &&
					condExpr.Right.AsIdentifier().Text == "undefined" {
					isUndefinedCheck = true
					checkedNode = condExpr.Left
				}
			}

			if !isNullCheck && !isUndefinedCheck {
				return
			}

			// Check if the consequent/alternate matches the checked node
			// Pattern: x !== null ? x : y  should use ??
			// Pattern: x === null ? y : x  should use ??
			var shouldReport bool
			if isTruthyCheck {
				// x !== null ? x : y  =>  x ?? y
				shouldReport = scanner.GetTextOfNode(checkedNode) == scanner.GetTextOfNode(cond.WhenTrue)
			} else {
				// x === null ? y : x  =>  x ?? y
				shouldReport = scanner.GetTextOfNode(checkedNode) == scanner.GetTextOfNode(cond.WhenFalse)
			}

			if shouldReport {
				// Verify the type is actually nullable
				if checkedNode != nil {
					checkedType := ctx.TypeChecker.GetTypeAtLocation(checkedNode)
					if isNullableType(checkedType) &&
					   !utils.IsTypeFlagSet(checkedType, checker.TypeFlagsAny|checker.TypeFlagsUnknown) &&
					   !shouldIgnorePrimitive(checkedType) {
						ctx.ReportNode(node, buildPreferNullishOverTernaryMessage())
					}
				}
			}
		}

		return rule.RuleListeners{
			ast.KindBinaryExpression:       checkBinaryExpression,
			ast.KindConditionalExpression:  checkConditionalExpression,
		}
	},
})
