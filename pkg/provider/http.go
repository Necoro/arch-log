package provider

import (
	"fmt"
	"io"
	"net/http"
)

func Request(req *http.Request) (io.ReadCloser, error) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", req.URL, err)
	}

	if resp.StatusCode >= 300 {
		resp.Body.Close()

		return nil, fmt.Errorf("fetching %s: Server returned status %s", req.URL, resp.Status)
	}

	return resp.Body, nil
}

func Fetch(url string) (io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return Request(req)
}
