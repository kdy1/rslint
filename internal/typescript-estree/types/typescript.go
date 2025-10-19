// Package types provides ESTree-compliant type definitions for TypeScript AST nodes.
// This file contains TypeScript-specific ESTree extensions (TSESTree types).
package types

// TSTypeAnnotation represents a TypeScript type annotation.
type TSTypeAnnotation struct {
	BaseNode
	TypeAnnotation TSType `json:"typeAnnotation"`
}

// TSType is a marker interface for all TypeScript type nodes.
type TSType interface {
	Node
	tsTypeNode()
}

// TSAnyKeyword represents the 'any' type keyword.
type TSAnyKeyword struct {
	BaseNode
}

func (t *TSAnyKeyword) tsTypeNode() {}

// TSUnknownKeyword represents the 'unknown' type keyword.
type TSUnknownKeyword struct {
	BaseNode
}

func (t *TSUnknownKeyword) tsTypeNode() {}

// TSNumberKeyword represents the 'number' type keyword.
type TSNumberKeyword struct {
	BaseNode
}

func (t *TSNumberKeyword) tsTypeNode() {}

// TSBooleanKeyword represents the 'boolean' type keyword.
type TSBooleanKeyword struct {
	BaseNode
}

func (t *TSBooleanKeyword) tsTypeNode() {}

// TSStringKeyword represents the 'string' type keyword.
type TSStringKeyword struct {
	BaseNode
}

func (t *TSStringKeyword) tsTypeNode() {}

// TSSymbolKeyword represents the 'symbol' type keyword.
type TSSymbolKeyword struct {
	BaseNode
}

func (t *TSSymbolKeyword) tsTypeNode() {}

// TSVoidKeyword represents the 'void' type keyword.
type TSVoidKeyword struct {
	BaseNode
}

func (t *TSVoidKeyword) tsTypeNode() {}

// TSNullKeyword represents the 'null' type keyword.
type TSNullKeyword struct {
	BaseNode
}

func (t *TSNullKeyword) tsTypeNode() {}

// TSUndefinedKeyword represents the 'undefined' type keyword.
type TSUndefinedKeyword struct {
	BaseNode
}

func (t *TSUndefinedKeyword) tsTypeNode() {}

// TSNeverKeyword represents the 'never' type keyword.
type TSNeverKeyword struct {
	BaseNode
}

func (t *TSNeverKeyword) tsTypeNode() {}

// TSBigIntKeyword represents the 'bigint' type keyword.
type TSBigIntKeyword struct {
	BaseNode
}

func (t *TSBigIntKeyword) tsTypeNode() {}

// TSObjectKeyword represents the 'object' type keyword.
type TSObjectKeyword struct {
	BaseNode
}

func (t *TSObjectKeyword) tsTypeNode() {}

// TSThisType represents the 'this' type.
type TSThisType struct {
	BaseNode
}

func (t *TSThisType) tsTypeNode() {}

// TSArrayType represents an array type (T[]).
type TSArrayType struct {
	BaseNode
	ElementType TSType `json:"elementType"`
}

func (t *TSArrayType) tsTypeNode() {}

// TSTupleType represents a tuple type ([T1, T2, ...]).
type TSTupleType struct {
	BaseNode
	ElementTypes []TSType `json:"elementTypes"`
}

func (t *TSTupleType) tsTypeNode() {}

// TSUnionType represents a union type (T1 | T2 | ...).
type TSUnionType struct {
	BaseNode
	Types []TSType `json:"types"`
}

func (t *TSUnionType) tsTypeNode() {}

// TSIntersectionType represents an intersection type (T1 & T2 & ...).
type TSIntersectionType struct {
	BaseNode
	Types []TSType `json:"types"`
}

func (t *TSIntersectionType) tsTypeNode() {}

// TSConditionalType represents a conditional type (T extends U ? X : Y).
type TSConditionalType struct {
	BaseNode
	CheckType   TSType `json:"checkType"`
	ExtendsType TSType `json:"extendsType"`
	TrueType    TSType `json:"trueType"`
	FalseType   TSType `json:"falseType"`
}

func (t *TSConditionalType) tsTypeNode() {}

// TSInferType represents an infer type (infer T).
type TSInferType struct {
	BaseNode
	TypeParameter *TSTypeParameter `json:"typeParameter"`
}

func (t *TSInferType) tsTypeNode() {}

// TSParenthesizedType represents a parenthesized type ((T)).
type TSParenthesizedType struct {
	BaseNode
	TypeAnnotation TSType `json:"typeAnnotation"`
}

