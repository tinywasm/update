package update

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// ResolveLatestVersion was MOVED from webtyp/installer (mode_binary.go:
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

// IsOutdated reports whether current is strictly older than latest (MAJOR.MINOR.PATCH,
// leading "v" optional, pre-release/build suffix ignored). Any parse failure -> false,
// so "dev"/empty builds are never reported as outdated.
func IsOutdated(current, latest string) bool {
	c, ok1 := parseSemver(current)
	l, ok2 := parseSemver(latest)
	if !ok1 || !ok2 {
		return false
	}
	for i := 0; i < 3; i++ {
		if c[i] != l[i] {
			return c[i] < l[i]
		}
	}
	return false
}

func parseSemver(v string) ([3]int, bool) {
	var out [3]int
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	if i := strings.IndexAny(v, "-+"); i >= 0 {
		v = v[:i]
	}
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return out, false
	}
	for i := 0; i < 3; i++ {
		n, err := strconv.Atoi(parts[i])
		if err != nil {
			return out, false
		}
		out[i] = n
	}
	return out, true
}
