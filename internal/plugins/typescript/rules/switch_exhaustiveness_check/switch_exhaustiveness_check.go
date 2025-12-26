package switch_exhaustiveness_check

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/ts"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type SwitchExhaustivenessCheckOptions struct {
	AllowDefaultCaseForExhaustiveSwitch bool   `json:"allowDefaultCaseForExhaustiveSwitch"`
	RequireDefaultForNonUnion           bool   `json:"requireDefaultForNonUnion"`
	DefaultCaseCommentPattern           string `json:"defaultCaseCommentPattern"`
	ConsiderDefaultExhaustiveForUnions  bool   `json:"considerDefaultExhaustiveForUnions"`
}

var defaultOptions = SwitchExhaustivenessCheckOptions{
	AllowDefaultCaseForExhaustiveSwitch: true,
	RequireDefaultForNonUnion:           false,
	DefaultCaseCommentPattern:           "^no default$",
	ConsiderDefaultExhaustiveForUnions:  false,
}

func buildSwitchIsNotExhaustiveMessage(missingBranches string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "switchIsNotExhaustive",
		Description: fmt.Sprintf("Switch is not exhaustive. Cases not matched: %s", missingBranches),
	}
}

func buildAddMissingCasesMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "addMissingCases",
		Description: "Add branches for missing cases.",
	}
}

func buildDangerousDefaultCaseMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "dangerousDefaultCase",
		Description: "The switch statement is exhaustive, so the default case is unnecessary.",
	}
}

type switchMetadata struct {
	containsNonLiteralType    bool
	defaultCase               *ast.Node
	missingLiteralBranchTypes []ts.Type
	symbolName                string
}

func isTypeLiteralLikeType(typeObj ts.Type) bool {
	flags := typeObj.GetFlags()
	literalFlags := ts.TypeFlagsLiteral | ts.TypeFlagsUndefined | ts.TypeFlagsNull | ts.TypeFlagsUniqueESSymbol
	return (flags & literalFlags) != 0
}

func doesTypeContainNonLiteralType(typeObj ts.Type) bool {
	unionTypes := ts.GetUnionConstituents(typeObj)
	for _, unionPart := range unionTypes {
		intersectionTypes := ts.GetIntersectionConstituents(unionPart)
		allNonLiteral := true
		for _, intersectionPart := range intersectionTypes {
			if isTypeLiteralLikeType(intersectionPart) {
				allNonLiteral = false
				break
			}
		}
		if allNonLiteral {
			return true
		}
	}
	return false
}

func typeToString(checker ts.TypeChecker, typeObj ts.Type) string {
	flags := ts.TypeFormatFlagsAllowUniqueESSymbolType |
		ts.TypeFormatFlagsUseAliasDefinedOutsideCurrentScope |
		ts.TypeFormatFlagsUseFullyQualifiedType
	return checker.TypeToString(typeObj, nil, flags)
}

func getSwitchMetadata(
	ctx rule.RuleContext,
	node *ast.Node,
	commentPattern *regexp.Regexp,
) switchMetadata {
	switchStmt := node.AsSwitchStatement()
	caseBlock := switchStmt.CaseBlock

	var defaultCase *ast.Node
	cases := caseBlock.Clauses

	// Find default case
	for _, caseClause := range cases {
		if ast.IsDefaultClause(caseClause) {
			defaultCase = caseClause
			break
		}
	}

	// Get discriminant type
	discriminantType := ctx.TypeChecker.GetTypeAtLocation(switchStmt.Expression)
	discriminantType = ctx.TypeChecker.GetBaseConstraintOfType(discriminantType)
	if discriminantType == nil {
		discriminantType = ctx.TypeChecker.GetTypeAtLocation(switchStmt.Expression)
	}

	symbolName := ""
	symbol := discriminantType.GetSymbol()
	if symbol != nil {
		symbolName = symbol.GetEscapedName()
	}

	containsNonLiteralType := doesTypeContainNonLiteralType(discriminantType)

	// Collect case types
	caseTypes := make(map[string]ts.Type)
	for _, caseClause := range cases {
		if ast.IsCaseClause(caseClause) {
			caseNode := caseClause.AsCaseClause()
			if caseNode.Expression != nil {
				caseType := ctx.TypeChecker.GetTypeAtLocation(caseNode.Expression)
				caseType = ctx.TypeChecker.GetBaseConstraintOfType(caseType)
				if caseType == nil {
					caseType = ctx.TypeChecker.GetTypeAtLocation(caseNode.Expression)
				}
				typeStr := typeToString(ctx.TypeChecker, caseType)
				caseTypes[typeStr] = caseType
			}
		}
	}

	// Find missing literal branch types
	var missingLiteralBranchTypes []ts.Type
	unionParts := ts.GetUnionConstituents(discriminantType)

	for _, unionPart := range unionParts {
		intersectionParts := ts.GetIntersectionConstituents(unionPart)
		for _, intersectionPart := range intersectionParts {
			typeStr := typeToString(ctx.TypeChecker, intersectionPart)

			// Skip if we already have this case
			if _, exists := caseTypes[typeStr]; exists {
				continue
			}

			// Skip if not a literal-like type
			if !isTypeLiteralLikeType(intersectionPart) {
				continue
			}

			// Check for undefined types
			isUndefinedType := (intersectionPart.GetFlags() & ts.TypeFlagsUndefined) != 0
			if isUndefinedType {
				hasUndefinedCase := false
				for _, caseType := range caseTypes {
					if (caseType.GetFlags() & ts.TypeFlagsUndefined) != 0 {
						hasUndefinedCase = true
						break
					}
				}
				if hasUndefinedCase {
					continue
				}
			}

			missingLiteralBranchTypes = append(missingLiteralBranchTypes, intersectionPart)
		}
	}

	return switchMetadata{
		containsNonLiteralType:    containsNonLiteralType,
		defaultCase:               defaultCase,
		missingLiteralBranchTypes: missingLiteralBranchTypes,
		symbolName:                symbolName,
	}
}