func (t *TSParenthesizedType) tsTypeNode() {}

// TSTypeReference represents a type reference.
type TSTypeReference struct {
	BaseNode
	TypeName       Node                          `json:"typeName"` // Identifier or TSQualifiedName
	TypeParameters *TSTypeParameterInstantiation `json:"typeParameters"`
}

func (t *TSTypeReference) tsTypeNode() {}

// TSQualifiedName represents a qualified name (A.B.C).
type TSQualifiedName struct {
	BaseNode
	Left  Node        `json:"left"` // Identifier or TSQualifiedName
	Right *Identifier `json:"right"`
}

// TSTypeParameterInstantiation represents type parameter instantiation (<T, U>).
type TSTypeParameterInstantiation struct {
	BaseNode
	Params []TSType `json:"params"`
}

// TSTypeParameterDeclaration represents type parameter declaration (<T extends U>).
type TSTypeParameterDeclaration struct {
	BaseNode
	Params []TSTypeParameter `json:"params"`
}

// TSTypeParameter represents a type parameter.
type TSTypeParameter struct {
	BaseNode
	Name       *Identifier `json:"name"`
	Constraint TSType      `json:"constraint"`
	Default    TSType      `json:"default"`
}

// TSFunctionType represents a function type ((args) => returnType).
type TSFunctionType struct {
	BaseNode
	Params         []Node                      `json:"params"`
	ReturnType     *TSTypeAnnotation           `json:"returnType"`
	TypeParameters *TSTypeParameterDeclaration `json:"typeParameters"`
}

func (t *TSFunctionType) tsTypeNode() {}

// TSConstructorType represents a constructor type (new (args) => T).
type TSConstructorType struct {
	BaseNode
	Params         []Node                      `json:"params"`
	ReturnType     *TSTypeAnnotation           `json:"returnType"`
	TypeParameters *TSTypeParameterDeclaration `json:"typeParameters"`
}

func (t *TSConstructorType) tsTypeNode() {}

// TSTypeLiteral represents a type literal ({prop: type}).
type TSTypeLiteral struct {
	BaseNode
	Members []Node `json:"members"` // TSPropertySignature, TSMethodSignature, etc.
}

func (t *TSTypeLiteral) tsTypeNode() {}

// TSPropertySignature represents a property signature in a type literal.
type TSPropertySignature struct {
	BaseNode
	Key            Expression        `json:"key"`
	TypeAnnotation *TSTypeAnnotation `json:"typeAnnotation"`
	Optional       bool              `json:"optional"`
	Readonly       bool              `json:"readonly"`
	Computed       bool              `json:"computed"`
}

// TSMethodSignature represents a method signature in a type literal.
type TSMethodSignature struct {
	BaseNode
	Key            Expression                  `json:"key"`
	Params         []Node                      `json:"params"`
	ReturnType     *TSTypeAnnotation           `json:"returnType"`
	TypeParameters *TSTypeParameterDeclaration `json:"typeParameters"`
	Optional       bool                        `json:"optional"`
	Computed       bool                        `json:"computed"`
}

// TSIndexSignature represents an index signature ([key: type]: valueType).
type TSIndexSignature struct {
	BaseNode
	Parameters     []Node            `json:"parameters"`
	TypeAnnotation *TSTypeAnnotation `json:"typeAnnotation"`
}

// TSCallSignatureDeclaration represents a call signature.
type TSCallSignatureDeclaration struct {
	BaseNode
	Params         []Node                      `json:"params"`
	ReturnType     *TSTypeAnnotation           `json:"returnType"`
	TypeParameters *TSTypeParameterDeclaration `json:"typeParameters"`
}

// TSConstructSignatureDeclaration represents a construct signature.
type TSConstructSignatureDeclaration struct {
	BaseNode
	Params         []Node                      `json:"params"`
	ReturnType     *TSTypeAnnotation           `json:"returnType"`
	TypeParameters *TSTypeParameterDeclaration `json:"typeParameters"`
}

// TSLiteralType represents a literal type (e.g., 123, "hello", true).
type TSLiteralType struct {
	BaseNode
	Literal Node `json:"literal"` // SimpleLiteral or UnaryExpression (for negative numbers)
}

func (t *TSLiteralType) tsTypeNode() {}

// TSTypeQuery represents a typeof type query (typeof x).
type TSTypeQuery struct {
	BaseNode
	ExprName Node `json:"exprName"` // Identifier or TSQualifiedName
}

