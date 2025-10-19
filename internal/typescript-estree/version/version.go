// Package version provides TypeScript version checking and compatibility detection
package version

import (
	"fmt"
	"sync"

	"github.com/Masterminds/semver/v3"
)

// SupportedVersions lists all TypeScript versions that are explicitly supported
var SupportedVersions = []string{
	"4.7",
	"4.8",
	"4.9",
	"5.0",
	"5.1",
	"5.2",
	"5.3",
	"5.4",
	"5.5",
	"5.6",
	"5.7",
}

var (
	// TypeScriptVersionIsAtLeast is a map from version strings to booleans indicating
	// whether the current TypeScript version is at least that version
	TypeScriptVersionIsAtLeast map[string]bool

	// currentVersion holds the detected TypeScript version
	currentVersion *semver.Version

	// versionCheckOnce ensures version checking only happens once
	versionCheckOnce sync.Once

	// versionWarningIssued tracks whether a version warning has been issued
	versionWarningIssued bool
)

// init initializes the version checking system
func init() {
	TypeScriptVersionIsAtLeast = make(map[string]bool)
}

// SetTypeScriptVersion sets the TypeScript version to use for version checking
// This should be called during parser initialization with the detected TypeScript version
func SetTypeScriptVersion(version string) error {
	versionCheckOnce.Do(func() {
		var err error
		currentVersion, err = parseTypeScriptVersion(version)
		if err != nil {
			return
		}

		// Check each supported version
		for _, supportedVer := range SupportedVersions {
			satisfied, checkErr := semverCheck(supportedVer, currentVersion)
			if checkErr != nil {
				continue
			}
			TypeScriptVersionIsAtLeast[supportedVer] = satisfied
		}

		// Issue warning if version is not explicitly supported
		if !isVersionSupported(currentVersion) && !versionWarningIssued {
			issueVersionWarning(currentVersion)
			versionWarningIssued = true
		}
	})

	return nil
}

// GetCurrentVersion returns the currently detected TypeScript version
func GetCurrentVersion() string {
	if currentVersion == nil {
		return "unknown"
	}
	return currentVersion.String()
}

// IsVersionAtLeast checks if the current TypeScript version is at least the specified version
func IsVersionAtLeast(version string) bool {
	if result, exists := TypeScriptVersionIsAtLeast[version]; exists {
		return result
	}

	// If not in our cache, perform a dynamic check
	if currentVersion == nil {
		return false
	}

	satisfied, err := semverCheck(version, currentVersion)
	if err != nil {
		return false
	}

	return satisfied
}

// parseTypeScriptVersion parses a TypeScript version string into a semver.Version
func parseTypeScriptVersion(version string) (*semver.Version, error) {
	v, err := semver.NewVersion(version)
	if err != nil {
		return nil, fmt.Errorf("invalid TypeScript version %q: %w", version, err)
	}
	return v, nil
}

// semverCheck checks if the current version satisfies the minimum version requirement
// It accepts versions in the form "X.Y" and checks:
// - >= X.Y.0
// - >= X.Y.1-rc
// - >= X.Y.0-beta
func semverCheck(minVersion string, current *semver.Version) (bool, error) {
	// Create a constraint that matches:
	// - X.Y.0 and above
	// - X.Y.1-rc and above (for release candidates)
	// - X.Y.0-beta and above (for beta releases)
	constraintStr := fmt.Sprintf(">= %s.0 || >= %s.1-rc || >= %s.0-beta", minVersion, minVersion, minVersion)

	constraint, err := semver.NewConstraint(constraintStr)
	if err != nil {
		return false, fmt.Errorf("invalid version constraint: %w", err)
	}

	return constraint.Check(current), nil
}

// isVersionSupported checks if a version is in our explicitly supported versions list
func isVersionSupported(version *semver.Version) bool {
	for _, supported := range SupportedVersions {
		// Check if version starts with the supported version string
		constraint, err := semver.NewConstraint(fmt.Sprintf("^%s.0", supported))
		if err != nil {
			continue
		}
		if constraint.Check(version) {
			return true
		}
	}
	return false
}

// issueVersionWarning logs a warning about using an unsupported TypeScript version
func issueVersionWarning(version *semver.Version) {
	// In a real implementation, this would use a proper logging system
	// For now, we'll just store the warning state
	// The warning message will be issued by the caller if needed
	fmt.Printf("WARNING: You are using TypeScript version %s which is not explicitly supported.\n", version.String())
	fmt.Printf("The supported TypeScript versions are: %v\n", SupportedVersions)
	fmt.Printf("Please consider upgrading to a supported version for the best experience.\n")
}

// GetSupportedVersions returns the list of explicitly supported TypeScript versions
func GetSupportedVersions() []string {
	return SupportedVersions
}

// ResetVersionCheck resets the version checking state (primarily for testing)
func ResetVersionCheck() {
	versionCheckOnce = sync.Once{}
	TypeScriptVersionIsAtLeast = make(map[string]bool)
	currentVersion = nil
	versionWarningIssued = false
}
