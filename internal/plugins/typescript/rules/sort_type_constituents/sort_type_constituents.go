package sort_type_constituents

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// SortTypeConstituentsOptions represents the configuration options
type SortTypeConstituentsOptions struct {
	CheckIntersections bool   `json:"checkIntersections"`
	CheckUnions        bool   `json:"checkUnions"`
	CaseSensitive      bool   `json:"caseSensitive"`
	GroupOrder         []string `json:"groupOrder"`
}

var SortTypeConstituentsRule = rule.CreateRule(rule.Rule{
	Name: "sort-type-constituents",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := SortTypeConstituentsOptions{
			CheckIntersections: true,
			CheckUnions:        true,
			CaseSensitive:      false,
			GroupOrder:         []string{},
		}

		// Parse options
		if options != nil {
			var optsMap map[string]interface{}
			if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
				optsMap, _ = optArray[0].(map[string]interface{})
			} else {
				optsMap, _ = options.(map[string]interface{})
			}

			if optsMap != nil {
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
					for _, item := range groupOrder {
						if str, ok := item.(string); ok {
							opts.GroupOrder = append(opts.GroupOrder, str)
						}
					}
				}
			}
		}

		// TODO: Implement type constituent sorting
		// This rule enforces sorted union and intersection types
		// 1. Check union types (A | B | C)
		// 2. Check intersection types (A & B & C)
		// 3. Sort constituents alphabetically or by group order
		// 4. Report unsorted types with fix suggestions

		return rule.RuleListeners{
			ast.KindUnionType: func(node *ast.Node) {
				// TODO: Check and enforce sorting of union types
			},
			ast.KindIntersectionType: func(node *ast.Node) {
				// TODO: Check and enforce sorting of intersection types
			},
		}
	},
})
