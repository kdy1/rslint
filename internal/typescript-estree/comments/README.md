# Comments Package

This package provides functionality for extracting and converting TypeScript comments to ESTree comment format.

## Overview

The comments package handles the extraction of comments from TypeScript source code and converts them to the ESTree-compliant comment format. It properly handles:

- Line comments (`//`)
- Block comments (`/* */`)
- JSDoc comments (`/** */`)
- Multiline comments
- Comment position and location tracking

## Usage

```go
package main

import (
    "github.com/web-infra-dev/rslint/internal/typescript-estree/comments"
    "github.com/web-infra-dev/rslint/internal/utils"
)

func main() {
    code := `
    // This is a line comment
    const x = 1;

    /* This is a block comment */
    const y = 2;
    `

    // Create TypeScript SourceFile
    sourceFile, err := utils.CreateProgram(code, "example.ts", nil)
    if err != nil {
        panic(err)
    }

    // Extract and convert comments
    estreeComments := comments.ConvertComments(sourceFile, code)

    // Use the comments
    for _, comment := range estreeComments {
        fmt.Printf("Type: %s, Value: %s\n", comment.Type, comment.Value)
    }
}
```

## Comment Structure

Comments are converted to the ESTree `Comment` type:

```go
type Comment struct {
    Type  string          // "Line" or "Block"
    Value string          // The comment text (without delimiters)
    Range [2]int          // Start and end positions in source
    Loc   *SourceLocation // Line and column information
}
```

## Comment Types

### Line Comments

Line comments start with `//` and continue to the end of the line.

**Input:**
```typescript
// This is a line comment
```

**Output:**
```go
Comment{
    Type:  "Line",
    Value: " This is a line comment",
    Range: [0, 26],
    Loc:   &SourceLocation{...},
}
```

### Block Comments

Block comments start with `/*` and end with `*/`.

**Input:**
```typescript
/* This is a block comment */
```

**Output:**
```go
Comment{
    Type:  "Block",
    Value: " This is a block comment ",
    Range: [0, 30],
    Loc:   &SourceLocation{...},
}
```

### JSDoc Comments

JSDoc comments are a special form of block comments that start with `/**`.

**Input:**
```typescript
/**
 * Function description
 * @param x Parameter description
 */
```

**Output:**
```go
Comment{
    Type:  "Block",
    Value: "*\n * Function description\n * @param x Parameter description\n ",
    Range: [0, ...],
    Loc:   &SourceLocation{...},
}
```

## Implementation Details

### Comment Extraction

The `ConvertComments` function uses TypeScript's scanner utilities to extract comments:

1. **Get Comment Ranges**: Uses `scanner.GetLeadingCommentRanges` to find all comments in the source
2. **Convert to ESTree**: Converts each TypeScript `CommentRange` to an ESTree `Comment`
3. **Extract Text**: Removes comment delimiters (`//` for line, `/*` and `*/` for block)
4. **Calculate Locations**: Converts byte positions to line/column locations

### Position Handling

- **Range**: Uses 0-based byte offsets for start and end positions
- **Line Numbers**: ESTree uses 1-based line numbers (TypeScript uses 0-based)
- **Columns**: Uses 0-based column numbers

### Comment Value Extraction

The comment value excludes the delimiters:

- Line comments: Remove `//` prefix
- Block comments: Remove `/*` prefix and `*/` suffix

## Reference Implementation

This implementation is based on the TypeScript-ESTree convert-comments module:
- [convert-comments.ts](https://github.com/typescript-eslint/typescript-eslint/blob/main/packages/typescript-estree/src/convert-comments.ts)

## Testing

The package includes comprehensive tests covering:

- Single line comments
- Block comments
- Multiple comments
- Empty comments
- JSDoc comments
- Multiline block comments
- Position and location tracking

Run tests with:

```bash
go test ./internal/typescript-estree/comments/...
```

## Integration

This package is designed to be used as part of the TypeScript-ESTree parser pipeline:

1. Parse TypeScript source to AST
2. Extract comments using `ConvertComments`
3. Attach comments to the ESTree AST nodes
4. Return complete ESTree with comments

## Future Enhancements

Potential future improvements:

- Comment attachment to AST nodes (leading/trailing/inner)
- Comment filtering based on parser options
- Special handling for directive comments
- Comment preservation during AST transformations
