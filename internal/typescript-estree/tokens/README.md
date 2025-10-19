# TypeScript-ESTree Token Converter

This package provides utilities for converting TypeScript tokens to ESTree-compatible token format, ported from [typescript-eslint](https://github.com/typescript-eslint/typescript-eslint).

## Overview

The token converter handles the extraction and transformation of tokens from TypeScript source files into the ESTree token format used by ESLint and other JavaScript tooling.

## Components

### Token Type Mapping

The `GetTokenType` function maps TypeScript `SyntaxKind` values to ESTree token types:

```go
func GetTokenType(token *ast.Node) types.TokenType
```

**Supported Token Types:**
- `TokenBoolean` - Boolean literals (`true`, `false`)
- `TokenNull` - Null keyword
- `TokenNumeric` - Numeric literals (including BigInt)
- `TokenString` - String literals
- `TokenRegularExpression` - Regular expression literals
- `TokenTemplate` - Template literal parts
- `TokenIdentifier` - Identifiers (including private identifiers)
- `TokenKeyword` - JavaScript/TypeScript keywords
- `TokenPunctuator` - Operators and punctuation
- `TokenJSXText` - JSX text content
- `TokenJSXIdentifier` - JSX identifiers

### Token Conversion

The `ConvertToken` function converts a TypeScript AST token node into an ESTree token:

```go
func ConvertToken(token *ast.Node, sourceFile *ast.SourceFile) *types.Token
```

This function:
- Extracts the token's text value from the source
- Determines the appropriate ESTree token type
- Calculates source position and location information
- Handles special cases (e.g., private identifiers with `#` prefix)

### Token Extraction

The `ConvertTokens` function extracts all tokens from a source file:

```go
func ConvertTokens(sourceFile *ast.SourceFile) []*types.Token
```

**Note:** This is currently a placeholder. Full implementation requires AST traversal logic that will be added in future iterations.

## Implementation Details

### TypeScript SyntaxKind to ESTree Mapping

The converter handles the following mappings:

| TypeScript SyntaxKind | ESTree TokenType |
|----------------------|------------------|
| `NullKeyword` | `Null` |
| `TrueKeyword`, `FalseKeyword` | `Boolean` |
| `NumericLiteral`, `BigIntLiteral` | `Numeric` |
| `StringLiteral`, `NoSubstitutionTemplateLiteral` | `String` |
| `RegularExpressionLiteral` | `RegularExpression` |
| `TemplateHead`, `TemplateMiddle`, `TemplateTail` | `Template` |
| `Identifier`, `PrivateIdentifier` | `Identifier` |
| Keywords (FirstKeyword-LastKeyword range) | `Keyword` |
| Operators and punctuation | `Punctuator` |
| `JsxText`, `JsxTextAllWhiteSpaces` | `JSXText` |

### Helper Functions

**`isKeyword(kind ast.Kind) bool`**
- Checks if a SyntaxKind represents a keyword
- Uses TypeScript's keyword range (FirstKeyword to LastKeyword)

**`isPunctuator(kind ast.Kind) bool`**
- Checks if a SyntaxKind represents a punctuator/operator
- Explicitly lists all punctuator token kinds

**`isToken(node *ast.Node) bool`**
- Determines if a node is a token (leaf node) vs. a composite AST node
- Used for filtering during AST traversal

## Usage Example

```go
import (
    "github.com/microsoft/typescript-go/shim/ast"
    "github.com/web-infra-dev/rslint/internal/typescript-estree/tokens"
)

// Convert a single token
func convertSingleToken(tokenNode *ast.Node, sourceFile *ast.SourceFile) {
    token := tokens.ConvertToken(tokenNode, sourceFile)
    fmt.Printf("Token: type=%s, value=%s, range=[%d,%d]\n",
        token.Type, token.Value, token.Range[0], token.Range[1])
}

// Get token type
func getType(node *ast.Node) {
    tokenType := tokens.GetTokenType(node)
    fmt.Printf("Token type: %s\n", tokenType)
}
```

## Testing

The package includes comprehensive unit tests covering:

- Token type mapping for all major token kinds
- Keyword detection
- Punctuator detection
- Token identification
- ESTree token type constants

Run tests with:

```bash
go test ./internal/typescript-estree/tokens/...
```

## References

- [TypeScript-ESTree node-utils.ts](https://github.com/typescript-eslint/typescript-eslint/blob/main/packages/typescript-estree/src/node-utils.ts)
- [ESTree Token Specification](https://github.com/estree/estree/blob/master/es5.md#tokens)
- [TypeScript Scanner API](https://github.com/microsoft/TypeScript/blob/main/src/compiler/scanner.ts)

## Future Work

- Complete AST traversal implementation in `ConvertTokens`
- Add support for comment token extraction
- Optimize token extraction performance
- Add integration tests with real TypeScript source files
- Handle edge cases for JSX and template literal tokens
