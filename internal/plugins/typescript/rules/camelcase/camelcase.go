package camelcase

import (
	"regexp"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// CamelcaseOptions represents the configuration options for the camelcase rule
type CamelcaseOptions struct {
	Properties               string   `json:"properties"`               // "always" or "never"
	IgnoreDestructuring      bool     `json:"ignoreDestructuring"`
	IgnoreImports            bool     `json:"ignoreImports"`
	IgnoreGlobals            bool     `json:"ignoreGlobals"`
	Allow                    []string `json:"allow"`                    // Array of regex patterns to allow
}

var CamelcaseRule = rule.CreateRule(rule.Rule{
	Name: "camelcase",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := CamelcaseOptions{
			Properties:          "always",
			IgnoreDestructuring: false,
			IgnoreImports:       false,
			IgnoreGlobals:       false,
			Allow:               []string{},
		}

		// Parse options with dual-format support
		if options != nil {
			var optsMap map[string]interface{}
			var ok bool

			if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
				optsMap, ok = optArray[0].(map[string]interface{})
			} else {
				optsMap, ok = options.(map[string]interface{})
			}

			if ok {
				if properties, ok := optsMap["properties"].(string); ok {
					opts.Properties = properties
				}
				if ignoreDestructuring, ok := optsMap["ignoreDestructuring"].(bool); ok {
					opts.IgnoreDestructuring = ignoreDestructuring
				}
				if ignoreImports, ok := optsMap["ignoreImports"].(bool); ok {
					opts.IgnoreImports = ignoreImports
				}
				if ignoreGlobals, ok := optsMap["ignoreGlobals"].(bool); ok {
					opts.IgnoreGlobals = ignoreGlobals
				}
				if allowVal, ok := optsMap["allow"].([]interface{}); ok {
					for _, pattern := range allowVal {
						if str, ok := pattern.(string); ok {
							opts.Allow = append(opts.Allow, str)
						}
					}
				}
			}
		}

		return rule.RuleListeners{
			ast.KindIdentifier: func(node *ast.Node) {
				identifier := node.AsIdentifier()
				if identifier == nil {
					return
				}

				// Get the identifier name
				nameRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
				name := ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()]

				// Check if name is allowed by regex patterns
				if isAllowedByPattern(name, opts.Allow) {
					return
				}

				// TODO: Implement full validation logic:
				// 1. Check if identifier is in destructuring and should be ignored
				// 2. Check if identifier is an import and should be ignored
				// 3. Check if identifier is a global and should be ignored
				// 4. Check if identifier is a property and properties="never"
				// 5. Validate camelCase format
				// 6. Report violations

				// Basic camelCase check (incomplete implementation)
				if !isCamelCase(name) && !isAllowedByPattern(name, opts.Allow) {
					// Only report for identifiers that we should check
					// This is a simplified implementation
					if shouldCheckIdentifier(node, opts) {
						ctx.ReportNode(node, rule.RuleMessage{
							Id:          "notCamelCase",
							Description: "Identifier '" + name + "' is not in camelCase.",
						})
					}
				}
			},
		}
	},
})

func isCamelCase(name string) bool {
	// Check for leading underscores (sometimes allowed)
	name = strings.TrimLeft(name, "_")

	// Empty after trimming
	if len(name) == 0 {
		return true
	}

	// Must start with lowercase letter
	if name[0] < 'a' || name[0] > 'z' {
		return false
	}

	// Check for underscores in the middle (not camelCase)
	if strings.Contains(name, "_") {
		return false
	}

	return true
}

func isAllowedByPattern(name string, patterns []string) bool {
	for _, pattern := range patterns {
		matched, err := regexp.MatchString(pattern, name)
		if err == nil && matched {
			return true
		}
	}
	return false
}

func shouldCheckIdentifier(node *ast.Node, opts CamelcaseOptions) bool {
	// TODO: Implement full checking logic based on context
	// This is a placeholder that needs full implementation
	return true
}
