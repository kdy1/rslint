# TypeScript Version Checking

This package provides TypeScript version detection and compatibility checking functionality, ported from [@typescript-eslint/typescript-estree](https://github.com/typescript-eslint/typescript-eslint/blob/main/packages/typescript-estree/src/version-check.ts).

## Overview

The version checking system ensures that the TypeScript ESTree parser is used with compatible TypeScript versions. It:

- Detects the current TypeScript version
- Compares against supported version ranges
- Issues warnings for unsupported versions
- Provides version feature flags for conditional functionality

## Usage

### Basic Version Checking

```go
import "github.com/web-infra-dev/rslint/internal/typescript-estree/version"

// Set the TypeScript version (typically called during parser initialization)
err := version.SetTypeScriptVersion("5.4.0")
if err != nil {
    // Handle invalid version
}

// Check if a specific version is available
if version.IsVersionAtLeast("5.4") {
    // Use 5.4+ features
}

// Get the current version
currentVersion := version.GetCurrentVersion()
fmt.Println("TypeScript version:", currentVersion)
```

### Using the Version Map

The package maintains a `TypeScriptVersionIsAtLeast` map for quick lookups:

```go
// Check from the pre-computed map
if version.TypeScriptVersionIsAtLeast["5.4"] {
    // TypeScript 5.4 features are available
}

if version.TypeScriptVersionIsAtLeast["5.0"] {
    // TypeScript 5.0 features are available
}
```

### Supported Versions

Get the list of explicitly supported versions:

```go
supportedVersions := version.GetSupportedVersions()
fmt.Println("Supported versions:", supportedVersions)
// Output: [4.7 4.8 4.9 5.0 5.1 5.2 5.3 5.4 5.5 5.6 5.7]
```

## Version Comparison Logic

The package uses semantic versioning with special handling for pre-release versions:

- **Stable versions**: Compared using standard semver (e.g., `5.4.0`, `5.4.5`)
- **RC versions**: Release candidates are checked (e.g., `5.4.1-rc` satisfies `>= 5.4`)
- **Beta versions**: Beta releases are checked (e.g., `5.4.0-beta` satisfies `>= 5.4`)

For a version check like `IsVersionAtLeast("5.4")`, the following constraint is used:
```
>= 5.4.0 || >= 5.4.1-rc || >= 5.4.0-beta
```

This ensures that stable, RC, and beta versions of the same minor version are all considered compatible.

## Warning System

When an unsupported TypeScript version is detected, the package automatically issues a warning:

```
WARNING: You are using TypeScript version 6.0.0 which is not explicitly supported.
The supported TypeScript versions are: [4.7 4.8 4.9 5.0 5.1 5.2 5.3 5.4 5.5 5.6 5.7]
Please consider upgrading to a supported version for the best experience.
```

Warnings are issued only once per program execution.

## Thread Safety

The version checking system is thread-safe and uses `sync.Once` to ensure initialization happens only once, even when called from multiple goroutines concurrently.

## Testing

The package includes comprehensive tests covering:

- Version setting and retrieval
- Version comparison logic
- Supported version detection
- Pre-release version handling
- Concurrent access
- Reset functionality (for testing)

Run tests with:

```bash
go test ./internal/typescript-estree/version/...
```

## Implementation Details

### Semantic Versioning

This package uses the [Masterminds/semver](https://github.com/Masterminds/semver) library for semantic version parsing and comparison, which provides:

- Full semver 2.0.0 compatibility
- Constraint checking
- Pre-release version support

### Version Detection Flow

1. **Initialization**: `SetTypeScriptVersion()` is called with the detected TypeScript version
2. **Parsing**: Version string is parsed into a semver object
3. **Comparison**: All supported versions are checked against the current version
4. **Caching**: Results are stored in `TypeScriptVersionIsAtLeast` map
5. **Warning**: If version is not in the supported list, a warning is issued

### Resetting (Testing Only)

For testing purposes, the package provides a `ResetVersionCheck()` function to clear the initialization state:

```go
version.ResetVersionCheck()
// Now you can call SetTypeScriptVersion() again
```

**Note**: This should only be used in tests, not in production code.

## Examples

### Conditional Feature Usage

```go
// Use TypeScript 5.4+ decorators if available
if version.IsVersionAtLeast("5.4") {
    // Parse with decorator metadata support
} else {
    // Use legacy decorator syntax
}

// Check for const type parameters (5.0+)
if version.IsVersionAtLeast("5.0") {
    // Handle const type parameters
}
```

### Integration with Parser

```go
import (
    "github.com/web-infra-dev/rslint/internal/typescript-estree/version"
    "github.com/web-infra-dev/rslint/internal/typescript-estree/parser"
)

func InitializeParser(tsVersion string) error {
    // Set the TypeScript version
    if err := version.SetTypeScriptVersion(tsVersion); err != nil {
        return fmt.Errorf("invalid TypeScript version: %w", err)
    }

    // Configure parser based on version capabilities
    parserConfig := parser.Config{
        SupportsDecorators: version.IsVersionAtLeast("5.0"),
        SupportsConstTypeParams: version.IsVersionAtLeast("5.0"),
        // ... other version-dependent features
    }

    return nil
}
```

## References

- [Original TypeScript ESTree version-check.ts](https://github.com/typescript-eslint/typescript-eslint/blob/main/packages/typescript-estree/src/version-check.ts)
- [TypeScript Release Notes](https://www.typescriptlang.org/docs/handbook/release-notes/overview.html)
- [Semantic Versioning 2.0.0](https://semver.org/)
- [Masterminds/semver Documentation](https://github.com/Masterminds/semver)

## Contributing

When adding support for new TypeScript versions:

1. Add the version to `SupportedVersions` slice in `version.go`
2. Add corresponding test cases in `version_test.go`
3. Document any version-specific features or breaking changes
4. Update this README with any new usage patterns

## License

This package is part of the rslint project and follows the same license. See the main [LICENSE](../../../LICENSE) file for details.
