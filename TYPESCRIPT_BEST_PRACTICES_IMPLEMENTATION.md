# TypeScript Best Practices Rules Implementation Plan

## Overview

This document outlines the implementation plan for adding TypeScript-ESLint best practices rules to RSLint. These rules enforce TypeScript idioms and best practices.

## Rules to Implement

### High Priority Rules (Quick Wins)

1. **prefer-optional-chain** - Use optional chaining instead of logical AND chains

   - AST Nodes: `BinaryExpression`, `ConditionalExpression`
   - Has autofix: Yes
   - Example: `foo && foo.bar && foo.bar.baz` → `foo?.bar?.baz`

2. **prefer-nullish-coalescing** - Use nullish coalescing instead of logical OR

   - AST Nodes: `BinaryExpression`
   - Has autofix: Yes
   - Example: `foo || 'default'` → `foo ?? 'default'`

3. **prefer-string-starts-ends-with** - Use startsWith/endsWith instead of complex comparisons

   - AST Nodes: `CallExpression`, `BinaryExpression`
   - Has autofix: Yes
   - Example: `str.indexOf('test') === 0` → `str.startsWith('test')`

4. **prefer-includes** - Use includes instead of indexOf comparisons

   - AST Nodes: `BinaryExpression`
   - Has autofix: Yes
   - Example: `arr.indexOf(x) !== -1` → `arr.includes(x)`

5. **ban-ts-comment** - Ban @ts-<directive> comments or require descriptions

   - AST Nodes: Custom comment processing
   - Has options: Yes (allow with description, specific directives)
   - Example: Disallow `// @ts-ignore` without explanation

6. **prefer-ts-expect-error** - Prefer @ts-expect-error over @ts-ignore

   - AST Nodes: Custom comment processing
   - Has autofix: Yes
   - Example: `// @ts-ignore` → `// @ts-expect-error`

7. **no-non-null-assertion** - Disallow non-null assertions

   - AST Nodes: `NonNullExpression`
   - Example: Disallow `foo!.bar`

8. **consistent-type-definitions** - Enforce interface or type for object type definitions

   - AST Nodes: `InterfaceDeclaration`, `TypeAliasDeclaration`
   - Has options: Yes ("interface" or "type")
   - Has autofix: Yes
   - Example: Enforce `type` over `interface` for object types

9. **consistent-type-imports** - Enforce type-only imports when possible
   - AST Nodes: `ImportDeclaration`
   - Has autofix: Yes
   - Example: `import { Type }` → `import type { Type }`

### Medium Priority Rules

10. **prefer-for-of** - Prefer for-of loops over traditional for loops

    - AST Nodes: `ForStatement`
    - Has autofix: Yes (in safe cases)

11. **prefer-function-type** - Use function types instead of interfaces with call signatures

    - AST Nodes: `InterfaceDeclaration`
    - Has autofix: Yes

12. **prefer-return-this-type** - Enforce return type of `this` in method chaining

    - AST Nodes: `MethodDeclaration`
    - Requires types: Yes

13. **ban-types** - Ban specific types

    - AST Nodes: `TypeReference`
    - Has options: Yes (custom type bans)

14. **no-non-null-asserted-optional-chain** - Disallow `foo?.bar!`

    - AST Nodes: `NonNullExpression`

15. **no-confusing-non-null-assertion** - Avoid confusing placements of non-null assertions

    - AST Nodes: `NonNullExpression`

16. **consistent-type-exports** - Enforce type-only exports

    - AST Nodes: `ExportDeclaration`
    - Has autofix: Yes

17. **consistent-indexed-object-style** - Enforce index signature or Record<> style

    - AST Nodes: `TypeLiteral`, `TypeReference`
    - Has options: Yes

18. **consistent-generic-constructors** - Enforce generic constructor style

    - AST Nodes: `NewExpression`

19. **no-import-type-side-effects** - Avoid import type with side effects

    - AST Nodes: `ImportDeclaration`

20. **prefer-enum-initializers** - Require explicit enum member values
    - AST Nodes: `EnumDeclaration`

### Already Implemented

- **no-require-imports** - ✅ Already exists in codebase
- **no-var-requires** - ✅ Already exists in codebase

### Lower Priority / Complex Rules

21. **prefer-readonly** - Require readonly modifiers

    - AST Nodes: `PropertyDeclaration`, `ParameterDeclaration`
    - Requires types: Yes (to detect if property is reassigned)

22. **prefer-readonly-parameter-types** - Require readonly on function parameters
    - AST Nodes: `Parameter`
    - Requires types: Yes

## Implementation Strategy

### Phase 1: Rule Generation (Current)