func (t *TSTypeQuery) tsTypeNode() {}

// TSMappedType represents a mapped type ({[K in keyof T]: U}).
type TSMappedType struct {
	BaseNode
	TypeParameter  *TSTypeParameter `json:"typeParameter"`
	TypeAnnotation TSType           `json:"typeAnnotation"`
	Optional       bool             `json:"optional"`
	Readonly       bool             `json:"readonly"`
}

func (t *TSMappedType) tsTypeNode() {}

// TSIndexedAccessType represents an indexed access type (T[K]).
type TSIndexedAccessType struct {
	BaseNode
	ObjectType TSType `json:"objectType"`
	IndexType  TSType `json:"indexType"`
}

func (t *TSIndexedAccessType) tsTypeNode() {}

// TSRestType represents a rest type (...T).
type TSRestType struct {
	BaseNode
	TypeAnnotation TSType `json:"typeAnnotation"`
}

func (t *TSRestType) tsTypeNode() {}

// TSOptionalType represents an optional type (T?).
type TSOptionalType struct {
	BaseNode
	TypeAnnotation TSType `json:"typeAnnotation"`
}

func (t *TSOptionalType) tsTypeNode() {}

// TSInterfaceDeclaration represents an interface declaration.
type TSInterfaceDeclaration struct {
	BaseNode
	ID             *Identifier                 `json:"id"`
	Body           *TSInterfaceBody            `json:"body"`
	Extends        []TSInterfaceHeritage       `json:"extends"`
	TypeParameters *TSTypeParameterDeclaration `json:"typeParameters"`
	Declare        bool                        `json:"declare"`
}

func (t *TSInterfaceDeclaration) statementNode()   {}
func (t *TSInterfaceDeclaration) declarationNode() {}

// TSInterfaceBody represents the body of an interface.
type TSInterfaceBody struct {
	BaseNode
	Body []Node `json:"body"` // TSPropertySignature, TSMethodSignature, etc.
}

// TSInterfaceHeritage represents an interface heritage clause.
type TSInterfaceHeritage struct {
	BaseNode
	Expression     Expression                    `json:"expression"`
	TypeParameters *TSTypeParameterInstantiation `json:"typeParameters"`
}

// TSTypeAliasDeclaration represents a type alias declaration.
type TSTypeAliasDeclaration struct {
	BaseNode
	ID             *Identifier                 `json:"id"`
	TypeAnnotation TSType                      `json:"typeAnnotation"`
	TypeParameters *TSTypeParameterDeclaration `json:"typeParameters"`
	Declare        bool                        `json:"declare"`
}

func (t *TSTypeAliasDeclaration) statementNode()   {}
func (t *TSTypeAliasDeclaration) declarationNode() {}

// TSEnumDeclaration represents an enum declaration.
type TSEnumDeclaration struct {
	BaseNode
	ID      *Identifier    `json:"id"`
	Members []TSEnumMember `json:"members"`
	Const   bool           `json:"const"`
	Declare bool           `json:"declare"`
}

func (t *TSEnumDeclaration) statementNode()   {}
func (t *TSEnumDeclaration) declarationNode() {}

// TSEnumMember represents a member of an enum.
type TSEnumMember struct {
	BaseNode
	ID          Node        `json:"id"` // Identifier or SimpleLiteral
	Initializer *Expression `json:"initializer"`
}

// TSModuleDeclaration represents a namespace/module declaration.
type TSModuleDeclaration struct {
	BaseNode
	ID      Node `json:"id"`   // Identifier or SimpleLiteral
	Body    Node `json:"body"` // TSModuleBlock or TSModuleDeclaration
	Global  bool `json:"global"`
	Declare bool `json:"declare"`
}

func (t *TSModuleDeclaration) statementNode()   {}
func (t *TSModuleDeclaration) declarationNode() {}

// TSModuleBlock represents a module block.
type TSModuleBlock struct {
	BaseNode
	Body []Statement `json:"body"`
}

// TSNamespaceExportDeclaration represents a namespace export declaration.
type TSNamespaceExportDeclaration struct {
	BaseNode
	ID *Identifier `json:"id"`
}

func (t *TSNamespaceExportDeclaration) statementNode()   {}
func (t *TSNamespaceExportDeclaration) declarationNode() {}

