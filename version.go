package update

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ResolveLatestVersion was MOVED from tinywasm/installer (mode_binary.go:
// resolveLatestVersion). The only change is decoupling: the *installer.Deps
// parameter was replaced by an injected download func so the proven logic lives
// here without depending on the installer package.
func ResolveLatestVersion(source string, download func(string) ([]byte, error)) (string, error) {
	// Source is expected to be https://github.com/owner/repo
	parts := strings.Split(strings.TrimSuffix(source, "/"), "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid source URL: %s", source)
	}
	owner := parts[len(parts)-2]
	repo := parts[len(parts)-1]

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	data, err := download(apiURL)
	if err != nil {
		return "", err
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.Unmarshal(data, &release); err != nil {
		return "", fmt.Errorf("failed to parse GitHub API response: %w", err)
	}

	if release.TagName == "" {
		return "", fmt.Errorf("no tag_name found in GitHub API response")
	}

	return release.TagName, nil
}
