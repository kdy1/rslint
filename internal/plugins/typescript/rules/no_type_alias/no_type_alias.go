package no_type_alias

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type AllowOption string

const (
	AllowAlways                      AllowOption = "always"
	AllowNever                       AllowOption = "never"
	AllowInUnions                    AllowOption = "in-unions"
	AllowInIntersections             AllowOption = "in-intersections"
	AllowInUnionsAndIntersections    AllowOption = "in-unions-and-intersections"
)

type NoTypeAliasOptions struct {
	AllowAliases          AllowOption `json:"allowAliases"`
	AllowCallbacks        AllowOption `json:"allowCallbacks"`
	AllowConditionalTypes AllowOption `json:"allowConditionalTypes"`
	AllowConstructors     AllowOption `json:"allowConstructors"`
	AllowLiterals         AllowOption `json:"allowLiterals"`
	AllowMappedTypes      AllowOption `json:"allowMappedTypes"`
	AllowTupleTypes       AllowOption `json:"allowTupleTypes"`
	AllowGenerics         AllowOption `json:"allowGenerics"`
}

func parseOptions(options any) NoTypeAliasOptions {
	opts := NoTypeAliasOptions{
		AllowAliases:          AllowNever,
		AllowCallbacks:        AllowNever,
		AllowConditionalTypes: AllowNever,
		AllowConstructors:     AllowNever,
		AllowLiterals:         AllowNever,
		AllowMappedTypes:      AllowNever,
		AllowTupleTypes:       AllowNever,
		AllowGenerics:         AllowNever,
	}

	if options == nil {
		return opts
	}

	var optsMap map[string]interface{}
	var ok bool

	// Handle array format: [{ option: value }]
	if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
		optsMap, ok = optArray[0].(map[string]interface{})
	} else {
		// Handle direct object format: { option: value }
		optsMap, ok = options.(map[string]interface{})
	}

	if !ok {
		return opts
	}

	if val, ok := optsMap["allowAliases"].(string); ok {
		opts.AllowAliases = AllowOption(val)
	}
	if val, ok := optsMap["allowCallbacks"].(string); ok {
		opts.AllowCallbacks = AllowOption(val)
	}
	if val, ok := optsMap["allowConditionalTypes"].(string); ok {
		opts.AllowConditionalTypes = AllowOption(val)
	}
	if val, ok := optsMap["allowConstructors"].(string); ok {
		opts.AllowConstructors = AllowOption(val)
	}
	if val, ok := optsMap["allowLiterals"].(string); ok {
		opts.AllowLiterals = AllowOption(val)
	}
	if val, ok := optsMap["allowMappedTypes"].(string); ok {
		opts.AllowMappedTypes = AllowOption(val)
	}
	if val, ok := optsMap["allowTupleTypes"].(string); ok {
		opts.AllowTupleTypes = AllowOption(val)
	}
	if val, ok := optsMap["allowGenerics"].(string); ok {
		opts.AllowGenerics = AllowOption(val)
	}

	return opts
}

type TypeCategory int

const (
	TypeCategoryAlias TypeCategory = iota
	TypeCategoryCallback
	TypeCategoryConditional
	TypeCategoryConstructor
	TypeCategoryLiteral
	TypeCategoryMapped
	TypeCategoryTuple
	TypeCategoryGeneric
)

func (tc TypeCategory) String() string {
	switch tc {
	case TypeCategoryAlias:
		return "aliases"
	case TypeCategoryCallback:
		return "callbacks"
	case TypeCategoryConditional:
		return "conditional types"
	case TypeCategoryConstructor:
		return "constructors"
	case TypeCategoryLiteral:
		return "literals"
	case TypeCategoryMapped:
		return "mapped types"
	case TypeCategoryTuple:
		return "tuples"
	case TypeCategoryGeneric:
		return "generics"
	default:
		return "unknown"
	}
}

