# TypeScript ESTree Infrastructure Setup

This document describes the infrastructure setup for the TypeScript ESTree port in rslint.

## Overview

This PR sets up the foundational infrastructure needed for porting [@typescript-eslint/typescript-estree](https://github.com/typescript-eslint/typescript-eslint/tree/main/packages/typescript-estree) to Go as part of the rslint project.

**⚠️ Important**: This PR contains **only infrastructure and scaffolding** - no actual parser implementation. The parser functionality will be added in subsequent PRs.

## What's Included

### 1. Module Structure

Created a new Go module at `internal/typescript-estree/` with the following structure:

```
internal/typescript-estree/
├── parser/          # Main parsing logic (placeholder)
├── converter/       # AST conversion logic (placeholder)
├── types/           # Type definitions for ESTree nodes
├── utils/           # Utility functions
├── testutils/       # Testing utilities
├── go.mod           # Module dependencies
└── README.md        # Module documentation
```

### 2. Build Configuration

- **Go Module**: Created `internal/typescript-estree/go.mod` with dependencies on typescript-go shim packages
- **Workspace**: Updated `go.work` to include the new module
- **Makefile**: Added targets for building and testing the module:
  - `make test-typescript-estree` - Run tests
  - `make test-coverage-typescript-estree` - Run tests with coverage
  - `make lint-typescript-estree` - Run linters

### 3. CI/CD Integration

The existing CI workflows in `.github/workflows/ci.yml` already cover:
- Running tests for all packages under `internal/` (line 57)
- Running golangci-lint on `internal/...` (line 80)
- Code formatting checks

The new module will be automatically included in these checks.

### 4. Type Definitions

Basic ESTree type definitions in `types/types.go`:
- `Node` interface - base interface for all AST nodes
- `SourceLocation` and `Position` - location tracking
- `BaseNode` - common node implementation
- `Program` and `Identifier` - example concrete nodes

These are minimal scaffolding types that will be expanded during implementation.

### 5. Test Infrastructure

- Test files for each package (`*_test.go`)
- `testutils` package with helper functions
- All tests pass and can run in parallel
- Example tests demonstrate the testing approach

### 6. Documentation

- Module README at `internal/typescript-estree/README.md`
- This setup document
- Inline code documentation

## Verification Steps

All infrastructure is verified to work:

1. **Tests Run Successfully**:
   ```bash
   go test ./internal/typescript-estree/...
   # All tests pass
   ```

2. **Module Dependencies Resolve**:
   ```bash
   cd internal/typescript-estree && go mod tidy
   # No errors
   ```

3. **Workspace Configuration Valid**:
   ```bash
   go work sync
   # Module included in workspace
   ```

## Next Steps

With infrastructure in place, the next phases are:

1. **Parser Implementation**:
   - Port core parsing logic from TypeScript ESTree
   - Integrate with typescript-go shim
   - Handle source type and ECMAScript version options

2. **Converter Implementation**:
   - Convert TypeScript AST to ESTree format
   - Handle TypeScript-specific extensions
   - Preserve location information

3. **Testing**:
   - Add comprehensive test cases
   - Port test fixtures from TypeScript ESTree
   - Achieve high code coverage

4. **JSX Support**:
   - Add JSX parsing capabilities
   - Handle JSX-specific node types

## Development Workflow

To work on the typescript-estree module:

```bash
# Run tests
make test-typescript-estree

# Run with coverage
make test-coverage-typescript-estree

# Run linting
make lint-typescript-estree

# Format code
make fmt
```

## References

- [TypeScript ESTree](https://github.com/typescript-eslint/typescript-eslint/tree/main/packages/typescript-estree)
- [ESTree Spec](https://github.com/estree/estree)
- [TypeScript-Go](https://github.com/microsoft/typescript-go)
- [Main README](internal/typescript-estree/README.md)
