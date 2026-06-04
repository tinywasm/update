package update

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// DefaultDownload was MOVED verbatim from tinywasm/installer (mode_binary.go).
func DefaultDownload(url string) ([]byte, error) {
	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}