func (tc TypeCategory) TitleCase() string {
	switch tc {
	case TypeCategoryAlias:
		return "Aliases"
	case TypeCategoryCallback:
		return "Callbacks"
	case TypeCategoryConditional:
		return "Conditional types"
	case TypeCategoryConstructor:
		return "Constructors"
	case TypeCategoryLiteral:
		return "Literals"
	case TypeCategoryMapped:
		return "Mapped types"
	case TypeCategoryTuple:
		return "Tuples"
	case TypeCategoryGeneric:
		return "Generics"
	default:
		return "Unknown"
	}
}

type compositionType int

const (
	compositionTypeNone compositionType = iota
	compositionTypeUnion
	compositionTypeIntersection
)

func (ct compositionType) String() string {
	switch ct {
	case compositionTypeUnion:
		return "union"
	case compositionTypeIntersection:
		return "intersection"
	default:
		return ""
	}
}

// isTypeOfQuery checks if node is a TypeQuery (typeof expression)
func isTypeOfQuery(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return node.Kind == ast.KindTypeQuery
}

// isMappedType checks if the type contains a mapped type
func isMappedType(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return node.Kind == ast.KindMappedType
}

// isTypeLiteral checks if the node is a TypeLiteral (object type literal)
func isTypeLiteral(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return node.Kind == ast.KindTypeLiteral
}

// isTupleType checks if the node is a TupleType
func isTupleType(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return node.Kind == ast.KindTupleType
}

// isConditionalType checks if the node is a ConditionalType
func isConditionalType(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return node.Kind == ast.KindConditionalType
}

// isConstructorType checks if the node is a ConstructorType
func isConstructorType(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return node.Kind == ast.KindConstructorType
}

// isFunctionType checks if the node is a FunctionType
func isFunctionType(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return node.Kind == ast.KindFunctionType
}

// isTypeReference checks if the node is a TypeReference
func isTypeReference(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return node.Kind == ast.KindTypeReference
}

// hasTypeParameters checks if a TypeReference has type parameters (is generic)
func hasTypeParameters(node *ast.Node) bool {
	if node == nil {
		return false
	}
	typeRef := node.AsTypeReference()
	if typeRef == nil {
		return false
	}
	return typeRef.TypeArguments != nil && len(typeRef.TypeArguments.Nodes) > 0
}

// isPrimitiveType checks if the node is a primitive type (string, number, boolean, etc.)
func isPrimitiveType(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindStringKeyword, ast.KindNumberKeyword, ast.KindBooleanKeyword,
		ast.KindAnyKeyword, ast.KindUnknownKeyword, ast.KindNeverKeyword,
		ast.KindVoidKeyword, ast.KindUndefinedKeyword, ast.KindNullKeyword,
		ast.KindSymbolKeyword, ast.KindBigIntKeyword, ast.KindObjectKeyword:
		return true
	case ast.KindArrayType:
		return true
	case ast.KindTypeReference:
		// Built-in generic types like Array, ReadonlyArray are considered primitive
		typeRef := node.AsTypeReference()
		if typeRef == nil || typeRef.TypeName == nil {
			return false
		}
		// Check if it's a simple identifier reference (not qualified)
		return typeRef.TypeName.Kind == ast.KindIdentifier
	default:
		return false
	}
}

// isLiteralType checks if the node is a literal type (string/number/boolean literal, template literal)
func isLiteralType(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindLiteralType, ast.KindTemplateLiteralType:
		return true
	case ast.KindTrueKeyword, ast.KindFalseKeyword:
		return true
	default:
		return false
	}
}

// categorizeType determines what category a type belongs to
func categorizeType(node *ast.Node) TypeCategory {
	if node == nil {
		return TypeCategoryAlias
	}

	// Check for conditional types first
	if isConditionalType(node) {
		return TypeCategoryConditional
	}

	// Check for constructor types
	if isConstructorType(node) {
		return TypeCategoryConstructor
	}

	// Check for function/callback types
	if isFunctionType(node) {
		return TypeCategoryCallback
	}

	// Check for mapped types
	if isMappedType(node) {
		return TypeCategoryMapped
	}

	// Check for tuple types
	if isTupleType(node) {
		return TypeCategoryTuple
	}

	// Check for type literals (object types)
	if isTypeLiteral(node) {
		return TypeCategoryLiteral
	}

	// Check for generics (type references with type arguments)
	if isTypeReference(node) && hasTypeParameters(node) {
		return TypeCategoryGeneric
	}

	// Everything else is an alias (primitives, literals, typeof, etc.)
	return TypeCategoryAlias
}

