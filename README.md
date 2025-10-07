<a href="https://gitmoji.dev">
  <img
    src="https://img.shields.io/badge/gitmoji-%20😜%20😍-FFDD67.svg?style=flat-square"
    alt="Gitmoji"
  />
</a>

[![Coverage Status](https://coveralls.io/repos/github/mikevdberge/jarversion-cli/badge.svg?branch=main)](https://coveralls.io/github/mikevdberge/jarversion-cli?branch=main)

# Introduction 
This application was created to support Bigfix Inventory to determine the version number or md5 hash of a jar file

# Getting Started
Just download the application file for your operating system and run it, it does not have any dependencies

## Flags

| Flag | Description | Default |
| --- | --- | --- |
| <code class="text-nowrap">--json</code> | Show the version number in JSON format. |  |
| <code class="text-nowrap">--json-file</code> <code class="text-nowrap">...<code class="text-nowrap"> | Write the version information in JSON formated output file. |  |
| <code class="text-nowrap">--text-file</code> <code class="text-nowrap">...<code class="text-nowrap"> | Write the version information in a text file. |  |
| <code class="text-nowrap">--md5</code> | Show the MD5 hash of the file | |
| <code class="text-nowrap">--help</code> | Show context-sensitive help. |  |
| <code class="text-nowrap">--version</code> | Show application version. |  |

# Build and Test
Run the build_all.sh to build the jarversion program.
To test the program just run go test in the src directory

# SBOM
cyclonedx-gomod app -json -output dist/jarversion-cli.bom.json -packages -files -licenses -main cmd/jarversion