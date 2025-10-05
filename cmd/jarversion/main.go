// jarversion CLI Tool
//
// Description:
//
//	jarversion is a command-line tool designed to inject and manage versioning
//	information for Java JAR files. It supports automated version tagging,
//	artifact publishing, and integration with CI/CD pipelines.
//
// Usage:
//
//	The tool is typically built and executed as part of an Azure DevOps pipeline.
//	It reads the version number from pipeline parameters and injects it into
//	the source code before building binaries for multiple platforms.
//
// Build Process:
//   - The version number is injected into `jarversion.go` using `sed`.
//   - The `build_all.sh` script compiles the tool for supported platforms.
//   - Build artifacts are published under a release name format:
//     `jarversion-cli-build-YYYYMMDDrN`
//   (- A Git tag is created and pushed using the generated release name.)
//
// Pipeline Configuration:
//   - Manual trigger only (no automatic build on push or PR).
//   - Uses `ubuntu-latest` build agent.
//   - Artifacts are published to the Azure DevOps container.
//     (- Git tagging uses OAuth token authentication.)
//
// Pipeline parameters:
//   - version: The semantic version to inject (default: 1.0.0)
//
// Author:
//
//	Maintained by Mike van den Berge
//
// License:
//
//	[Insert license info here if applicable]
package main

import (
    "archive/zip"
    "crypto/md5"
    "encoding/json"
    "flag"
    "fmt"
    "io"
    "log"
    "os"
    "strings"
)

const toolVersion = "latest"

type VersionInfo struct {
    ImplementationVersion string `json:"implementation_version,omitempty"`
    SpecificationVersion  string `json:"specification_version,omitempty"`
    MD5                   string `json:"md5,omitempty"`
}

func printHelp(w io.Writer) {
    fmt.Fprintln(w, `Usage: jarversion [options] <path-to-jar-file>

Options:
  --json             Output version info in JSON format
  --json-file <file> Write JSON output to specified file
  --text-file <file> Write version info to specified text file
  --md5              Output MD5 hash of the JAR file
  --version          Show tool version
  --help             Show this help message`)
}

// ParseManifest parses the contents of a MANIFEST.MF file and extracts 
// version-related metadata.
// It scans each line of the manifest string for known version keys:
// 	- "Implementation-Version"
// - "Specification-Version"
//
// Matching lines are trimmed and stored in a VersionInfo struct.
//
// Parameters:
// 	manifest string - Raw content of the MANIFEST.MF file.
// 
// Returns:
// 	VersionInfo - A struct containing the extracted version fields.
// 
// Notes:
//  - Lines are matched using prefix checks and trimmed for whitespace.
// 	- Unmatched lines are ignored. // 
// 	- This function assumes a simple flat manifest format without continuation lines.
//
// Example usage:
//  info := ParseManifest(manifestContent) 
//  fmt.Println(info.ImplementationVersion)
func ParseManifest(manifest string) VersionInfo {
    var version VersionInfo
    for _, line := range strings.Split(manifest, "\n") {
        line = strings.TrimSpace(line)
        if strings.HasPrefix(line, "Implementation-Version:") {
            version.ImplementationVersion = strings.TrimSpace(strings.TrimPrefix(line, "Implementation-Version:"))
        }
        if strings.HasPrefix(line, "Specification-Version:") {
            version.SpecificationVersion = strings.TrimSpace(strings.TrimPrefix(line, "Specification-Version:"))
        }
    }
    return version
}

