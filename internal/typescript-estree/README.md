# TypeScript ESTree

A Go port of [@typescript-eslint/typescript-estree](https://github.com/typescript-eslint/typescript-eslint/tree/main/packages/typescript-estree), which parses TypeScript source code and produces an ESTree-compliant AST.

## Overview

This module provides functionality to parse TypeScript/JavaScript source code and convert it to the ESTree AST format, which is the standard format used by ESLint and other JavaScript tooling. It builds upon the [typescript-go](https://github.com/microsoft/typescript-go) project to leverage TypeScript's official compiler infrastructure.

## Project Structure

```
internal/typescript-estree/
├── parser/          # Main parsing logic - converts source code to AST
├── converter/       # AST conversion logic - transforms TypeScript AST to ESTree
├── types/           # Type definitions for ESTree nodes
├── utils/           # Utility functions for AST manipulation
├── version/         # TypeScript version detection and compatibility checking
├── testutils/       # Testing utilities and helpers
└── README.md        # This file
```

## Current Status

**⚠️ Infrastructure Setup Phase**

This module is currently in the infrastructure setup phase. The directory structure, build configuration, and testing infrastructure are in place, but the actual parser and converter implementations are pending.

### What's Ready

- ✅ Go module configuration (`go.mod`)
- ✅ Directory structure following TypeScript ESTree organization
- ✅ Basic type definitions for ESTree nodes
- ✅ Test infrastructure and example tests
- ✅ CI/CD integration
- ✅ Linting configuration
- ✅ Build scripts and Makefile targets

### What's Next

- ⏳ Port core parser functionality from TypeScript ESTree
- ⏳ Implement AST converter from TypeScript to ESTree format
- ⏳ Add comprehensive test coverage
- ⏳ Implement JSX parsing support
- ⏳ Add TypeScript-specific extensions to ESTree

## Building and Testing

### Prerequisites

- Go 1.21 or later
- Git (for submodules)

### Setup

Initialize the git submodules (required for typescript-go):

```bash
git submodule update --init --recursive
```

### Running Tests

Run all tests for the typescript-estree module:

```bash
make test-typescript-estree
```

Or using go directly:

```bash
go test ./internal/typescript-estree/...
```

### Running Tests with Coverage

```bash
make test-coverage-typescript-estree
```

This will generate a `coverage-typescript-estree.html` file that you can open in a browser.

### Linting

```bash
make lint-typescript-estree
```

Or using golangci-lint directly:

```bash
golangci-lint run ./internal/typescript-estree/...
```

### Formatting

```bash
make fmt
```

## Development Guidelines

### Code Style

- Follow standard Go conventions and idioms
- Use `golangci-lint` for linting (configuration in `.golangci.yml`)
- Run `go fmt` before committing
- Add tests for new functionality

### Testing

- Write tests for all new functionality
- Use table-driven tests where appropriate
- Leverage the `testutils` package for common test operations
- Ensure tests run with `go test -parallel`

### Documentation

- Document all exported types, functions, and methods
- Use clear, descriptive names
- Add examples where helpful

## Architecture

### Parser Package

The `parser` package is responsible for:
- Taking source code as input
- Configuring parsing options (source type, ECMAScript version, JSX support)
- Delegating to typescript-go for initial parsing
- Returning an ESTree-compliant Program node

### Converter Package

The `converter` package handles:
- Converting TypeScript AST nodes to ESTree format
- Preserving source location information
- Handling TypeScript-specific extensions
- Managing comments and tokens when requested

### Types Package

The `types` package defines:
- ESTree node interfaces and types
- Source location types
- Base node implementations
- All ESTree-compliant node structures

### Utils Package

The `utils` package provides:
- Helper functions for AST manipulation
- Position and location utilities
- Common operations on nodes

### Version Package

The `version` package handles:
- TypeScript version detection and parsing
- Semantic version comparison
- Compatibility checking and warnings
- Version-specific feature flags

### Testutils Package

The `testutils` package contains:
- Test helper functions
- Assertion utilities
- Test data generators

## Contributing

This module is part of the rslint project. See the main [CONTRIBUTING.md](../../CONTRIBUTING.md) for general contribution guidelines.

### Running CI Checks Locally

Before submitting a PR, ensure:

```bash
# Tests pass
make test-typescript-estree

# Linting passes
make lint-typescript-estree

# Code is formatted
make fmt
```

## References

- [TypeScript ESTree Documentation](https://github.com/typescript-eslint/typescript-eslint/tree/main/packages/typescript-estree)
- [ESTree Specification](https://github.com/estree/estree)
- [TypeScript-Go Project](https://github.com/microsoft/typescript-go)
- [Go Modules Documentation](https://go.dev/doc/modules/managing-dependencies)

## License

This project follows the same license as the main rslint project. See [LICENSE](../../LICENSE) for details.