func generateMissingCaseFix(
	ctx rule.RuleContext,
	node *ast.Node,
	missingBranchTypes []ts.Type,
	defaultCase *ast.Node,
	symbolName string,
	includeDefault bool,
) string {
	switchStmt := node.AsSwitchStatement()
	caseBlock := switchStmt.CaseBlock
	cases := caseBlock.Clauses

	var caseIndent string
	if len(cases) > 0 {
		lastCase := cases[len(cases)-1]
		lastCaseRange := utils.TrimNodeTextRange(ctx.SourceFile, lastCase)
		line := ctx.SourceFile.GetLineAndCharacterOfPosition(lastCaseRange.Pos())
		caseIndent = strings.Repeat(" ", line.Character)
	} else {
		// Use switch statement indentation
		switchRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
		line := ctx.SourceFile.GetLineAndCharacterOfPosition(switchRange.Pos())
		caseIndent = strings.Repeat(" ", line.Character)
	}

	var missingCases []string

	if includeDefault {
		missingCases = append(missingCases, "default: { throw new Error('default case') }")
	}

	for _, missingBranchType := range missingBranchTypes {
		symbol := missingBranchType.GetSymbol()
		var caseTest string

		flags := missingBranchType.GetFlags()
		isESSymbol := (flags & ts.TypeFlagsESSymbolLike) != 0

		if isESSymbol && symbol != nil {
			caseTest = symbol.GetEscapedName()
		} else {
			caseTest = typeToString(ctx.TypeChecker, missingBranchType)
		}

		// Escape single quotes and backslashes in the error message
		escapedCaseTest := strings.ReplaceAll(caseTest, "\\", "\\\\")
		escapedCaseTest = strings.ReplaceAll(escapedCaseTest, "'", "\\'")

		missingCases = append(missingCases,
			fmt.Sprintf("case %s: { throw new Error('Not implemented yet: %s case') }", caseTest, escapedCaseTest))
	}

	return strings.Join(missingCases, "\n"+caseIndent)
}

var SwitchExhaustivenessCheckRule = rule.CreateRule(rule.Rule{
	Name: "switch-exhaustiveness-check",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := defaultOptions
		if options != nil {
			if optMap, ok := options.(map[string]interface{}); ok {
				if val, ok := optMap["allowDefaultCaseForExhaustiveSwitch"].(bool); ok {
					opts.AllowDefaultCaseForExhaustiveSwitch = val
				}
				if val, ok := optMap["requireDefaultForNonUnion"].(bool); ok {
					opts.RequireDefaultForNonUnion = val
				}
				if val, ok := optMap["defaultCaseCommentPattern"].(string); ok {
					opts.DefaultCaseCommentPattern = val
				}
				if val, ok := optMap["considerDefaultExhaustiveForUnions"].(bool); ok {
					opts.ConsiderDefaultExhaustiveForUnions = val
				}
			}
		}

		commentPattern := regexp.MustCompile(opts.DefaultCaseCommentPattern)

		return rule.RuleListeners{
			ast.KindSwitchStatement: func(node *ast.Node) {
				if node.Kind != ast.KindSwitchStatement {
					return
				}

				metadata := getSwitchMetadata(ctx, node, commentPattern)

				// Check 1: Switch exhaustiveness
				checkSwitchExhaustive(ctx, node, metadata, opts)

				// Check 2: Unnecessary default case
				checkSwitchUnnecessaryDefaultCase(ctx, metadata, opts)

				// Check 3: Missing default for non-union
				checkSwitchNoUnionDefaultCase(ctx, node, metadata, opts)
			},
		}
	},
})