Use the scaffolding tool to generate boilerplate for each rule:

```bash
# Generate individual rules
go run scripts/generate-rule.go -rule prefer-optional-chain -plugin typescript-eslint -fetch -has-autofix
go run scripts/generate-rule.go -rule prefer-nullish-coalescing -plugin typescript-eslint -fetch -has-autofix
go run scripts/generate-rule.go -rule prefer-string-starts-ends-with -plugin typescript-eslint -fetch -has-autofix
go run scripts/generate-rule.go -rule prefer-includes -plugin typescript-eslint -fetch -has-autofix
go run scripts/generate-rule.go -rule ban-ts-comment -plugin typescript-eslint -fetch -has-options
go run scripts/generate-rule.go -rule prefer-ts-expect-error -plugin typescript-eslint -fetch -has-autofix
go run scripts/generate-rule.go -rule no-non-null-assertion -plugin typescript-eslint -fetch
go run scripts/generate-rule.go -rule consistent-type-definitions -plugin typescript-eslint -fetch -has-options -has-autofix
go run scripts/generate-rule.go -rule consistent-type-imports -plugin typescript-eslint -fetch -has-autofix
```

### Phase 2: Rule Implementation

For each rule:

1. Study the upstream TypeScript-ESLint implementation
2. Port test cases from https://github.com/typescript-eslint/typescript-eslint/tree/main/packages/eslint-plugin/tests/rules
3. Implement the rule logic using typescript-go AST
4. Implement autofixes where applicable
5. Register the rule in `internal/config/config.go`

### Phase 3: Testing

- Ensure all rules compile
- Run full test suite
- Add integration tests
- Test autofixes with real-world code

### Phase 4: Documentation

- Document each rule in README
- Add examples for each rule
- Document configuration options

## Technical Notes

### AST Node Mapping

- TypeScript AST nodes are accessed via `ast.Kind*` constants
- Use `node.As*()` methods to get typed node access
- Example:
  ```go
  case ast.KindBinaryExpression:
      binExpr := node.AsBinaryExpression()
      if binExpr.OperatorToken == ast.TokenBarBar {
          // Potential prefer-nullish-coalescing violation
      }
  ```

### Autofix Pattern

```go
ctx.ReportNodeWithFixes(node, rule.RuleMessage{
    Id:          "useNullishCoalescing",
    Description: "Prefer nullish coalescing operator (??) over logical or (||)",
}, rule.RuleFixReplace(ctx.SourceFile, node, fixedCode))
```

### Options Parsing Pattern

```go
type RuleOptions struct {
    AllowSomething bool `json:"allowSomething"`
}

func parseOptions(options any) RuleOptions {
    opts := RuleOptions{
        AllowSomething: false, // default
    }

    var optsMap map[string]interface{}
    if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
        optsMap, _ = optArray[0].(map[string]interface{})
    } else {
        optsMap, _ = options.(map[string]interface{})
    }

    if optsMap != nil {
        if v, ok := optsMap["allowSomething"].(bool); ok {
            opts.AllowSomething = v
        }
    }

    return opts
}
```

## Test Case Sources

All test cases should be ported from TypeScript-ESLint:

- Base URL: https://github.com/typescript-eslint/typescript-eslint/blob/main/packages/eslint-plugin/tests/rules/
- Example: `prefer-optional-chain.test.ts`

Each rule should have:

- Minimum 15 test cases
- Valid cases (code that should pass)
- Invalid cases (code that should fail)
- Option variations
- Autofix test cases

## Success Criteria

- ✅ All rules compile successfully
- ✅ All tests pass
- ✅ Autofixes work correctly
- ✅ Rules properly registered in config
- ✅ Documentation complete
- ✅ Integration with existing RSLint infrastructure

## Resources

- [TypeScript-ESLint Rules Documentation](https://typescript-eslint.io/rules/)
- [TypeScript-ESLint Source](https://github.com/typescript-eslint/typescript-eslint)
- [typescript-go AST Reference](https://github.com/microsoft/typescript-go)
- [RSLint Rule Scaffolding Guide](docs/RULE_SCAFFOLDING_GUIDE.md)

## Next Steps

1. Complete typescript-go submodule initialization
2. Run rule generation script for all rules
3. Begin implementation of high-priority rules
4. Port test cases for each rule
5. Implement and test autofixes
6. Register all rules
7. Run full test suite
8. Create comprehensive PR

## Notes

- The `no-require-imports` and `no-var-requires` rules already exist in the codebase
- Focus on high-priority rules first (those with autofixes and high impact)
- Some rules may require TypeScript type checker integration
- Comment-based rules (ban-ts-comment, prefer-ts-expect-error) may need custom comment parsing logic
