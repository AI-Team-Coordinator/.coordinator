package service

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"coordinator/model"
)

const (
	relayRequestTimeout     = 12 * time.Second
	relayConfigRequestTimeout = 90 * time.Second
)

func relayRequest(access *model.RelayAccess, method, path string, body []byte) ([]byte, int, error) {
	return relayRequestWithTimeout(access, method, path, body, relayRequestTimeout)
}

func relayConfigRequest(access *model.RelayAccess, method, path string, body []byte) ([]byte, int, error) {
	return relayRequestWithTimeout(access, method, path, body, relayConfigRequestTimeout)
}

func relayRequestWithTimeout(access *model.RelayAccess, method, path string, body []byte, timeout time.Duration) ([]byte, int, error) {
	if access == nil || access.URL == "" {
		return nil, 0, fmt.Errorf("relay is not connected")
	}
	base := strings.TrimRight(access.URL, "/")
	var bodyReader io.Reader
	if len(body) > 0 {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, base+path, bodyReader)
	if err != nil {
		return nil, 0, err
	}
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
		req.ContentLength = int64(len(body))
	}
	if access.Token != "" {
		req.Header.Set("Authorization", "Bearer "+access.Token)
	}
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("relay unreachable: %v", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return data, resp.StatusCode, nil
}
