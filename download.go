package update

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	EnvGitHubToken = "GITHUB_TOKEN"
	EnvGHToken     = "GH_TOKEN"
)

// DefaultDownload was MOVED verbatim from webtyp/installer (mode_binary.go).
func DefaultDownload(url string) ([]byte, error) {
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	token := os.Getenv(EnvGitHubToken)
	if token == "" {
		token = os.Getenv(EnvGHToken)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}
