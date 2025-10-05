package main

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
    "regexp" 	
	"strings"
	"testing"
)

func TestParseManifest(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected VersionInfo
	}{
		{
			name: "Both versions present",
			input: `
                Manifest-Version: 1.0
                Implementation-Version: 1.2.3
                Specification-Version: 4.5.6
            `,
			expected: VersionInfo{
				ImplementationVersion: "1.2.3",
				SpecificationVersion:  "4.5.6",
			},
		},
		{
			name: "Only Implementation-Version",
			input: `
                Implementation-Version: 2.0.0
            `,
			expected: VersionInfo{
				ImplementationVersion: "2.0.0",
			},
		},
		{
			name: "Only Specification-Version",
			input: `
                Specification-Version: 3.1.4
            `,
			expected: VersionInfo{
				SpecificationVersion: "3.1.4",
			},
		},
		{
			name: "No version info",
			input: `
                Manifest-Version: 1.0
            `,
			expected: VersionInfo{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseManifest(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %+v, got %+v", tt.expected, result)
			}
		})
	}
}

func createTestJar(t *testing.T, manifestContent string) string {
	t.Helper()

	tmpDir := t.TempDir()
	jarPath := filepath.Join(tmpDir, "sample.jar")

	jarFile, err := os.Create(jarPath)
	if err != nil {
		t.Fatalf("Failed to create JAR file: %v", err)
	}
	defer jarFile.Close()

	zipWriter := zip.NewWriter(jarFile)

	w, err := zipWriter.Create("META-INF/MANIFEST.MF")
	if err != nil {
		t.Fatalf("Failed to create MANIFEST.MF in JAR: %v", err)
	}
	_, err = w.Write([]byte(manifestContent))
	if err != nil {
		t.Fatalf("Failed to write MANIFEST.MF content: %v", err)
	}

	zipWriter.Close()
	return jarPath
}

func TestJarVersion_EmptyJar(t *testing.T) {
    tmpDir := t.TempDir()
    jarPath := filepath.Join(tmpDir, "empty.jar")

    jarFile, err := os.Create(jarPath)
    if err != nil {
        t.Fatalf("Failed to create empty JAR: %v", err)
    }
    defer jarFile.Close()

    zipWriter := zip.NewWriter(jarFile)
    zipWriter.Close()

    var out bytes.Buffer
    err = RunCLI([]string{jarPath}, &out)
if err == nil {
    t.Log("No error returned, checking output instead")
}
if !strings.Contains(out.String(), "MANIFEST.MF not found") {
    t.Errorf("Expected manifest error, got: %s", out.String())
}
}


func TestJarVersion_MissingFile(t *testing.T) {
    var out bytes.Buffer
    err := RunCLI([]string{"nonexistent.jar"}, &out)
	if err == nil && !strings.Contains(out.String(), "no such file or directory") {
    	t.Errorf("Expected file error, got: %s", out.String())
	}
}

func TestJarVersionIntegration(t *testing.T) {
	manifest := `
Manifest-Version: 1.0
Implementation-Version: 9.8.7
Specification-Version: 6.5.4
`
	jarPath := createTestJar(t, manifest)

	var out bytes.Buffer
	err := RunCLI([]string{jarPath}, &out)
	if err != nil {
		t.Fatalf("RunCLI failed: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Implementation-Version: 9.8.7") {
		t.Errorf("Expected Implementation-Version in output, got: %s", output)
	}
	if !strings.Contains(output, "Specification-Version: 6.5.4") {
		t.Errorf("Expected Specification-Version in output, got: %s", output)
	}
}

func TestParseManifest_IrregularFormatting(t *testing.T) {
    input := `
Manifest-Version: 1.0

Implementation-Version:     1.2.3

Specification-Version: 4.5.6
`
    expected := VersionInfo{
        ImplementationVersion: "1.2.3",
        SpecificationVersion:  "4.5.6",
    }
    result := ParseManifest(input)
    if result != expected {
        t.Errorf("Expected %+v, got %+v", expected, result)
    }
}

func TestRunCLI_HelpFlag(t *testing.T) {
	var out bytes.Buffer
	err := RunCLI([]string{"--help"}, &out)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !strings.Contains(out.String(), "Usage: jarversion") {
		t.Errorf("Expected help output, got: %s", out.String())
	}
}

func TestRunCLI_VersionFlag(t *testing.T) {
	var out bytes.Buffer
	err := RunCLI([]string{"--version"}, &out)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !strings.Contains(out.String(), toolVersion) {
		t.Errorf("Expected version output, got: %s", out.String())
	}
}

func TestRunCLI_MD5Flag(t *testing.T) {
    // Create a JAR with known content
    manifest := `
Manifest-Version: 1.0
Implementation-Version: 1.2.3
Specification-Version: 4.5.6
`
    jarPath := createTestJar(t, manifest)

    var out bytes.Buffer
    err := RunCLI([]string{"--md5", jarPath}, &out)
    if err != nil {
        t.Fatalf("RunCLI with --md5 failed: %v", err)
    }

	output := strings.TrimSpace(out.String())

    // Optional: validate hash format
	if !strings.HasPrefix(output, "MD5: ") {
		t.Fatalf("Expected 'MD5:' prefix, got: %s", output)
	}

	hash := strings.TrimSpace(strings.TrimPrefix(output, "MD5: "))
	if !regexp.MustCompile(`^[a-f0-9]{32}$`).MatchString(hash) {
		t.Errorf("Expected valid MD5 hash, got: %s", hash)
	}

    if len(hash) != 32 {
        t.Errorf("Expected 32-character MD5 hash, got: %s", hash)
    }
}

func TestRunCLI_InvalidFlag(t *testing.T) {
	var out bytes.Buffer
	err := RunCLI([]string{"--unknown"}, &out)
	if err == nil {
		t.Error("Expected error for unknown flag, got nil")
	}
	if !strings.Contains(out.String(), "flag provided but not defined") {
		t.Errorf("Expected error message, got: %s", out.String())
	}
}
