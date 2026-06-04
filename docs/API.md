# API Reference

- `VerifyChecksum(asset string, data, sums []byte) error`: Verifies data's SHA256 against the entry for asset in a checksums.txt body.
- `ResolveLatestVersion(source string, download func(string) ([]byte, error)) (string, error)`: Returns the latest release tag name from a GitHub repository URL.
- `IsOutdated(current, latest string) bool`: Reports whether the current version is strictly older than the latest semver version.
- `DefaultDownload(url string) ([]byte, error)`: Downloads content from a URL with GitHub-specific headers and authentication support.
- `Swap(targetPath, newFilePath string) (string, error)`: Backs up the target and moves the new file into its place; returns the backup path.
- `Rollback(targetPath, backupPath string) error`: Restores a backup file over the target path.
- `EnvGitHubToken`: Environment variable name for the primary GitHub token.
- `EnvGHToken`: Environment variable name for the fallback GitHub token.
