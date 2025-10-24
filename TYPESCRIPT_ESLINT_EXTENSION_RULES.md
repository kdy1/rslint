# TypeScript-ESLint Extension Rules Implementation

## Overview

This document tracks the implementation of TypeScript-ESLint extension rules. Extension rules are TypeScript-aware versions of ESLint core rules that properly handle TypeScript-specific syntax.

## What are Extension Rules?

Extension rules extend ESLint's core rules to support TypeScript syntax. They handle:

- Type parameters and generic syntax
- Enums and namespaces
- Decorators
- Abstract classes and members
- TypeScript-specific modifiers (public, private, protected, readonly)
- Optional chaining and nullish coalescing
- Type assertions and annotations

## Implementation Status

### Rules to Implement

The following extension rules need to be implemented:

#### Formatting & Style Rules (High Priority)

- [ ] **block-spacing** - Enforce consistent spacing inside single-line blocks
- [ ] **brace-style** - Enforce consistent brace style for blocks
- [ ] **comma-dangle** - Require or disallow trailing commas
- [ ] **comma-spacing** - Enforce consistent spacing before and after commas
- [ ] **func-call-spacing** - Require or disallow spacing between function identifiers and their invocations
- [ ] **indent** - Enforce consistent indentation
- [ ] **key-spacing** - Enforce consistent spacing between keys and values in object literals
- [ ] **keyword-spacing** - Enforce consistent spacing before and after keywords
- [ ] **lines-around-comment** - Require empty lines around comments
- [ ] **lines-between-class-members** - Require or disallow an empty line between class members
- [ ] **object-curly-spacing** - Enforce consistent spacing inside braces
- [ ] **padding-line-between-statements** - Require or disallow padding lines between statements
- [ ] **quotes** - Enforce the consistent use of backticks, double, or single quotes
- [ ] **semi** - Require or disallow semicolons instead of ASI
- [ ] **space-before-blocks** - Enforce consistent spacing before blocks
- [ ] **space-before-function-paren** - Enforce consistent spacing before function parenthesis
- [ ] **space-infix-ops** - Require spacing around infix operators
- [ ] **no-extra-parens** - Disallow unnecessary parentheses
- [ ] **no-extra-semi** - Disallow unnecessary semicolons

#### Logic & Best Practices Rules (Medium Priority)

- [ ] **class-methods-use-this** - Enforce that class methods utilize `this`
- [ ] **default-param-last** - Enforce default parameters to be last
- [ ] **init-declarations** - Require or disallow initialization in variable declarations
- [ ] **no-array-constructor** - Disallow `Array` constructors
- [ ] **no-dupe-class-members** - Disallow duplicate class members
- [ ] **no-invalid-this** - Disallow `this` keywords outside of classes or class-like objects
- [ ] **no-loop-func** - Disallow function declarations that contain unsafe references inside loop statements
- [ ] **no-loss-of-precision** - Disallow number literals that lose precision
- [ ] **no-magic-numbers** - Disallow magic numbers
- [ ] **no-redeclare** - Disallow variable redeclaration
- [ ] **no-restricted-imports** - Disallow specified modules when loaded by `import`
- [ ] **no-useless-constructor** - Disallow unnecessary constructors

### Already Implemented

- [x] **dot-notation** - Enforce dot notation whenever possible (exists in `internal/rules/dot_notation/`)
- [x] **no-empty-function** - Disallow empty functions (exists in `internal/plugins/typescript/rules/no_empty_function/`)
- [x] **no-implied-eval** - Disallow the use of `eval()`-like methods (exists in `internal/plugins/typescript/rules/no_implied_eval/`)

## Implementation Approach

### 1. Rule Generation

Use the automated scaffolding tool:

```bash
go run scripts/generate-rule.go -rule <rule-name> -plugin typescript-eslint -has-autofix -fetch
```

### 2. TypeScript-Specific Considerations

Each rule must handle:

- **Type Parameters**: `function foo<T>(param: T) {}`
- **Enums**: `enum Foo { A, B }`
- **Namespaces**: `namespace Foo {}`
- **Decorators**: `@decorator class Foo {}`
- **Abstract Members**: `abstract class Foo { abstract method(): void; }`
- **Access Modifiers**: `private`, `protected`, `public`, `readonly`
- **Optional Chaining**: `obj?.prop`
- **Nullish Coalescing**: `value ?? default`
- **Type Assertions**: `value as Type`, `<Type>value`

