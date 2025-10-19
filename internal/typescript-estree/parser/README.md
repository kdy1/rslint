# TypeScript ESTree Parser

This package provides the main parsing functionality for converting TypeScript source code into ESTree-compliant AST nodes, matching the API of [@typescript-eslint/typescript-estree](https://typescript-eslint.io/packages/typescript-estree/).

## Overview

The parser acts as a bridge between the TypeScript compiler (via typescript-go) and the ESTree AST format used by ESLint and other JavaScript tooling. It provides two main parsing functions:

- **`Parse()`** - Basic parsing without type information
- **`ParseAndGenerateServices()`** - Full parsing with TypeScript Program services for type checking

## API

### Parse

```go
func Parse(source string, options *ParseOptions) (*types.Program, error)
```

Parses TypeScript/JavaScript source code and returns an ESTree-compliant AST.

**Example:**

```go
ast, err := parser.Parse("const x = 42;", &parser.ParseOptions{
    FilePath:   "example.ts",
    SourceType: "script",
    Loc:        true,
    Range:      true,
})
```

### ParseAndGenerateServices

```go
func ParseAndGenerateServices(source string, options *ParseOptions) (*ParseResult, error)
```

Parses source code and generates parser services that provide access to TypeScript type information.

**Example:**

```go
result, err := parser.ParseAndGenerateServices("const x: number = 42;", &parser.ParseOptions{
    FilePath: "example.ts",
    Project:  "./tsconfig.json",
    Loc:      true,
    Range:    true,
})

// Access type information via result.Services.Program
```

## ParseOptions

Configuration options for parsing, matching the typescript-estree API:

```go
type ParseOptions struct {
    // SourceType specifies whether to parse as "script" or "module"
    SourceType string

    // AllowInvalidAST prevents throwing errors on invalid ASTs
    AllowInvalidAST bool

    // Comment creates a top-level comments array
    Comment bool

    // SuppressDeprecatedPropertyWarnings skips warnings for deprecated properties
    SuppressDeprecatedPropertyWarnings bool

    // DebugLevel controls debugging output (bool or []string)
    DebugLevel interface{}

    // ErrorOnUnknownASTType throws error on unknown AST node types
    ErrorOnUnknownASTType bool

    // FilePath is the path to the file being parsed
    FilePath string

    // JSDocParsingMode controls JSDoc comment parsing ("all", "none", "type-info")
    JSDocParsingMode JSDocParsingMode

    // JSX enables JSX syntax parsing
    JSX bool

    // Loc includes location information (line/column) for nodes
    Loc bool

    // LoggerFn overrides logging (func(string) or bool)
    LoggerFn interface{}

    // Range includes [start, end] byte offsets for nodes
    Range bool

    // Tokens creates a top-level array of tokens
    Tokens bool

    // Project specifies path to tsconfig.json
    Project string

    // TsconfigRootDir specifies root directory for tsconfig paths
    TsconfigRootDir string

    // Programs provides pre-created TypeScript programs
    Programs []*compiler.Program
}
```

## ParserServices

When using `ParseAndGenerateServices()`, you get access to parser services:

```go
type ParserServices struct {
    // Program is the TypeScript compiler program
    Program *compiler.Program

    // ESTreeNodeToTSNodeMap maps ESTree nodes to TypeScript AST nodes
    ESTreeNodeToTSNodeMap map[types.Node]*ast.Node

    // TSNodeToESTreeNodeMap maps TypeScript AST nodes to ESTree nodes
    TSNodeToESTreeNodeMap map[*ast.Node]types.Node
}
```

These services allow you to:
- Access TypeScript type checker for type information
- Navigate between ESTree and TypeScript AST representations
- Leverage full TypeScript compiler capabilities

## Implementation Status

### ✅ Completed

- ParseOptions structure with all typescript-estree fields
- ParseSettings for internal configuration
- ParserServices structure for type information
- Parse() function entry point
- ParseAndGenerateServices() function entry point
- TypeScript version checking
- Integration with typescript-go compiler
- Integration with converter package
- Comprehensive test coverage for API

### 🚧 In Progress

- **Standalone parsing** - Currently requires a TypeScript Program/Project configuration. Direct source file creation API is not yet available in the typescript-go shim.
- **Full AST conversion** - The converter package has placeholder implementation for node conversion. Actual node-by-node conversion logic will be implemented in a follow-up PR.

### 📋 TODO

- Implement standalone source file parsing (waiting on typescript-go API)
- Add support for dynamic source updates in existing programs
- Implement diagnostic collection and formatting
- Add performance optimizations for batch parsing
- Support for custom module resolution
- JSX transformation options

## Architecture

```
┌─────────────────────┐
│   Parse/ParseAndGen │  Entry points matching typescript-estree API
│      Services       │
└──────────┬──────────┘
           │
           ├──> ParseSettings (config normalization)
           │
           ├──> parseWithTypeScript
           │    ├──> typescript-go compiler
           │    └──> TypeScript Program (optional)
           │
           └──> converter.ConvertProgram
                ├──> ESTree AST generation
                └──> Node mappings (optional)
```

## Differences from typescript-estree

1. **Language**: Go instead of TypeScript/JavaScript
2. **TypeScript Integration**: Uses typescript-go instead of direct TypeScript compiler
3. **Error Handling**: Go-style error returns instead of exceptions
4. **Standalone Parsing**: Not yet fully supported (requires Program/Project config)

## Examples

### Basic Script Parsing

```go
source := `
function greet(name) {
    console.log("Hello, " + name);
}
`

ast, err := parser.Parse(source, &parser.ParseOptions{
    FilePath:   "greet.js",
    SourceType: "script",
    Loc:        true,
    Range:      true,
})

if err != nil {
    log.Fatal(err)
}

// ast is now an ESTree Program node
fmt.Printf("Parsed %d statements\n", len(ast.Body))
```

### Module with Type Information

```go
source := `
export function add(a: number, b: number): number {
    return a + b;
}
`

result, err := parser.ParseAndGenerateServices(source, &parser.ParseOptions{
    FilePath: "math.ts",
    Project:  "./tsconfig.json",
    Loc:      true,
    Range:    true,
    Tokens:   true,
})

if err != nil {
    log.Fatal(err)
}

// Access type information
program := result.Services.Program
// Use program for type checking...
```

### JSX Parsing

```go
source := `
const App = () => {
    return <div className="app">Hello World</div>;
};
`

ast, err := parser.Parse(source, &parser.ParseOptions{
    FilePath: "App.tsx",
    JSX:      true,
    Loc:      true,
    Range:    true,
})
```

## Testing

Run the parser tests:

```bash
go test ./internal/typescript-estree/parser/...
```

Run with verbose output:

```bash
go test -v ./internal/typescript-estree/parser/...
```

## References

- [typescript-eslint/typescript-estree](https://github.com/typescript-eslint/typescript-eslint/tree/main/packages/typescript-estree)
- [typescript-estree Documentation](https://typescript-eslint.io/packages/typescript-estree/)
- [ESTree Specification](https://github.com/estree/estree)
- [TypeScript-Go](https://github.com/microsoft/typescript-go)
- [TypeScript Compiler API](https://github.com/microsoft/TypeScript/wiki/Using-the-Compiler-API)

## Contributing

When adding new parser features:

1. Match the typescript-estree API when possible
2. Add comprehensive tests
3. Update this README with new functionality
4. Document any deviations from typescript-estree behavior
