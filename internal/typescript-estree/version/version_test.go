package version

import (
	"testing"
)

func TestSetTypeScriptVersion(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		expectError bool
	}{
		{
			name:        "Valid version 5.4.0",
			version:     "5.4.0",
			expectError: false,
		},
		{
			name:        "Valid version 4.9.5",
			version:     "4.9.5",
			expectError: false,
		},
		{
			name:        "Valid version with patch",
			version:     "5.3.2",
			expectError: false,
		},
		{
			name:        "Beta version",
			version:     "5.5.0-beta",
			expectError: false,
		},
		{
			name:        "RC version",
			version:     "5.4.1-rc",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset state before each test
			ResetVersionCheck()

			err := SetTypeScriptVersion(tt.version)
			if (err != nil) != tt.expectError {
				t.Errorf("SetTypeScriptVersion() error = %v, expectError %v", err, tt.expectError)
			}

			if !tt.expectError {
				version := GetCurrentVersion()
				if version == "unknown" {
					t.Errorf("Expected version to be set, got %q", version)
				}
			}
		})
	}
}

func TestIsVersionAtLeast(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		checkVersion   string
		expectedResult bool
	}{
		{
			name:           "5.4.0 is at least 5.4",
			currentVersion: "5.4.0",
			checkVersion:   "5.4",
			expectedResult: true,
		},
		{
			name:           "5.4.0 is at least 5.3",
			currentVersion: "5.4.0",
			checkVersion:   "5.3",
			expectedResult: true,
		},
		{
			name:           "5.3.0 is not at least 5.4",
			currentVersion: "5.3.0",
			checkVersion:   "5.4",
			expectedResult: false,
		},
		{
			name:           "4.9.5 is at least 4.9",
			currentVersion: "4.9.5",
			checkVersion:   "4.9",
			expectedResult: true,
		},
		{
			name:           "4.9.5 is not at least 5.0",
			currentVersion: "4.9.5",
			checkVersion:   "5.0",
			expectedResult: false,
		},
		{
			name:           "5.4.1-rc is at least 5.4",
			currentVersion: "5.4.1-rc",
			checkVersion:   "5.4",
			expectedResult: true,
		},
		{
			name:           "5.4.0-beta is at least 5.4",
			currentVersion: "5.4.0-beta",
			checkVersion:   "5.4",
			expectedResult: true,
		},
		{
			name:           "5.3.0-beta is not at least 5.4",
			currentVersion: "5.3.0-beta",
			checkVersion:   "5.4",
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset state before each test
			ResetVersionCheck()

			err := SetTypeScriptVersion(tt.currentVersion)
			if err != nil {
				t.Fatalf("Failed to set TypeScript version: %v", err)
			}

			result := IsVersionAtLeast(tt.checkVersion)
			if result != tt.expectedResult {
				t.Errorf("IsVersionAtLeast(%q) = %v, want %v (current version: %s)",
					tt.checkVersion, result, tt.expectedResult, tt.currentVersion)
			}
		})
	}
}

func TestTypeScriptVersionIsAtLeast(t *testing.T) {
	// Reset state before test
	ResetVersionCheck()

	// Set a TypeScript version
	err := SetTypeScriptVersion("5.4.0")
	if err != nil {
		t.Fatalf("Failed to set TypeScript version: %v", err)
	}

	// Test that the map is populated correctly
	tests := []struct {
		version  string
		expected bool
	}{
		{"4.7", true},
		{"4.8", true},
		{"4.9", true},
		{"5.0", true},
		{"5.1", true},
		{"5.2", true},
		{"5.3", true},
		{"5.4", true},
		{"5.5", false},
		{"5.6", false},
		{"5.7", false},
	}

	for _, tt := range tests {
		t.Run("Version "+tt.version, func(t *testing.T) {
			result, exists := TypeScriptVersionIsAtLeast[tt.version]
			if !exists {
				t.Errorf("Version %q not found in TypeScriptVersionIsAtLeast map", tt.version)
			}
			if result != tt.expected {
				t.Errorf("TypeScriptVersionIsAtLeast[%q] = %v, want %v", tt.version, result, tt.expected)
			}
		})
	}
}

func TestGetCurrentVersion(t *testing.T) {
	// Reset state
	ResetVersionCheck()

	// Before setting version
	version := GetCurrentVersion()
	if version != "unknown" {
		t.Errorf("GetCurrentVersion() before setting = %q, want %q", version, "unknown")
	}

	// After setting version
	err := SetTypeScriptVersion("5.4.2")
	if err != nil {
		t.Fatalf("Failed to set TypeScript version: %v", err)
	}

	version = GetCurrentVersion()
	if version != "5.4.2" {
		t.Errorf("GetCurrentVersion() after setting = %q, want %q", version, "5.4.2")
	}
}

func TestGetSupportedVersions(t *testing.T) {
	versions := GetSupportedVersions()
	if len(versions) == 0 {
		t.Error("GetSupportedVersions() returned empty list")
	}

	// Check that known versions are in the list
	expectedVersions := []string{"4.7", "4.8", "4.9", "5.0", "5.1", "5.2", "5.3", "5.4"}
	for _, expected := range expectedVersions {
		found := false
		for _, v := range versions {
			if v == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected version %q not found in supported versions", expected)
		}
	}
}

