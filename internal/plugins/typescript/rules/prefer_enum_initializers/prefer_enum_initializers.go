package prefer_enum_initializers

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var PreferEnumInitializersRule = rule.CreateRule(rule.Rule{
	Name: "prefer-enum-initializers",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindEnumDeclaration: func(node *ast.Node) {
				enumDecl := node.AsEnumDeclaration()
				if enumDecl == nil || enumDecl.Members == nil {
					return
				}

				// Check each enum member for missing initializer
				for i, memberNode := range enumDecl.Members.Nodes {
					member := memberNode.AsEnumMember()
					if member == nil {
						continue
					}

					// If member has no initializer, report it
					if member.Initializer == nil {
						// Generate suggestions for the fix
						suggestions := generateSuggestions(member, i)

						ctx.ReportNodeWithSuggestions(member.Name(), rule.RuleMessage{
							Id:          "noInitializer",
							Description: fmt.Sprintf("Prefer initializing all enum members explicitly."),
						}, suggestions)
					}
				}
			},
		}
	},
})

func generateSuggestions(member *ast.EnumMember, index int) []rule.RuleSuggestion {
	if member.Name() == nil {
		return nil
	}

	name := member.Name()
	var memberName string

	// Extract the member name
	if identifier := name.AsIdentifier(); identifier != nil {
		memberName = identifier.EscapedText
	} else if stringLiteral := name.AsStringLiteral(); stringLiteral != nil {
		memberName = stringLiteral.Text
		// Remove quotes if present
		if len(memberName) >= 2 && memberName[0] == '"' && memberName[len(memberName)-1] == '"' {
			memberName = memberName[1 : len(memberName)-1]
		}
	}

	suggestions := []rule.RuleSuggestion{}

	// Suggestion 1: Initialize with index
	suggestions = append(suggestions, rule.RuleSuggestion{
		MessageId:   "addNumericInitializer",
		Description: fmt.Sprintf("Initialize with value %d", index),
		Fix: func(node *ast.Node) []rule.RuleFix {
			return []rule.RuleFix{
				{
					Range: rule.Range{
						Start: int(name.Pos()) + int(name.End()-name.Pos()),
						End:   int(name.Pos()) + int(name.End()-name.Pos()),
					},
					NewText: fmt.Sprintf(" = %d", index),
				},
			}
		},
	})

	// Suggestion 2: Initialize with index + 1
	suggestions = append(suggestions, rule.RuleSuggestion{
		MessageId:   "addNumericInitializer",
		Description: fmt.Sprintf("Initialize with value %d", index+1),
		Fix: func(node *ast.Node) []rule.RuleFix {
			return []rule.RuleFix{
				{
					Range: rule.Range{
						Start: int(name.Pos()) + int(name.End()-name.Pos()),
						End:   int(name.Pos()) + int(name.End()-name.Pos()),
					},
					NewText: fmt.Sprintf(" = %d", index+1),
				},
			}
		},
	})

	// Suggestion 3: Initialize with string value matching the name
	if memberName != "" {
		suggestions = append(suggestions, rule.RuleSuggestion{
			MessageId:   "addStringInitializer",
			Description: fmt.Sprintf("Initialize with value '%s'", memberName),
			Fix: func(node *ast.Node) []rule.RuleFix {
				return []rule.RuleFix{
					{
						Range: rule.Range{
							Start: int(name.Pos()) + int(name.End()-name.Pos()),
							End:   int(name.Pos()) + int(name.End()-name.Pos()),
						},
						NewText: fmt.Sprintf(" = '%s'", memberName),
					},
				}
			},
		})
	}

	return suggestions
}
