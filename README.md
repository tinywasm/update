# update

Reusable self-update for Go CLIs: latest-version resolution, SHA256 checksum verification, and atomic binary replacement with rollback.

## Features

- **Latest Version Resolution**: Find the latest release tag from a GitHub repository.
- **Checksum Verification**: Verify downloaded assets against a `checksums.txt` file.
- **Atomic Binary Replacement**: Swap the current binary with a new one, with automatic backup and rollback support.
- **Cross-Filesystem Support**: Robustly move files even across different mount points.
- **GitHub API Integration**: Uses recommended headers and supports authentication via `GITHUB_TOKEN` or `GH_TOKEN` to avoid rate limits.

## API Documentation

See [docs/API.md](docs/API.md) for a concise list of exported symbols and their contracts.

## Usage Examples

### 1. Resolve, Download, and Verify

```go
package main

import (
    "fmt"
    "github.com/tinywasm/update"
)

func main() {
    source := "https://github.com/tinywasm/tinywasm"

    // 1. Resolve latest version
    latest, err := update.ResolveLatestVersion(source, update.DefaultDownload)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Latest version: %s\n", latest)

    // 2. Download asset and checksums
    assetURL := fmt.Sprintf("%s/releases/download/%s/tinywasm-linux-amd64", source, latest)
    sumsURL := fmt.Sprintf("%s/releases/download/%s/checksums.txt", source, latest)

    data, err := update.DefaultDownload(assetURL)
    if err != nil {
        panic(err)
    }

    sums, err := update.DefaultDownload(sumsURL)
    if err != nil {
        panic(err)
    }

    // 3. Verify checksum
    if err := update.VerifyChecksum("tinywasm-linux-amd64", data, sums); err != nil {
        panic(err)
    }
    fmt.Println("Checksum verified!")
}
```

### 2. Check if Outdated

```go
current := "v0.1.0"
latest := "v0.2.0"

if update.IsOutdated(current, latest) {
    fmt.Println("A new version is available!")
}
```

### 3. Atomic Swap with Rollback

```go
target := "/usr/local/bin/mytool"
newBinary := "/tmp/mytool-new"

backup, err := update.Swap(target, newBinary)
if err != nil {
    panic(err)
}

// Perform health check on the new binary...
success := runHealthCheck(target)
if !success {
    fmt.Println("Health check failed, rolling back...")
    if err := update.Rollback(target, backup); err != nil {
        panic(err)
    }
}
```

## Environment Variables

- `GITHUB_TOKEN`: Optional GitHub personal access token.
- `GH_TOKEN`: Fallback GitHub personal access token.