// getAllowOption gets the allow option for a specific category
func getAllowOption(opts NoTypeAliasOptions, category TypeCategory) AllowOption {
	switch category {
	case TypeCategoryAlias:
		return opts.AllowAliases
	case TypeCategoryCallback:
		return opts.AllowCallbacks
	case TypeCategoryConditional:
		return opts.AllowConditionalTypes
	case TypeCategoryConstructor:
		return opts.AllowConstructors
	case TypeCategoryLiteral:
		return opts.AllowLiterals
	case TypeCategoryMapped:
		return opts.AllowMappedTypes
	case TypeCategoryTuple:
		return opts.AllowTupleTypes
	case TypeCategoryGeneric:
		return opts.AllowGenerics
	default:
		return AllowNever
	}
}

// isAllowed checks if a type is allowed based on options and composition context
func isAllowed(opts NoTypeAliasOptions, category TypeCategory, composition compositionType) bool {
	allowOpt := getAllowOption(opts, category)

	switch allowOpt {
	case AllowAlways:
		return true
	case AllowNever:
		return false
	case AllowInUnions:
		return composition == compositionTypeUnion
	case AllowInIntersections:
		return composition == compositionTypeIntersection
	case AllowInUnionsAndIntersections:
		return composition == compositionTypeUnion || composition == compositionTypeIntersection
	default:
		return false
	}
}

// getCompositionType determines if we're in a union or intersection context
func getCompositionType(parent *ast.Node) compositionType {
	if parent == nil {
		return compositionTypeNone
	}
	if parent.Kind == ast.KindUnionType {
		return compositionTypeUnion
	}
	if parent.Kind == ast.KindIntersectionType {
		return compositionTypeIntersection
	}
	return compositionTypeNone
}

// checkTypeNode checks a type node and reports if it's not allowed
func checkTypeNode(ctx rule.RuleContext, opts NoTypeAliasOptions, node *ast.Node, parent *ast.Node) {
	if node == nil {
		return
	}

	category := categorizeType(node)
	composition := getCompositionType(parent)

	if !isAllowed(opts, category, composition) {
		if composition != compositionTypeNone {
			// Report composition error
			description := fmt.Sprintf("Type %s of %s types is not allowed.", composition.String(), category.TitleCase())
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "noCompositionAlias",
				Description: description,
			})
		} else {
			// Report simple type alias error
			description := fmt.Sprintf("Type %s are not allowed.", category.String())
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "noTypeAlias",
				Description: description,
			})
		}
	}
}

// walkUnionOrIntersection walks through union or intersection type members
func walkUnionOrIntersection(ctx rule.RuleContext, opts NoTypeAliasOptions, node *ast.Node) {
	if node == nil {
		return
	}

	var types []*ast.Node
	if node.Kind == ast.KindUnionType {
		unionType := node.AsUnionTypeNode()
		if unionType != nil && unionType.Types != nil {
			types = unionType.Types.Nodes
		}
	} else if node.Kind == ast.KindIntersectionType {
		intersectionType := node.AsIntersectionTypeNode()
		if intersectionType != nil && intersectionType.Types != nil {
			types = intersectionType.Types.Nodes
		}
	}

	for _, typeNode := range types {
		checkTypeNode(ctx, opts, typeNode, node)
	}
}

var NoTypeAliasRule = rule.CreateRule(rule.Rule{
	Name: "no-type-alias",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			ast.KindTypeAliasDeclaration: func(node *ast.Node) {
				typeAlias := node.AsTypeAliasDeclaration()
				if typeAlias == nil || typeAlias.Type == nil {
					return
				}

				typeNode := typeAlias.Type

				// Check if it's a union or intersection type
				if typeNode.Kind == ast.KindUnionType || typeNode.Kind == ast.KindIntersectionType {
					walkUnionOrIntersection(ctx, opts, typeNode)
				} else {
					// Simple type alias
					checkTypeNode(ctx, opts, typeNode, nil)
				}
			},
		}
	},
})
