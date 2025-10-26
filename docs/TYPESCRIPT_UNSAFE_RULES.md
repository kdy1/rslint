# TypeScript-ESLint Unsafe Type Rules

This document describes the three critical TypeScript-ESLint type safety rules that have been ported to RSLint to prevent unsafe usage of `any` types.

## Overview

These rules work together to prevent the TypeScript `any` type from silently bypassing the type system, which can lead to runtime errors and reduced code quality. They require TypeScript type information and are essential for maintaining type safety in TypeScript projects.

## Rules

### 1. no-unsafe-argument

**Rule ID**: `@typescript-eslint/no-unsafe-argument`

**Description**: Disallow calling a function with a value with type 'any'

**Category**: Type Safety

**Requires Type Information**: Yes

#### What it does

This rule prevents passing arguments of type `any` to functions that expect specific types. This is dangerous because TypeScript cannot verify that the value matches the expected parameter type.

#### Examples

**Incorrect**:
```typescript
declare function foo(arg: number): void;
foo(1 as any);  // ❌ Unsafe argument of type 'any'

declare function bar(...args: number[]): void;
bar(1, 2, 1 as any);  // ❌ Unsafe argument of type 'any'

// Unsafe spread
declare function baz(arg1: string, arg2: number): void;
baz(...(x as any));  // ❌ Unsafe spread of type 'any'
```

**Correct**:
```typescript
declare function foo(arg: number): void;
foo(42);  // ✅ Properly typed argument

declare function bar(arg: any): void;
bar(1 as any);  // ✅ Function accepts 'any'

declare function baz(...args: any[]): void;
baz(1, 2, 1 as any);  // ✅ Function accepts 'any[]'
```

#### Message IDs

- `unsafeArgument`: "Unsafe argument of an `any` typed value."
- `unsafeSpread`: "Unsafe spread of an `any` typed value."
- `unsafeArraySpread`: "Unsafe spread of an `any[]` typed value."
- `unsafeTupleSpread`: "Unsafe spread of a tuple with `any` elements."

#### Implementation

**Location**: `internal/plugins/typescript/rules/no_unsafe_argument/no_unsafe_argument.go`

**Key Features**:
- Validates regular function arguments
- Checks spread arguments and tuples
- Handles tagged template expressions
- Detects both explicit `any` and error types
- Special handling for `noImplicitThis` compiler option

**Tests**: 448 lines with 22 valid cases and 20 invalid cases

---

### 2. no-unsafe-call

**Rule ID**: `@typescript-eslint/no-unsafe-call`

**Description**: Disallow calling a value with type 'any'

**Category**: Type Safety

**Requires Type Information**: Yes

#### What it does

This rule prevents calling expressions where the callee has type `any`. This is unsafe because TypeScript cannot verify that the value is actually callable or that it will behave as expected.

#### Examples

**Incorrect**:
```typescript
declare const anyValue: any;
anyValue();  // ❌ Unsafe call of an `any` typed value

new anyValue();  // ❌ Unsafe construction of an `any` typed value

anyValue`template`;  // ❌ Unsafe use of an `any` typed template tag

declare const func: Function;
func();  // ❌ `Function` type is unsafe without proper signatures
```

**Correct**:
```typescript
declare const func: () => void;
func();  // ✅ Properly typed function

declare const Constructor: new () => Object;
new Constructor();  // ✅ Properly typed constructor

declare const tag: (strings: TemplateStringsArray) => string;
tag`template`;  // ✅ Properly typed template tag
```

#### Message IDs

- `unsafeCall`: "Unsafe call of a(n) {type} typed value."
- `unsafeCallThis`: "Unsafe call of a(n) {type} typed value. `this` is typed as {type}."
- `unsafeNew`: "Unsafe construction of a(n) {type} typed value."
- `unsafeTemplateTag`: "Unsafe use of a(n) {type} typed template tag."

#### Implementation

**Location**: `internal/plugins/typescript/rules/no_unsafe_call/no_unsafe_call.go`

