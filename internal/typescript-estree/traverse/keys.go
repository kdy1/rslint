// Package traverse provides AST traversal utilities for TypeScript ESTree nodes.
package traverse

import (
	"reflect"

	"github.com/web-infra-dev/rslint/internal/typescript-estree/types"
)

// VisitorKeys maps node types to the names of their child node fields.
// The order of keys matters - they should be ordered by their appearance in source code.
// Field names must match the exported Go struct field names (capitalized).
var VisitorKeys = map[string][]string{
	// Base nodes
	"Program":    {"Body"},
	"Identifier": {},

	// Literals
	"Literal":        {},
	"SimpleLiteral":  {},
	"RegExpLiteral":  {},
	"BigIntLiteral":  {},
	"TemplateLiteral": {"Quasis", "Expressions"},

	// Expressions
	"ThisExpression":       {},
	"ArrayExpression":      {"Elements"},
	"ObjectExpression":     {"Properties"},
	"Property":             {"Key", "Value"},
	"SpreadElement":        {"Argument"},
	"FunctionExpression":   {"ID", "Params", "Body"},
	"ArrowFunctionExpression": {"Params", "Body"},
	"UnaryExpression":      {"Argument"},
	"UpdateExpression":     {"Argument"},
	"BinaryExpression":     {"Left", "Right"},
	"AssignmentExpression": {"Left", "Right"},
	"LogicalExpression":    {"Left", "Right"},
	"MemberExpression":     {"Object", "Property"},
	"ConditionalExpression": {"Test", "Consequent", "Alternate"},
	"CallExpression":       {"Callee", "Arguments"},
	"NewExpression":        {"Callee", "Arguments"},
	"SequenceExpression":   {"Expressions"},
	"TaggedTemplateExpression": {"Tag", "Quasi"},
	"ClassExpression":      {"ID", "SuperClass", "Body"},
	"MetaProperty":         {"Meta", "Property"},
	"AwaitExpression":      {"Argument"},
	"ImportExpression":     {"Source"},
	"ChainExpression":      {"Expression"},
	"YieldExpression":      {"Argument"},

	// Patterns
	"ObjectPattern":      {"Properties"},
	"ArrayPattern":       {"Elements"},
	"RestElement":        {"Argument"},
	"AssignmentPattern":  {"Left", "Right"},
	"MemberPattern":      {"Object", "Property"},

	// Statements
	"ExpressionStatement": {"Expression"},
	"BlockStatement":      {"Body"},
	"EmptyStatement":      {},
	"DebuggerStatement":   {},
	"WithStatement":       {"Object", "Body"},
	"ReturnStatement":     {"Argument"},
	"LabeledStatement":    {"Label", "Body"},
	"BreakStatement":      {"Label"},
	"ContinueStatement":   {"Label"},
	"IfStatement":         {"Test", "Consequent", "Alternate"},
	"SwitchStatement":     {"Discriminant", "Cases"},
	"SwitchCase":          {"Test", "Consequent"},
	"ThrowStatement":      {"Argument"},
	"TryStatement":        {"Block", "Handler", "Finalizer"},
	"CatchClause":         {"Param", "Body"},
	"WhileStatement":      {"Test", "Body"},
	"DoWhileStatement":    {"Body", "Test"},
	"ForStatement":        {"Init", "Test", "Update", "Body"},
	"ForInStatement":      {"Left", "Right", "Body"},
	"ForOfStatement":      {"Left", "Right", "Body"},

	// Declarations
	"FunctionDeclaration": {"ID", "Params", "Body"},
	"VariableDeclaration": {"Declarations"},
	"VariableDeclarator":  {"ID", "Init"},
	"ClassDeclaration":    {"ID", "SuperClass", "Body"},
	"ClassBody":           {"Body"},

	// Class members
	"MethodDefinition":       {"Key", "Value"},
	"PropertyDefinition":     {"Key", "Value"},
	"StaticBlock":            {"Body"},
	"PrivateIdentifier":      {},
	"AccessorProperty":       {"Key", "Value"},

	// Modules
	"ImportDeclaration":       {"Specifiers", "Source"},
	"ImportSpecifier":         {"Imported", "Local"},
	"ImportDefaultSpecifier":  {"Local"},
	"ImportNamespaceSpecifier": {"Local"},
	"ExportNamedDeclaration":  {"Declaration", "Specifiers", "Source"},
	"ExportSpecifier":         {"Exported", "Local"},
	"ExportDefaultDeclaration": {"Declaration"},
	"ExportAllDeclaration":    {"Exported", "Source"},

	// TypeScript-specific nodes
	"TSTypeAnnotation":        {"TypeAnnotation"},
	"TSTypeParameterDeclaration": {"Params"},
	"TSTypeParameter":         {"Name", "Constraint", "Default"},
	"TSTypeParameterInstantiation": {"Params"},

	// TypeScript type nodes
	"TSAnyKeyword":        {},
	"TSBooleanKeyword":    {},
	"TSBigIntKeyword":     {},
	"TSNeverKeyword":      {},
	"TSNullKeyword":       {},
	"TSNumberKeyword":     {},
	"TSObjectKeyword":     {},
	"TSStringKeyword":     {},
	"TSSymbolKeyword":     {},
	"TSUndefinedKeyword":  {},
	"TSUnknownKeyword":    {},
	"TSVoidKeyword":       {},
	"TSThisType":          {},
	"TSLiteralType":       {"Literal"},
	"TSArrayType":         {"ElementType"},
	"TSTupleType":         {"ElementTypes"},
	"TSUnionType":         {"Types"},
	"TSIntersectionType":  {"Types"},
	"TSConditionalType":   {"CheckType", "ExtendsType", "TrueType", "FalseType"},
	"TSInferType":         {"TypeParameter"},
	"TSParenthesizedType": {"TypeAnnotation"},
	"TSTypeReference":     {"TypeName", "TypeArguments"},
	"TSQualifiedName":     {"Left", "Right"},
	"TSIndexedAccessType": {"ObjectType", "IndexType"},
	"TSMappedType":        {"TypeParameter", "TypeAnnotation"},
	"TSTypeLiteral":       {"Members"},
	"TSFunctionType":      {"TypeParameters", "Parameters", "ReturnType"},
	"TSConstructorType":   {"TypeParameters", "Parameters", "ReturnType"},
	"TSTypeQuery":         {"ExprName"},
	"TSTypePredicate":     {"ParameterName", "TypeAnnotation"},
	"TSTypeOperator":      {"TypeAnnotation"},
	"TSRestType":          {"TypeAnnotation"},
	"TSOptionalType":      {"TypeAnnotation"},
	"TSNamedTupleMember":  {"ElementType", "Label"},
	"TSImportType":        {"Argument", "Qualifier", "TypeArguments"},
	"TSTemplateLiteralType": {"Quasis", "Types"},

	// TypeScript declarations
	"TSInterfaceDeclaration": {"ID", "TypeParameters", "Extends", "Body"},
	"TSInterfaceBody":        {"Body"},
	"TSTypeAliasDeclaration": {"ID", "TypeParameters", "TypeAnnotation"},
	"TSEnumDeclaration":      {"ID", "Members"},
	"TSEnumMember":           {"ID", "Initializer"},
	"TSModuleDeclaration":    {"ID", "Body"},
	"TSModuleBlock":          {"Body"},
	"TSNamespaceExportDeclaration": {"ID"},

	// TypeScript expressions
	"TSAsExpression":          {"Expression", "TypeAnnotation"},
	"TSSatisfiesExpression":   {"Expression", "TypeAnnotation"},
	"TSTypeAssertion":         {"TypeAnnotation", "Expression"},
	"TSNonNullExpression":     {"Expression"},
	"TSInstantiationExpression": {"Expression", "TypeArguments"},

	// TypeScript signatures
	"TSCallSignatureDeclaration":     {"TypeParameters", "Parameters", "ReturnType"},
	"TSConstructSignatureDeclaration": {"TypeParameters", "Parameters", "ReturnType"},
	"TSMethodSignature":              {"Key", "TypeParameters", "Parameters", "ReturnType"},
	"TSPropertySignature":            {"Key", "TypeAnnotation"},
	"TSIndexSignature":               {"Parameters", "TypeAnnotation"},

	// TypeScript class members
	"TSAbstractMethodDefinition":   {"Key", "Value"},
	"TSAbstractPropertyDefinition": {"Key", "TypeAnnotation"},

	// JSX (for completeness)
	"JSXElement":            {"OpeningElement", "Children", "ClosingElement"},
	"JSXOpeningElement":     {"Name", "Attributes"},
	"JSXClosingElement":     {"Name"},
	"JSXFragment":           {"OpeningFragment", "Children", "ClosingFragment"},
	"JSXAttribute":          {"Name", "Value"},
	"JSXSpreadAttribute":    {"Argument"},
	"JSXIdentifier":         {},
	"JSXNamespacedName":     {"Namespace", "Name"},
	"JSXMemberExpression":   {"Object", "Property"},
	"JSXExpressionContainer": {"Expression"},
	"JSXEmptyExpression":    {},
	"JSXText":               {},
	"JSXSpreadChild":        {"Expression"},
}

