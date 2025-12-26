package sort_type_constituents

import (
	"fmt"
	"sort"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type SortTypeConstituentsOptions struct {
	CheckIntersections bool     `json:"checkIntersections"`
	CheckUnions        bool     `json:"checkUnions"`
	CaseSensitive      bool     `json:"caseSensitive"`
	GroupOrder         []string `json:"groupOrder"`
}

var defaultGroupOrder = []string{
	"named",
	"keyword",
	"operator",
	"literal",
	"function",
	"import",
	"conditional",
	"object",
	"tuple",
	"intersection",
	"union",
	"nullish",
}

type typeGroup string

const (
	groupNamed       typeGroup = "named"
	groupKeyword     typeGroup = "keyword"
	groupOperator    typeGroup = "operator"
	groupLiteral     typeGroup = "literal"
	groupFunction    typeGroup = "function"
	groupImport      typeGroup = "import"
	groupConditional typeGroup = "conditional"
	groupObject      typeGroup = "object"
	groupTuple       typeGroup = "tuple"
	groupIntersection typeGroup = "intersection"
	groupUnion       typeGroup = "union"
	groupNullish     typeGroup = "nullish"
)

type typeConstituent struct {
	node       *ast.Node
	text       string
	group      typeGroup
	groupIndex int
}

func getTypeGroup(node *ast.Node) typeGroup {
	// Handle parenthesized types - unwrap to check inner type
	if ast.IsParenthesizedTypeNode(node) {
		paren := node.AsParenthesizedTypeNode()
		if paren != nil && paren.Type != nil {
			innerGroup := getTypeGroup(paren.Type)
			// If the inner type is a union or intersection with leading operator,
			// treat it specially
			if innerGroup == groupUnion || innerGroup == groupIntersection {
				return innerGroup
			}
		}
	}

	switch node.Kind {
	case ast.KindNullKeyword, ast.KindUndefinedKeyword:
		return groupNullish
	case ast.KindAnyKeyword, ast.KindBooleanKeyword, ast.KindBigIntKeyword,
		ast.KindNeverKeyword, ast.KindNumberKeyword, ast.KindObjectKeyword,
		ast.KindStringKeyword, ast.KindSymbolKeyword, ast.KindUnknownKeyword,
		ast.KindVoidKeyword, ast.KindIntrinsicKeyword:
		return groupKeyword
	case ast.KindThisType:
		return groupKeyword
	case ast.KindLiteralType:
		return groupLiteral
	case ast.KindTemplateLiteralType:
		return groupLiteral
	case ast.KindFunctionType, ast.KindConstructorType:
		return groupFunction
	case ast.KindImportType:
		return groupImport
	case ast.KindConditionalType:
		return groupConditional
	case ast.KindTypeLiteral, ast.KindMappedType:
		return groupObject
	case ast.KindTupleType:
		return groupTuple
	case ast.KindIntersectionType:
		return groupIntersection
	case ast.KindUnionType:
		return groupUnion
	case ast.KindTypeOperator:
		return groupOperator
	case ast.KindArrayType:
		return groupNamed
	case ast.KindTypeReference, ast.KindIdentifier, ast.KindQualifiedName:
		return groupNamed
	case ast.KindIndexedAccessType:
		return groupOperator
	default:
		return groupNamed
	}
}

func buildNotSortedMessage(typeName, typeType string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "notSortedNamed",
		Description: fmt.Sprintf("%s type '%s' is not sorted.", typeType, typeName),
		Data: map[string]interface{}{
			"name": typeName,
			"type": typeType,
		},
	}
}

func buildNotSortedUnnamedMessage(typeType string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "notSorted",
		Description: fmt.Sprintf("%s type is not sorted.", typeType),
		Data: map[string]interface{}{
			"type": typeType,
		},
	}
}

func buildSuggestFixMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "suggestFix",
		Description: "Sort type constituents.",
	}
}