**Key Features**:
- Validates call expressions
- Checks new expressions (constructors)
- Handles tagged template expressions
- Ignores import expressions (which are safe)
- Special checks for `Function` type with call/construct signatures
- Detects unsafe `this` context when `noImplicitThis` is disabled

**Tests**: 388 lines with 15 valid cases and 21 invalid cases

---

### 3. no-unsafe-member-access

**Rule ID**: `@typescript-eslint/no-unsafe-member-access`

**Description**: Disallow member access on a value with type 'any'

**Category**: Type Safety

**Requires Type Information**: Yes

#### What it does

This rule prevents accessing properties on values with type `any`. Since TypeScript doesn't know what properties exist on `any`, such access is inherently unsafe and can lead to runtime errors.

#### Examples

**Incorrect**:
```typescript
declare const anyValue: any;
anyValue.foo;  // ❌ Unsafe member access on an `any` value

anyValue['bar'];  // ❌ Unsafe member access on an `any` value

declare const obj: { prop: any };
obj.prop.nested;  // ❌ Unsafe member access on an `any` value

const index: any = 'key';
obj[index];  // ❌ Computed name resolves to an `any` value
```

**Correct**:
```typescript
declare const typedValue: { foo: string };
typedValue.foo;  // ✅ Safe access on typed value

declare const record: Record<string, unknown>;
record['bar'];  // ✅ Safe access with proper typing

declare const obj: { prop: { nested: number } };
obj.prop.nested;  // ✅ Fully typed access chain
```

#### Message IDs

- `unsafeMemberExpression`: "Unsafe member access {property} on an {type} value."
- `unsafeThisMemberExpression`: "Unsafe member access {property} on an `any` value. `this` is typed as `any`."
- `unsafeComputedMemberAccess`: "Computed name {property} resolves to an {type} value."

#### Implementation

**Location**: `internal/plugins/typescript/rules/no_unsafe_member_access/no_unsafe_member_access.go`

**Key Features**:
- Monitors property access expressions
- Checks element access expressions
- Caches state to avoid duplicate reports on nested member access
- Handles computed property names
- Ignores heritage clauses (implements/extends)
- Performance optimizations for literal expressions and update operators
- Special handling for unsafe `this` context

**Tests**: 327 lines with 12 valid cases and 15 invalid cases

---

## Usage

### Configuration

These rules are automatically registered in RSLint. To enable them in your `rslint.json`:

```json
{
  "rules": {
    "@typescript-eslint/no-unsafe-argument": "error",
    "@typescript-eslint/no-unsafe-call": "error",
    "@typescript-eslint/no-unsafe-member-access": "error"
  }
}
```

### Prerequisites

These rules require:
1. TypeScript type checking to be enabled
2. A valid `tsconfig.json` in your project
3. TypeScript files to be analyzed

### When to Use

Enable these rules when:
- You want to maintain strict type safety in your TypeScript codebase
- You're migrating from JavaScript and want to eliminate `any` usage
- You need to catch potential runtime errors at compile time
- You want to improve code quality and maintainability

### When Not to Use

You might disable these rules when:
- Working with legacy code that heavily uses `any`
- Interfacing with untyped third-party libraries (consider using `@ts-ignore` or proper type definitions instead)
- During initial migration phases (enable progressively)

---

## Related Rules

- `@typescript-eslint/no-explicit-any` - Disallow the `any` type altogether
- `@typescript-eslint/no-unsafe-assignment` - Disallow assigning `any` to variables
- `@typescript-eslint/no-unsafe-return` - Disallow returning `any` from functions

## Resources

- [TypeScript-ESLint Documentation](https://typescript-eslint.io/rules/)
- [no-unsafe-argument](https://typescript-eslint.io/rules/no-unsafe-argument/)
- [no-unsafe-call](https://typescript-eslint.io/rules/no-unsafe-call/)
- [no-unsafe-member-access](https://typescript-eslint.io/rules/no-unsafe-member-access/)

---

**Implementation Status**: ✅ Complete

All three rules have been fully implemented with comprehensive test coverage ported from the upstream TypeScript-ESLint project.