func TestSemverCheck(t *testing.T) {
	tests := []struct {
		name           string
		minVersion     string
		currentVersion string
		expected       bool
	}{
		{
			name:           "Exact match",
			minVersion:     "5.4",
			currentVersion: "5.4.0",
			expected:       true,
		},
		{
			name:           "Higher patch version",
			minVersion:     "5.4",
			currentVersion: "5.4.5",
			expected:       true,
		},
		{
			name:           "Higher minor version",
			minVersion:     "5.3",
			currentVersion: "5.4.0",
			expected:       true,
		},
		{
			name:           "Lower version",
			minVersion:     "5.4",
			currentVersion: "5.3.0",
			expected:       false,
		},
		{
			name:           "RC version matching",
			minVersion:     "5.4",
			currentVersion: "5.4.1-rc",
			expected:       true,
		},
		{
			name:           "Beta version matching",
			minVersion:     "5.4",
			currentVersion: "5.4.0-beta",
			expected:       true,
		},
		{
			name:           "RC version not matching",
			minVersion:     "5.4",
			currentVersion: "5.3.1-rc",
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current, err := parseTypeScriptVersion(tt.currentVersion)
			if err != nil {
				t.Fatalf("Failed to parse current version: %v", err)
			}

			result, err := semverCheck(tt.minVersion, current)
			if err != nil {
				t.Fatalf("semverCheck failed: %v", err)
			}

			if result != tt.expected {
				t.Errorf("semverCheck(%q, %q) = %v, want %v",
					tt.minVersion, tt.currentVersion, result, tt.expected)
			}
		})
	}
}

func TestIsVersionSupported(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected bool
	}{
		{
			name:     "Supported version 5.4.0",
			version:  "5.4.0",
			expected: true,
		},
		{
			name:     "Supported version 4.9.5",
			version:  "4.9.5",
			expected: true,
		},
		{
			name:     "Supported version 5.3.2",
			version:  "5.3.2",
			expected: true,
		},
		{
			name:     "Unsupported version 3.9.0",
			version:  "3.9.0",
			expected: false,
		},
		{
			name:     "Future version 6.0.0",
			version:  "6.0.0",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := parseTypeScriptVersion(tt.version)
			if err != nil {
				t.Fatalf("Failed to parse version: %v", err)
			}

			result := isVersionSupported(version)
			if result != tt.expected {
				t.Errorf("isVersionSupported(%q) = %v, want %v", tt.version, result, tt.expected)
			}
		})
	}
}

func TestResetVersionCheck(t *testing.T) {
	// Set a version
	err := SetTypeScriptVersion("5.4.0")
	if err != nil {
		t.Fatalf("Failed to set TypeScript version: %v", err)
	}

	// Verify it's set
	if GetCurrentVersion() == "unknown" {
		t.Error("Version should be set before reset")
	}

	// Reset
	ResetVersionCheck()

	// Verify reset worked
	if GetCurrentVersion() != "unknown" {
		t.Errorf("Version should be unknown after reset, got %q", GetCurrentVersion())
	}

	if len(TypeScriptVersionIsAtLeast) != 0 {
		t.Error("TypeScriptVersionIsAtLeast map should be empty after reset")
	}
}

func TestConcurrentVersionCheck(t *testing.T) {
	// Reset state
	ResetVersionCheck()

	// This test ensures that the sync.Once works correctly
	// Multiple goroutines should be able to call SetTypeScriptVersion
	// but only the first one should actually set the version
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			_ = SetTypeScriptVersion("5.4.0")
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify version was set
	version := GetCurrentVersion()
	if version != "5.4.0" {
		t.Errorf("GetCurrentVersion() = %q, want %q", version, "5.4.0")
	}
}

func TestSetTypeScriptVersion_InvalidVersions(t *testing.T) {
	tests := []struct {
		name    string
		version string
	}{
		{"Empty string", ""},
		{"Invalid format", "invalid"},
		{"Letters in version", "x.y.z"},
		{"Special characters", "5.4.@"},
		{"Double dots", "5..4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ResetVersionCheck()
			err := SetTypeScriptVersion(tt.version)
			if err == nil {
				t.Errorf("Expected error for invalid version %q, got nil", tt.version)
			}
		})
	}
}

func TestSetTypeScriptVersion_ErrorPropagation(t *testing.T) {
	// Reset state
	ResetVersionCheck()

	// Test that errors are properly propagated
	err := SetTypeScriptVersion("not-a-version")
	if err == nil {
		t.Error("Expected error when setting invalid version, got nil")
	}

	// Verify the version was not set
	version := GetCurrentVersion()
	if version != "unknown" {
		t.Errorf("Expected version to be 'unknown' after error, got %q", version)
	}
}

func TestGetSupportedVersions_ReturnsCopy(t *testing.T) {
	// Get the supported versions
	versions1 := GetSupportedVersions()
	versions2 := GetSupportedVersions()

	// Verify they have the same content
	if len(versions1) != len(versions2) {
		t.Error("GetSupportedVersions() returned different length slices")
	}

	// Modify the first slice
	if len(versions1) > 0 {
		versions1[0] = "modified"
	}

	// Verify the second slice is unchanged
	versions3 := GetSupportedVersions()
	if len(versions3) > 0 && versions3[0] == "modified" {
		t.Error("GetSupportedVersions() returns a mutable reference instead of a copy")
	}
}
