# ESTree Type Definitions for TypeScript

This package provides comprehensive Go type definitions for ESTree-compliant Abstract Syntax Trees (ASTs), including full support for TypeScript-specific extensions (TSESTree).

## Overview

The types in this package represent the complete ESTree specification from ES5 through ES2022+, plus all TypeScript-specific node types. These types are used to represent JavaScript and TypeScript code as structured data that can be analyzed, transformed, and manipulated programmatically.

## Package Structure

The types are organized into logical files:

- **`positions.go`** - Source location types (Position, SourceLocation, Range)
- **`base.go`** - Core node interfaces and base types (Node, Statement, Expression, etc.)
- **`expressions.go`** - All expression node types (BinaryExpression, CallExpression, etc.)
- **`statements.go`** - All statement and declaration types (IfStatement, FunctionDeclaration, etc.)
- **`patterns.go`** - Destructuring pattern types (ArrayPattern, ObjectPattern, etc.)
- **`typescript.go`** - TypeScript-specific node types (TSInterfaceDeclaration, TSTypeAnnotation, etc.)
- **`tokens.go`** - Token and comment type definitions

## Core Interfaces

### Node

All AST nodes implement the `Node` interface:

```go
type Node interface {
    Type() string                  // Returns the node type (e.g., "Identifier")
    Loc() *SourceLocation          // Returns source location information
    GetRange() Range               // Returns character range [start, end]
}
```

### Marker Interfaces

The package defines several marker interfaces to categorize nodes:

- **`Statement`** - All statement nodes
- **`Expression`** - All expression nodes
- **`Declaration`** - All declaration nodes (subset of statements)
- **`Pattern`** - All pattern nodes (used in destructuring)
- **`ModuleDeclaration`** - All module-level declarations
- **`TSType`** - All TypeScript type nodes

## Base Types

### BaseNode

All concrete node types embed `BaseNode`, which provides the standard implementation of the `Node` interface:

```go
type BaseNode struct {
    NodeType string          `json:"type"`
    Location *SourceLocation `json:"loc,omitempty"`
    Span     Range           `json:"range,omitempty"`
}
```

### Position and Location

```go
type Position struct {
    Line   int `json:"line"`   // 1-indexed line number
    Column int `json:"column"` // 0-indexed column number
}

type SourceLocation struct {
    Start Position `json:"start"`
    End   Position `json:"end"`
}

type Range [2]int  // Character offsets: [start, end]
```

## Expression Types

The package includes all ESTree expression types:

- **Literals**: `SimpleLiteral`, `RegExpLiteral`, `BigIntLiteral`
- **Identifiers**: `Identifier`, `PrivateIdentifier`
- **Operators**: `UnaryExpression`, `BinaryExpression`, `LogicalExpression`, `UpdateExpression`
- **Functions**: `FunctionExpression`, `ArrowFunctionExpression`
- **Objects**: `ObjectExpression`, `ArrayExpression`
- **Member Access**: `MemberExpression`, `CallExpression`, `NewExpression`
- **Conditionals**: `ConditionalExpression`
- **Templates**: `TemplateLiteral`, `TaggedTemplateExpression`
- **Classes**: `ClassExpression`
- **Modern Features**: `AwaitExpression`, `YieldExpression`, `ChainExpression`, `ImportExpression`

## Statement Types

All statement types are included:

- **Control Flow**: `IfStatement`, `SwitchStatement`, `SwitchCase`
- **Loops**: `ForStatement`, `ForInStatement`, `ForOfStatement`, `WhileStatement`, `DoWhileStatement`
- **Blocks**: `BlockStatement`, `EmptyStatement`
- **Declarations**: `VariableDeclaration`, `FunctionDeclaration`, `ClassDeclaration`
- **Exception Handling**: `TryStatement`, `CatchClause`, `ThrowStatement`
- **Flow Control**: `ReturnStatement`, `BreakStatement`, `ContinueStatement`
- **Modules**: `ImportDeclaration`, `ExportNamedDeclaration`, `ExportDefaultDeclaration`, `ExportAllDeclaration`
- **Other**: `ExpressionStatement`, `LabeledStatement`, `WithStatement`, `DebuggerStatement`

## Pattern Types

Destructuring patterns for ES2015+:

- `ArrayPattern` - Array destructuring `[a, b, c]`
- `ObjectPattern` - Object destructuring `{x, y, z}`
- `AssignmentPattern` - Default values `{x = 10}`
- `RestElement` - Rest patterns `{...rest}`
- `AssignmentProperty` - Object pattern properties

## TypeScript Types

### Type Annotations

- `TSTypeAnnotation` - Wraps a type annotation
- `TSType` - Interface for all TypeScript type nodes

### Primitive Type Keywords

All TypeScript primitive types:

```
TSAnyKeyword, TSUnknownKeyword, TSNumberKeyword, TSBooleanKeyword,
TSStringKeyword, TSSymbolKeyword, TSVoidKeyword, TSNullKeyword,
TSUndefinedKeyword, TSNeverKeyword, TSBigIntKeyword, TSObjectKeyword
```

### Complex Types

- **Union/Intersection**: `TSUnionType`, `TSIntersectionType`
- **Arrays/Tuples**: `TSArrayType`, `TSTupleType`
- **Functions**: `TSFunctionType`, `TSConstructorType`
- **Literals**: `TSLiteralType`, `TSTemplateLiteralType`
- **References**: `TSTypeReference`, `TSQualifiedName`
- **Advanced**: `TSConditionalType`, `TSInferType`, `TSMappedType`, `TSIndexedAccessType`
- **Queries**: `TSTypeQuery`, `TSImportType`
- **Modifiers**: `TSOptionalType`, `TSRestType`

### Declarations

- `TSInterfaceDeclaration` - Interface declarations
- `TSTypeAliasDeclaration` - Type alias declarations
- `TSEnumDeclaration` - Enum declarations
- `TSModuleDeclaration` - Namespace/module declarations

### Expressions

- `TSAsExpression` - Type assertions with `as`
- `TSTypeAssertion` - Type assertions with `<Type>`
- `TSNonNullExpression` - Non-null assertions `expr!`
- `TSSatisfiesExpression` - Satisfies expressions (TypeScript 4.9+)
- `TSInstantiationExpression` - Type instantiation `expr<T>`

## Token Types

The package defines token types for lexical analysis:

```go
type TokenType string

const (
    TokenBoolean    TokenType = "Boolean"
    TokenNull       TokenType = "Null"
    TokenNumeric    TokenType = "Numeric"
    TokenString     TokenType = "String"
    TokenIdentifier TokenType = "Identifier"
    TokenKeyword    TokenType = "Keyword"
    TokenPunctuator TokenType = "Punctuator"
    // ... and more
)
```

All JavaScript and TypeScript keywords are defined as constants, along with all punctuators.

## JSON Serialization

All types include `json` struct tags for seamless serialization/deserialization:

```go
program := &types.Program{
    BaseNode: types.BaseNode{
        NodeType: "Program",
        Location: &types.SourceLocation{
            Start: types.Position{Line: 1, Column: 0},
            End:   types.Position{Line: 1, Column: 10},
        },
        Span: types.Range{0, 10},
    },
    SourceType: "module",
    Body:       []types.Statement{},
}

data, err := json.Marshal(program)
// Produces ESTree-compliant JSON
```

## Usage Examples

### Creating an AST Node

```go
// Create an identifier
id := &types.Identifier{
    BaseNode: types.BaseNode{
        NodeType: "Identifier",
        Location: &types.SourceLocation{
            Start: types.Position{Line: 1, Column: 0},
            End:   types.Position{Line: 1, Column: 3},
        },
        Span: types.Range{0, 3},
    },
    Name: "foo",
}

// Create a variable declaration
varDecl := &types.VariableDeclaration{
    BaseNode: types.BaseNode{NodeType: "VariableDeclaration"},
    Declarations: []types.VariableDeclarator{
        {
            BaseNode: types.BaseNode{NodeType: "VariableDeclarator"},
            ID:       id,
        },
    },
    Kind: "const",
}
```

### TypeScript Type Annotations

```go
// Create a TypeScript interface
iface := &types.TSInterfaceDeclaration{
    BaseNode: types.BaseNode{NodeType: "TSInterfaceDeclaration"},
    ID: &types.Identifier{
        BaseNode: types.BaseNode{NodeType: "Identifier"},
        Name:     "MyInterface",
    },
    Body: &types.TSInterfaceBody{
        BaseNode: types.BaseNode{NodeType: "TSInterfaceBody"},
        Body:     []types.Node{},
    },
    Extends: []types.TSInterfaceHeritage{},
}
```

## References

- [ESTree Specification](https://github.com/estree/estree)
- [TypeScript-ESTree](https://github.com/typescript-eslint/typescript-eslint/tree/main/packages/typescript-estree)
- [TypeScript-ESTree AST Spec](https://typescript-eslint.io/packages/typescript-estree/ast-spec/)
- [@types/estree](https://www.npmjs.com/package/@types/estree)

## Testing

The package includes comprehensive unit tests covering:

- Node interface implementation
- JSON serialization/deserialization
- Type correctness
- All major node types

Run tests with:

```bash
go test ./types/...
```

## Contributing

When adding new node types:

1. Place them in the appropriate file based on their category
2. Ensure they embed `BaseNode`
3. Implement the appropriate marker interface(s)
4. Add `json` struct tags for all fields
5. Add comprehensive documentation comments
6. Include unit tests