// GetChildNodes returns all child nodes of a given node using reflection.
// This is a fallback for nodes without explicit visitor keys.
func GetChildNodes(node types.Node) []types.Node {
	if node == nil {
		return nil
	}

	var children []types.Node
	nodeType := node.Type()

	// Try to use visitor keys first
	if keys, ok := VisitorKeys[nodeType]; ok {
		v := reflect.ValueOf(node)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		for _, key := range keys {
			field := v.FieldByName(key)
			if field.IsValid() {
				children = append(children, extractNodes(field)...)
			}
		}
		return children
	}

	// Fallback: use reflection to find all node fields
	v := reflect.ValueOf(node)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		children = append(children, extractNodes(field)...)
	}

	return children
}

// extractNodes extracts Node instances from a reflect.Value.
// Handles single nodes, slices of nodes, and interfaces containing nodes.
func extractNodes(v reflect.Value) []types.Node {
	var nodes []types.Node

	if !v.IsValid() || v.IsZero() {
		return nodes
	}

	switch v.Kind() {
	case reflect.Interface, reflect.Ptr:
		if !v.IsNil() {
			elem := v.Elem()
			if elem.IsValid() {
				// Check if it implements Node interface
				if node, ok := v.Interface().(types.Node); ok {
					nodes = append(nodes, node)
				} else {
					// Recurse into the value
					nodes = append(nodes, extractNodes(elem)...)
				}
			}
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			// Handle both pointer and non-pointer elements
			if elem.Kind() == reflect.Ptr || elem.Kind() == reflect.Interface {
				// For pointers and interfaces, check if they implement Node
				if !elem.IsNil() {
					if node, ok := elem.Interface().(types.Node); ok {
						nodes = append(nodes, node)
					} else {
						// Recurse
						nodes = append(nodes, extractNodes(elem)...)
					}
				}
			} else if elem.CanAddr() {
				// For non-pointer structs, get their address
				addr := elem.Addr()
				if node, ok := addr.Interface().(types.Node); ok {
					nodes = append(nodes, node)
				}
			} else {
				// If we can't get address, try the value directly
				nodes = append(nodes, extractNodes(elem)...)
			}
		}
	case reflect.Struct:
		// Handle struct values by trying to get their address if possible
		if v.CanAddr() {
			addr := v.Addr()
			if node, ok := addr.Interface().(types.Node); ok {
				nodes = append(nodes, node)
			}
		}
	default:
		// Try to convert to Node
		if node, ok := v.Interface().(types.Node); ok {
			nodes = append(nodes, node)
		}
	}

	return nodes
}
