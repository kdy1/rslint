package consistent_indexed_object_style

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// ConsistentIndexedObjectStyleOptions defines the configuration
type ConsistentIndexedObjectStyleOptions struct {
	Style string `json:"style"` // "record" or "index-signature"
}

func parseOptions(options interface{}) ConsistentIndexedObjectStyleOptions {
	opts := ConsistentIndexedObjectStyleOptions{
		Style: "record", // Default
	}

	if options == nil {
		return opts
	}

	switch v := options.(type) {
	case string:
		if v == "index-signature" || v == "record" {
			opts.Style = v
		}
	case map[string]interface{}:
		if style, ok := v["style"].(string); ok {
			if style == "index-signature" || style == "record" {
				opts.Style = style
			}
		}
	}

	return opts
}

func buildPreferRecordMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferRecord",
		Description: "Prefer using the 'Record' utility type instead of an index signature.",
	}
}

func buildPreferIndexSignatureMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferIndexSignature",
		Description: "Prefer using an index signature instead of the 'Record' utility type.",
	}
}

// Check if a type is a Record utility type
func isRecordType(typeNode *ast.Node) bool {
	if typeNode == nil || typeNode.Kind != ast.KindTypeReference {
		return false
	}

	typeRef := typeNode.AsTypeReference()
	if typeRef == nil || typeRef.TypeName == nil {
		return false
	}

	// Check if the type name is "Record"
	if typeRef.TypeName.Kind == ast.KindIdentifier {
		identifier := typeRef.TypeName.AsIdentifier()
		if identifier != nil && identifier.EscapedText == "Record" {
			return true
		}
	}

	return false
}

// Check if an interface/type has an index signature
func hasIndexSignature(members []ast.TypeElement) bool {
	for _, member := range members {
		if member.Kind == ast.KindIndexSignature {
			return true
		}
	}
	return false
}

// Convert index signature to Record type
func convertIndexSignatureToRecord(ctx rule.RuleContext, indexSig *ast.IndexSignatureDeclaration) string {
	if indexSig == nil || len(indexSig.Parameters.Arr) == 0 {
		return "Record<string, unknown>"
	}

	param := indexSig.Parameters.Arr[0]
	if param == nil {
		return "Record<string, unknown>"
	}

	// Get key type
	keyType := "string"
	if param.Type != nil {
		keyRange := utils.TrimNodeTextRange(ctx.SourceFile, param.Type)
		keyType = ctx.SourceFile.Text()[keyRange.Pos():keyRange.End()]
	}

	// Get value type
	valueType := "unknown"
	if indexSig.Type != nil {
		valueRange := utils.TrimNodeTextRange(ctx.SourceFile, indexSig.Type)
		valueType = ctx.SourceFile.Text()[valueRange.Pos():valueRange.End()]
	}

	return "Record<" + keyType + ", " + valueType + ">"
}

// Convert Record type to index signature
func convertRecordToIndexSignature(ctx rule.RuleContext, typeRef *ast.TypeReference) string {
	if typeRef == nil || typeRef.TypeArguments == nil || len(typeRef.TypeArguments.Arr) < 2 {
		return "{ [key: string]: unknown }"
	}

	keyType := typeRef.TypeArguments.Arr[0]
	valueType := typeRef.TypeArguments.Arr[1]

	keyRange := utils.TrimNodeTextRange(ctx.SourceFile, keyType)
	keyText := ctx.SourceFile.Text()[keyRange.Pos():keyRange.End()]

	valueRange := utils.TrimNodeTextRange(ctx.SourceFile, valueType)
	valueText := ctx.SourceFile.Text()[valueRange.Pos():valueRange.End()]

	return "{ [key: " + keyText + "]: " + valueText + " }"
}

var ConsistentIndexedObjectStyleRule = rule.CreateRule(rule.Rule{
	Name: "consistent-indexed-object-style",
	Run: func(ctx rule.RuleContext, options interface{}) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			// Check TypeAliasDeclaration for index signatures
			ast.KindTypeAliasDeclaration: func(node *ast.Node) {
				if node.Kind != ast.KindTypeAliasDeclaration {
					return
				}

				typeAlias := node.AsTypeAliasDeclaration()
				if typeAlias == nil || typeAlias.Type == nil {
					return
				}

				// Check if it's a Record type (prefer index-signature style)
				if opts.Style == "index-signature" && isRecordType(typeAlias.Type) {
					typeRef := typeAlias.Type.AsTypeReference()
					recordText := convertRecordToIndexSignature(ctx, typeRef)

					ctx.ReportNodeWithFixes(
						typeAlias.Type,
						buildPreferIndexSignatureMessage(),
						rule.RuleFixReplace(ctx.SourceFile, typeAlias.Type, recordText),
					)
					return
				}

				// Check if it's an index signature in a type literal
				if opts.Style == "record" && typeAlias.Type.Kind == ast.KindTypeLiteral {
					typeLiteral := typeAlias.Type.AsTypeLiteralNode()
					if typeLiteral == nil {
						return
					}

					members := typeLiteral.Members.Arr
					if len(members) == 1 && members[0].Kind == ast.KindIndexSignature {
						indexSig := members[0].AsIndexSignatureDeclaration()
						recordText := convertIndexSignatureToRecord(ctx, indexSig)

						ctx.ReportNodeWithFixes(
							typeAlias.Type,
							buildPreferRecordMessage(),
							rule.RuleFixReplace(ctx.SourceFile, typeAlias.Type, recordText),
						)
					}
				}
			},

			// Check InterfaceDeclaration for index signatures
			ast.KindInterfaceDeclaration: func(node *ast.Node) {
				if node.Kind != ast.KindInterfaceDeclaration {
					return
				}

				interfaceDecl := node.AsInterfaceDeclaration()
				if interfaceDecl == nil {
					return
				}

				members := interfaceDecl.Members.Arr

				// Only report if interface has a single index signature and nothing else
				if opts.Style == "record" && len(members) == 1 && members[0].Kind == ast.KindIndexSignature {
					indexSig := members[0].AsIndexSignatureDeclaration()
					recordText := convertIndexSignatureToRecord(ctx, indexSig)

					// Get interface name
					nameRange := utils.TrimNodeTextRange(ctx.SourceFile, interfaceDecl.Name)
					interfaceName := ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()]

					// Get type parameters if any
					typeParams := ""
					if interfaceDecl.TypeParameters != nil && len(interfaceDecl.TypeParameters.Arr) > 0 {
						typeParamsRange := utils.TrimNodeTextRange(ctx.SourceFile, interfaceDecl.TypeParameters)
						typeParams = ctx.SourceFile.Text()[typeParamsRange.Pos():typeParamsRange.End()]
					}

					replacement := "type " + interfaceName + typeParams + " = " + recordText

					ctx.ReportNodeWithFixes(
						node,
						buildPreferRecordMessage(),
						rule.RuleFixReplace(ctx.SourceFile, node, replacement),
					)
				}
			},
		}
	},
})
