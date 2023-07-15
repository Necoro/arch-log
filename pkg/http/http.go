package http

import (
	"fmt"
	"io"
	"net/http"

	"github.com/Necoro/arch-log/pkg/entries"
)

func Fetch(url string) (io.ReadCloser, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", url, err)
	}

	if resp.StatusCode >= 300 {
		resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			return nil, entries.ErrNotFound
		}
		return nil, fmt.Errorf("fetching %s: Server returned status %s", url, resp.Status)
	}

	return resp.Body, nil
}