// TSImportEqualsDeclaration represents an import equals declaration.
type TSImportEqualsDeclaration struct {
	BaseNode
	ID              *Identifier `json:"id"`
	ModuleReference Node        `json:"moduleReference"` // TSEntityName or TSExternalModuleReference
	IsExport        bool        `json:"isExport"`
}

func (t *TSImportEqualsDeclaration) statementNode()   {}
func (t *TSImportEqualsDeclaration) declarationNode() {}

// TSExternalModuleReference represents an external module reference (require()).
type TSExternalModuleReference struct {
	BaseNode
	Expression Expression `json:"expression"`
}

// TSAsExpression represents a type assertion using 'as'.
type TSAsExpression struct {
	BaseNode
	Expression     Expression `json:"expression"`
	TypeAnnotation TSType     `json:"typeAnnotation"`
}

func (t *TSAsExpression) expressionNode() {}

// TSTypeAssertion represents a type assertion using angle brackets.
type TSTypeAssertion struct {
	BaseNode
	TypeAnnotation TSType     `json:"typeAnnotation"`
	Expression     Expression `json:"expression"`
}

func (t *TSTypeAssertion) expressionNode() {}

// TSNonNullExpression represents a non-null assertion (expr!).
type TSNonNullExpression struct {
	BaseNode
	Expression Expression `json:"expression"`
}

func (t *TSNonNullExpression) expressionNode() {}

// TSParameterProperty represents a parameter property in a constructor.
type TSParameterProperty struct {
	BaseNode
	Parameter     Node   `json:"parameter"`     // Identifier, AssignmentPattern, or BindingPattern
	Accessibility string `json:"accessibility"` // "public", "protected", "private", or ""
	Readonly      bool   `json:"readonly"`
	Static        bool   `json:"static"`
	Override      bool   `json:"override"`
}

// TSAbstractKeyword represents the abstract keyword.
type TSAbstractKeyword struct {
	BaseNode
}

// TSAccessibility represents accessibility modifiers.
type TSAccessibility string

const (
	TSAccessibilityPublic    TSAccessibility = "public"
	TSAccessibilityProtected TSAccessibility = "protected"
	TSAccessibilityPrivate   TSAccessibility = "private"
)

// TSDeclareFunction represents a declare function.
type TSDeclareFunction struct {
	BaseNode
	ID             *Identifier                 `json:"id"`
	Params         []Node                      `json:"params"`
	ReturnType     *TSTypeAnnotation           `json:"returnType"`
	TypeParameters *TSTypeParameterDeclaration `json:"typeParameters"`
	Generator      bool                        `json:"generator"`
	Async          bool                        `json:"async"`
	Declare        bool                        `json:"declare"`
}

func (t *TSDeclareFunction) statementNode()   {}
func (t *TSDeclareFunction) declarationNode() {}

// TSExportAssignment represents an export assignment (export = x).
type TSExportAssignment struct {
	BaseNode
	Expression Expression `json:"expression"`
}

func (t *TSExportAssignment) statementNode() {}

// TSSatisfiesExpression represents a satisfies expression (expr satisfies Type).
type TSSatisfiesExpression struct {
	BaseNode
	Expression     Expression `json:"expression"`
	TypeAnnotation TSType     `json:"typeAnnotation"`
}

func (t *TSSatisfiesExpression) expressionNode() {}

// TSInstantiationExpression represents a type instantiation expression (expr<T>).
type TSInstantiationExpression struct {
	BaseNode
	Expression     Expression                    `json:"expression"`
	TypeParameters *TSTypeParameterInstantiation `json:"typeParameters"`
}

func (t *TSInstantiationExpression) expressionNode() {}

// TSTemplateLiteralType represents a template literal type.
type TSTemplateLiteralType struct {
	BaseNode
	Quasis []TemplateElement `json:"quasis"`
	Types  []TSType          `json:"types"`
}

func (t *TSTemplateLiteralType) tsTypeNode() {}

// TSImportType represents an import type (import("module").Type).
type TSImportType struct {
	BaseNode
	Argument       TSType                        `json:"argument"`
	Qualifier      Node                          `json:"qualifier"`
	TypeParameters *TSTypeParameterInstantiation `json:"typeParameters"`
}

func (t *TSImportType) tsTypeNode() {}

// TSNamedTupleMember represents a named tuple member ([name: Type]).
type TSNamedTupleMember struct {
	BaseNode
	ElementType TSType      `json:"elementType"`
	Label       *Identifier `json:"label"`
	Optional    bool        `json:"optional"`
}

func (t *TSNamedTupleMember) tsTypeNode() {}
