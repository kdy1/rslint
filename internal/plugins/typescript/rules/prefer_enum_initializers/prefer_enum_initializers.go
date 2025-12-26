package prefer_enum_initializers

import (
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
				for _, memberNode := range enumDecl.Members.Nodes {
					member := memberNode.AsEnumMember()
					if member == nil {
						continue
					}

					// If the member has no initializer, report it
					if member.Initializer == nil {
						ctx.ReportNode(member.Name(), rule.RuleMessage{
							Id:          "defineInitializer",
							Description: "Enum member values should be explicitly defined.",
						})
					}
				}
			},
		}
	},
})