### 3. Test Requirements

Each rule needs comprehensive tests including:

- Valid TypeScript syntax cases
- Invalid TypeScript syntax cases
- Edge cases with generic types
- Decorator scenarios
- Enum and namespace handling
- Integration with autofix (where applicable)

### 4. Documentation

Each rule should reference:

- TypeScript-ESLint docs: https://typescript-eslint.io/rules/[rule-name]
- ESLint core rule: https://eslint.org/docs/latest/rules/[rule-name]
- Test cases from upstream: https://github.com/typescript-eslint/typescript-eslint/tree/main/packages/eslint-plugin/tests/rules

## Technical Implementation Details

### Rule Structure

```go
package rule_name

import (
    "github.com/microsoft/typescript-go/shim/ast"
    "github.com/web-infra-dev/rslint/internal/rule"
)

// RuleNameRule implements the rule-name extension rule
var RuleNameRule = rule.CreateRule(rule.Rule{
    Name: "rule-name",
    Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
    // Parse options if needed

    return rule.RuleListeners{
        // Listen to relevant AST node types
        ast.KindFunctionDeclaration: func(node *ast.Node) {
            // Check TypeScript-specific syntax
            // Report violations with autofix
        },
    }
}
```

### AST Node Types Commonly Used

- `FunctionDeclaration`, `ArrowFunction`, `MethodDeclaration`
- `ClassDeclaration`, `InterfaceDeclaration`
- `EnumDeclaration`, `ModuleDeclaration` (namespaces)
- `TypeParameterDeclaration`, `TypeReference`
- `Decorator`
- `PropertyDeclaration`, `ParameterDeclaration`

## Next Steps

1. **Initialize Development Environment**

   - Ensure `typescript-go` submodule is initialized
   - Set up Go development environment
   - Install required dependencies

2. **Implement Priority Rules**

   - Start with formatting rules (most commonly used)
   - Focus on rules with clear TypeScript-specific behavior
   - Implement autofixes for style rules

3. **Testing & Validation**

   - Port test cases from TypeScript-ESLint
   - Add custom test cases for edge scenarios
   - Validate against real-world TypeScript code

4. **Documentation & Examples**
   - Document TypeScript-specific behavior
   - Provide migration guide from ESLint core rules
   - Add configuration examples

## Resources

### Official Documentation

- [TypeScript-ESLint Extension Rules](https://typescript-eslint.io/rules/?=extension)
- [ESLint Core Rules](https://eslint.org/docs/latest/rules/)
- [TypeScript Handbook](https://www.typescriptlang.org/docs/handbook/intro.html)

### Test Resources

- [TypeScript-ESLint Test Cases](https://github.com/typescript-eslint/typescript-eslint/tree/main/packages/eslint-plugin/tests/rules)
- [ESLint Rule Tests](https://github.com/eslint/eslint/tree/main/tests/lib/rules)

### Example Implementations

- Existing TypeScript rules in `internal/plugins/typescript/rules/`
- Rule scaffolding tool in `scripts/generate-rule.go`

## Configuration Example

```json
{
  "rules": {
    "@typescript-eslint/brace-style": ["error", "1tbs"],
    "@typescript-eslint/comma-dangle": ["error", "always-multiline"],
    "@typescript-eslint/quotes": ["error", "single"],
    "@typescript-eslint/semi": ["error", "always"],
    "@typescript-eslint/space-before-function-paren": [
      "error",
      {
        "anonymous": "always",
        "named": "never",
        "asyncArrow": "always"
      }
    ]
  }
}
```

## Contributing

When implementing a rule:

1. Check if the rule already exists (see "Already Implemented" section)
2. Generate scaffolding using the provided script
3. Implement TypeScript-specific logic
4. Add comprehensive tests
5. Update this document with implementation status
6. Submit PR with detailed description of TypeScript-specific handling

## Notes

- Many extension rules provide autofix capabilities for formatting issues
- TypeScript-specific syntax requires careful AST traversal
- Consider performance implications when analyzing type information
- Maintain compatibility with ESLint configuration patterns
- Follow existing code patterns in `internal/plugins/typescript/rules/`