func removeParentheses(text string) string {
	text = strings.TrimSpace(text)
	for strings.HasPrefix(text, "(") && strings.HasSuffix(text, ")") {
		text = text[1 : len(text)-1]
		text = strings.TrimSpace(text)
	}
	return text
}

func naturalCompare(a, b string, caseSensitive bool) int {
	if !caseSensitive {
		a = strings.ToLower(a)
		b = strings.ToLower(b)
	}
	
	// Natural sort: try to compare numbers naturally
	aLen := len(a)
	bLen := len(b)
	i, j := 0, 0
	
	for i < aLen && j < bLen {
		aChar := a[i]
		bChar := b[j]
		
		// Check if both are digits
		if aChar >= '0' && aChar <= '9' && bChar >= '0' && bChar <= '9' {
			// Extract numbers
			aNum := 0
			for i < aLen && a[i] >= '0' && a[i] <= '9' {
				aNum = aNum*10 + int(a[i]-'0')
				i++
			}
			bNum := 0
			for j < bLen && b[j] >= '0' && b[j] <= '9' {
				bNum = bNum*10 + int(b[j]-'0')
				j++
			}
			if aNum != bNum {
				return aNum - bNum
			}
		} else {
			if aChar != bChar {
				return int(aChar) - int(bChar)
			}
			i++
			j++
		}
	}
	
	return aLen - bLen
}

