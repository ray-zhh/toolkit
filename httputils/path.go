package httputils

import (
	"net/http"
	"path/filepath"
)

func GetFileNameFromUrl(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Content-Disposition
	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		// TODO::
	}

	return filepath.Base(resp.Request.URL.Path), nil

}
