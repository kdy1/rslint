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
	// Trim leading and trailing underscores (allowed)
	trimmed := strings.Trim(name, "_")

	// Empty after trimming is allowed
	if len(trimmed) == 0 {
		return true
	}

	// Check if it's all uppercase (CONSTANT_CASE is allowed)
	if trimmed == strings.ToUpper(trimmed) {
		return true
	}

	// Check for underscores in the middle (not camelCase/PascalCase)
	if strings.Contains(trimmed, "_") {
		return false
	}

	// camelCase or PascalCase (no underscores) is valid
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
	if node.Parent == nil {
		return false
	}

	parent := node.Parent
	parentKind := parent.Kind

	// Check if we're in an import declaration and should ignore
	if opts.IgnoreImports {
		if isInImportDeclaration(node) {
			return false
		}
	}

	// Check if we're in a destructuring pattern and should ignore
	if opts.IgnoreDestructuring {
		if isInDestructuring(node) {
			return false
		}
	}

	// Check if identifier is a property access and properties = "never"
	if opts.Properties == "never" {
		if isPropertyName(node) {
			return false
		}
	}

	// Only check specific contexts where camelcase should apply:
	// Variable declarations, function declarations, class declarations, etc.
	switch parentKind {
	case ast.KindVariableDeclaration,
		ast.KindFunctionDeclaration,
		ast.KindFunctionExpression,
		ast.KindArrowFunction,
		ast.KindClassDeclaration,
		ast.KindClassExpression,
		ast.KindMethodDeclaration,
		ast.KindParameter,
		ast.KindCatchClause,
		ast.KindPropertyDeclaration,
		ast.KindPropertySignature,
		ast.KindEnumMember:
		return true
	case ast.KindPropertyAssignment:
		// Check properties only if properties = "always"
		return opts.Properties == "always"
	}

	return false
}

func isInImportDeclaration(node *ast.Node) bool {
	current := node
	for current != nil {
		if current.Kind == ast.KindImportDeclaration || current.Kind == ast.KindImportEqualsDeclaration {
			return true
		}
		current = current.Parent
	}
	return false
}

func isInDestructuring(node *ast.Node) bool {
	current := node
	for current != nil {
		if current.Kind == ast.KindObjectBindingPattern || current.Kind == ast.KindArrayBindingPattern {
			return true
		}
		current = current.Parent
	}
	return false
}

func isPropertyName(node *ast.Node) bool {
	if node.Parent == nil {
		return false
	}
	parent := node.Parent
	// Check if this identifier is a property name (not property value)
	switch parent.Kind {
	case ast.KindPropertyAssignment:
		prop := parent.AsPropertyAssignment()
		return prop.Name() == node
	case ast.KindPropertyDeclaration:
		prop := parent.AsPropertyDeclaration()
		return prop.Name() == node
	case ast.KindMethodDeclaration:
		method := parent.AsMethodDeclaration()
		return method.Name() == node
	case ast.KindPropertySignature:
		// PropertySignature is for interface/type properties
		// For now, just check if it's a name by checking the parent kind
		return true
	}
	return false
}