var SortTypeConstituentsRule = rule.CreateRule(rule.Rule{
	Name: "sort-type-constituents",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := SortTypeConstituentsOptions{
			CheckIntersections: true,
			CheckUnions:        true,
			CaseSensitive:      false,
			GroupOrder:         defaultGroupOrder,
		}

		// Parse options
		if options != nil {
			var optsMap map[string]interface{}
			var ok bool

			if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
				optsMap, ok = optArray[0].(map[string]interface{})
			} else {
				optsMap, ok = options.(map[string]interface{})
			}

			if ok {
				if checkIntersections, ok := optsMap["checkIntersections"].(bool); ok {
					opts.CheckIntersections = checkIntersections
				}
				if checkUnions, ok := optsMap["checkUnions"].(bool); ok {
					opts.CheckUnions = checkUnions
				}
				if caseSensitive, ok := optsMap["caseSensitive"].(bool); ok {
					opts.CaseSensitive = caseSensitive
				}
				if groupOrder, ok := optsMap["groupOrder"].([]interface{}); ok {
					opts.GroupOrder = make([]string, 0, len(groupOrder))
					for _, g := range groupOrder {
						if gs, ok := g.(string); ok {
							opts.GroupOrder = append(opts.GroupOrder, gs)
						}
					}
				}
			}
		}

		// Create group order map
		groupOrderMap := make(map[string]int)
		for i, g := range opts.GroupOrder {
			groupOrderMap[g] = i
		}

		checkType := func(node *ast.Node, isUnion bool) {
			var types []*ast.Node
			var typeNode *ast.Node

			if isUnion {
				if !opts.CheckUnions {
					return
				}
				unionType := node.AsUnionTypeNode()
				if unionType == nil {
					return
				}
				types = unionType.Types.Nodes
				typeNode = node
			} else {
				if !opts.CheckIntersections {
					return
				}
				intersectionType := node.AsIntersectionTypeNode()
				if intersectionType == nil {
					return
				}
				types = intersectionType.Types.Nodes
				typeNode = node
			}

			if len(types) <= 1 {
				return
			}

			// Get constituents with their text and group
			constituents := make([]typeConstituent, 0, len(types))
			hasComments := false

			for _, t := range types {
				text := getNodeText(ctx.SourceFile, t)
				group := getTypeGroup(t)
				
				// Check for comments between constituents
				if hasCommentsBetween(ctx.SourceFile, t) {
					hasComments = true
				}

				groupIdx, ok := groupOrderMap[string(group)]
				if !ok {
					// Groups not in the order list go to the end
					groupIdx = len(opts.GroupOrder)
				}

				constituents = append(constituents, typeConstituent{
					node:       t,
					text:       text,
					group:      group,
					groupIndex: groupIdx,
				})
			}

			// Check if already sorted
			sorted := make([]typeConstituent, len(constituents))
			copy(sorted, constituents)
			sort.SliceStable(sorted, func(i, j int) bool {
				// First sort by group
				if sorted[i].groupIndex != sorted[j].groupIndex {
					return sorted[i].groupIndex < sorted[j].groupIndex
				}
				// Then by text (natural sort)
				cmp := naturalCompare(
					removeParentheses(sorted[i].text),
					removeParentheses(sorted[j].text),
					opts.CaseSensitive,
				)
				return cmp < 0
			})

			// Check if order changed
			isSorted := true
			for i := range constituents {
				if constituents[i].node != sorted[i].node {
					isSorted = false
					break
				}
			}

			if isSorted {
				return
			}

			// Find the parent type alias to get the name
			typeName := ""
			parent := typeNode.Parent
			for parent != nil {
				if parent.Kind == ast.KindTypeAliasDeclaration {
					typeAlias := parent.AsTypeAliasDeclaration()
					if typeAlias != nil && typeAlias.Name != nil {
						if id := typeAlias.Name.AsIdentifier(); id != nil {
							typeName = id.Text
						}
					}
					break
				}
				parent = parent.Parent
			}

			typeStr := "Union"
			if !isUnion {
				typeStr = "Intersection"
			}

			// Build the fixed text
			operator := " | "
			if !isUnion {
				operator = " & "
			}

			sortedTexts := make([]string, len(sorted))
			for i, c := range sorted {
				sortedTexts[i] = removeParentheses(c.text)
			}
			fixedText := strings.Join(sortedTexts, operator)

			// If there are comments between constituents, provide suggestion instead of auto-fix
			if hasComments {
				if typeName != "" {
					ctx.ReportNodeWithSuggestions(
						typeNode,
						buildNotSortedMessage(typeName, typeStr),
						rule.RuleSuggestion{
							Message:  buildSuggestFixMessage(),
							FixesArr: []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, typeNode, fixedText)},
						},
					)
				} else {
					ctx.ReportNodeWithSuggestions(
						typeNode,
						buildNotSortedUnnamedMessage(typeStr),
						rule.RuleSuggestion{
							Message:  buildSuggestFixMessage(),
							FixesArr: []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, typeNode, fixedText)},
						},
					)
				}
			} else {
				// Auto-fix
				if typeName != "" {
					ctx.ReportNodeWithFixes(
						typeNode,
						buildNotSortedMessage(typeName, typeStr),
						rule.RuleFixReplace(ctx.SourceFile, typeNode, fixedText),
					)
				} else {
					ctx.ReportNodeWithFixes(
						typeNode,
						buildNotSortedUnnamedMessage(typeStr),
						rule.RuleFixReplace(ctx.SourceFile, typeNode, fixedText),
					)
				}
			}
		}

		return rule.RuleListeners{
			ast.KindUnionType: func(node *ast.Node) {
				checkType(node, true)
			},
			ast.KindIntersectionType: func(node *ast.Node) {
				checkType(node, false)
			},
		}
	},
})

func getNodeText(sourceFile *core.SourceFile, node *ast.Node) string {
	nodeRange := utils.TrimNodeTextRange(sourceFile, node)
	return sourceFile.Text()[nodeRange.Pos():nodeRange.End()]
}

func hasCommentsBetween(sourceFile *core.SourceFile, node *ast.Node) bool {
	// Simple heuristic: check if the node text contains comments
	// This is a simplified version - a full implementation would need to check
	// the actual comment ranges
	text := getNodeText(sourceFile, node)
	return strings.Contains(text, "/*") || strings.Contains(text, "//")
}