func RunCLI(args []string, stdout io.Writer) error {
    flag.CommandLine = flag.NewFlagSet("jarversion", flag.ContinueOnError)
    flag.CommandLine.SetOutput(stdout)

    jsonOutput := flag.Bool("json", false, "Output version info in JSON format")
    jsonFile := flag.String("json-file", "", "Write JSON output to specified file")
    textFile := flag.String("text-file", "", "Write version info to specified text file")
    md5Output := flag.Bool("md5", false, "Output MD5 hash of the JAR file")
    showVersion := flag.Bool("version", false, "Show tool version")
    showHelp := flag.Bool("help", false, "Show help message")

    err := flag.CommandLine.Parse(args)
    if err != nil {
        return err
    }

    if *showHelp {
        printHelp(stdout)
        return nil
    }

    if *showVersion {
        fmt.Fprintln(stdout, "jarversion CLI tool version:", toolVersion)
        return nil
    }

    if flag.CommandLine.NArg() < 1 {
        fmt.Fprintln(stdout, "jarversion - Jar version CLI to query the version information in the MANIFEST.MF file.")
        printHelp(stdout)
        return nil
    }

    jarPath := flag.CommandLine.Arg(0)

    // If only --md5 is set, output hash and exit
    if *md5Output && !*jsonOutput && *jsonFile == "" && *textFile == "" {
        file, err := os.Open(jarPath)
        if err != nil {
            return fmt.Errorf("failed to open JAR file for hashing: %w", err)
        }
        defer file.Close()

        hash := md5.New()
        if _, err := io.Copy(hash, file); err != nil {
            return fmt.Errorf("failed to compute MD5 hash: %w", err)
        }
        fmt.Fprintf(stdout, "MD5: %x\n", hash.Sum(nil))
        return nil
    }

    r, err := zip.OpenReader(jarPath)
    if err != nil {
        return fmt.Errorf("failed to open JAR file: %w", err)
    }
    defer r.Close()

    for _, f := range r.File {
        if strings.EqualFold(f.Name, "META-INF/MANIFEST.MF") {
            rc, err := f.Open()
            if err != nil {
                return fmt.Errorf("failed to open MANIFEST.MF: %w", err)
            }
            defer rc.Close()

            data, err := io.ReadAll(rc)
            if err != nil {
                return fmt.Errorf("failed to read MANIFEST.MF: %w", err)
            }

            version := ParseManifest(string(data))

            // Include MD5 if requested alongside other output
            if *md5Output {
                file, err := os.Open(jarPath)
                if err != nil {
                    return fmt.Errorf("failed to open JAR file for hashing: %w", err)
                }
                defer file.Close()

                hash := md5.New()
                if _, err := io.Copy(hash, file); err != nil {
                    return fmt.Errorf("failed to compute MD5 hash: %w", err)
                }
                version.MD5 = fmt.Sprintf("%x", hash.Sum(nil))
            }

            if *jsonOutput || *jsonFile != "" {
                jsonBytes, err := json.MarshalIndent(version, "", "  ")
                if err != nil {
                    return fmt.Errorf("failed to encode JSON: %w", err)
                }

                if *jsonFile != "" {
                    err := os.WriteFile(*jsonFile, jsonBytes, 0644)
                    if err != nil {
                        return fmt.Errorf("failed to write JSON to file: %w", err)
                    }
                    fmt.Fprintf(stdout, "✅ JSON written to %s\n", *jsonFile)
                } else {
                    fmt.Fprintln(stdout, string(jsonBytes))
                }
            } else if *textFile != "" {
                var lines []string
                if version.ImplementationVersion != "" {
                    lines = append(lines, "Implementation-Version: "+version.ImplementationVersion)
                }
                if version.SpecificationVersion != "" {
                    lines = append(lines, "Specification-Version: "+version.SpecificationVersion)
                }
                if version.MD5 != "" {
                    lines = append(lines, "MD5: "+version.MD5)
                }
                err := os.WriteFile(*textFile, []byte(strings.Join(lines, "\n")), 0644)
                if err != nil {
                    return fmt.Errorf("failed to write text to file: %w", err)
                }
                fmt.Fprintf(stdout, "✅ Version info written to %s\n", *textFile)
            } else {
                if version.ImplementationVersion != "" {
                    fmt.Fprintln(stdout, "Implementation-Version:", version.ImplementationVersion)
                }
                if version.SpecificationVersion != "" {
                    fmt.Fprintln(stdout, "Specification-Version:", version.SpecificationVersion)
                }
                if version.MD5 != "" {
                    fmt.Fprintln(stdout, "MD5: ", version.MD5)
                }
            }
            return nil
        }
    }

    fmt.Fprintln(stdout, "MANIFEST.MF not found in JAR file.",err)
    return nil
}

func main() {
    err := RunCLI(os.Args[1:], os.Stdout)
    if err != nil {
        log.Fatal(err)
    }
}