func checkSwitchExhaustive(
	ctx rule.RuleContext,
	node *ast.Node,
	metadata switchMetadata,
	opts SwitchExhaustivenessCheckOptions,
) {
	// If considerDefaultExhaustiveForUnions is enabled, the presence of a default case
	// always makes the switch exhaustive.
	if opts.ConsiderDefaultExhaustiveForUnions && metadata.defaultCase != nil {
		return
	}

	if len(metadata.missingLiteralBranchTypes) > 0 {
		var missingBranchStrs []string
		for _, missingType := range metadata.missingLiteralBranchTypes {
			flags := missingType.GetFlags()
			isESSymbol := (flags & ts.TypeFlagsESSymbolLike) != 0

			if isESSymbol {
				symbol := missingType.GetSymbol()
				if symbol != nil {
					missingBranchStrs = append(missingBranchStrs,
						fmt.Sprintf("typeof %s", symbol.GetEscapedName()))
				}
			} else {
				missingBranchStrs = append(missingBranchStrs,
					typeToString(ctx.TypeChecker, missingType))
			}
		}

		missingBranches := strings.Join(missingBranchStrs, " | ")

		switchStmt := node.AsSwitchStatement()
		discriminant := switchStmt.Expression

		// Generate fix suggestion
		fixStr := generateMissingCaseFix(ctx, node, metadata.missingLiteralBranchTypes,
			metadata.defaultCase, metadata.symbolName, false)

		// Find insertion point
		caseBlock := switchStmt.CaseBlock
		cases := caseBlock.Clauses
		var insertPos int

		if len(cases) > 0 {
			if metadata.defaultCase != nil {
				// Insert before default case
				insertPos = utils.TrimNodeTextRange(ctx.SourceFile, metadata.defaultCase).Pos()
			} else {
				// Insert after last case
				lastCase := cases[len(cases)-1]
				insertPos = utils.TrimNodeTextRange(ctx.SourceFile, lastCase).End()
			}
		} else {
			// No cases exist, insert inside the case block
			insertPos = utils.TrimNodeTextRange(ctx.SourceFile, caseBlock).Pos() + 1
		}

		fix := rule.RuleFixInsertAfter(core.NewTextRange(insertPos, insertPos), "\n"+fixStr)

		ctx.ReportNodeWithSuggestions(
			discriminant,
			buildSwitchIsNotExhaustiveMessage(missingBranches),
			rule.RuleSuggestion{
				Message:  buildAddMissingCasesMessage(),
				FixesArr: []rule.RuleFix{fix},
			},
		)
	}
}

func checkSwitchUnnecessaryDefaultCase(
	ctx rule.RuleContext,
	metadata switchMetadata,
	opts SwitchExhaustivenessCheckOptions,
) {
	if opts.AllowDefaultCaseForExhaustiveSwitch {
		return
	}

	if len(metadata.missingLiteralBranchTypes) == 0 &&
		metadata.defaultCase != nil &&
		!metadata.containsNonLiteralType {
		ctx.ReportNode(
			metadata.defaultCase,
			buildDangerousDefaultCaseMessage(),
		)
	}
}

func checkSwitchNoUnionDefaultCase(
	ctx rule.RuleContext,
	node *ast.Node,
	metadata switchMetadata,
	opts SwitchExhaustivenessCheckOptions,
) {
	if !opts.RequireDefaultForNonUnion {
		return
	}

	if metadata.containsNonLiteralType && metadata.defaultCase == nil {
		switchStmt := node.AsSwitchStatement()
		discriminant := switchStmt.Expression

		// Generate fix for adding default case
		fixStr := generateMissingCaseFix(ctx, node, nil, metadata.defaultCase, metadata.symbolName, true)

		caseBlock := switchStmt.CaseBlock
		cases := caseBlock.Clauses
		var insertPos int

		if len(cases) > 0 {
			lastCase := cases[len(cases)-1]
			insertPos = utils.TrimNodeTextRange(ctx.SourceFile, lastCase).End()
		} else {
			insertPos = utils.TrimNodeTextRange(ctx.SourceFile, caseBlock).Pos() + 1
		}

		fix := rule.RuleFixInsertAfter(core.NewTextRange(insertPos, insertPos), "\n"+fixStr)

		ctx.ReportNodeWithSuggestions(
			discriminant,
			buildSwitchIsNotExhaustiveMessage("default"),
			rule.RuleSuggestion{
				Message:  buildAddMissingCasesMessage(),
				FixesArr: []rule.RuleFix{fix},
			},
		)
	}
}
