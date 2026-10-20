package service

import (
	"context"
	"coordinator/domain/validate"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"coordinator/model"
	relaydto "coordinator/service/relay/dto"
)

type relayAccessStore interface {
	GetRelayAccess(ctx context.Context) (*model.RelayAccess, error)
	SaveRelayAccess(ctx context.Context, access *model.RelayAccess) error
	ClearRelayAccess(ctx context.Context) error
}

func (s *Service) GetRelayAccess(ctx context.Context) (*relaydto.RelayAccessResponse, error) {
	store, ok := s.repo.(relayAccessStore)
	if !ok {
		return &relaydto.RelayAccessResponse{Connected: false}, nil
	}
	access, err := store.GetRelayAccess(ctx)
	if err != nil {
		return nil, err
	}
	if access == nil || access.URL == "" {
		return &relaydto.RelayAccessResponse{Connected: false}, nil
	}
	return &relaydto.RelayAccessResponse{
		Connected: true,
		URL:       access.URL,
		Scope:     access.Scope,
	}, nil
}

func (s *Service) ConnectRelay(ctx context.Context, rawURL, rawToken string) (*relaydto.RelayAccessResponse, error) {
	store, ok := s.repo.(relayAccessStore)
	if !ok {
		return nil, &validate.Error{Msg: "relay pairing is not available on this store"}
	}
	normalized, err := normalizeRelayURL(rawURL)
	if err != nil {
		return nil, &validate.Error{Msg: err.Error()}
	}
	token := strings.TrimSpace(rawToken)
	if token == "" {
		return nil, &validate.Error{Msg: "relay access token is required"}
	}
	if err := pingRelayHealth(normalized, token); err != nil {
		return nil, &validate.Error{Msg: err.Error()}
	}
	access := &model.RelayAccess{
		URL:   normalized,
		Token: token,
		Scope: "read",
	}
	if err := store.SaveRelayAccess(ctx, access); err != nil {
		return nil, err
	}
	return &relaydto.RelayAccessResponse{
		Connected: true,
		URL:       access.URL,
		Scope:     access.Scope,
	}, nil
}

func (s *Service) DisconnectRelay(ctx context.Context) (*relaydto.RelayAccessResponse, error) {
	store, ok := s.repo.(relayAccessStore)
	if !ok {
		return &relaydto.RelayAccessResponse{Connected: false}, nil
	}
	if err := store.ClearRelayAccess(ctx); err != nil {
		return nil, err
	}
	return &relaydto.RelayAccessResponse{Connected: false}, nil
}

func (s *Service) GetRelaySources(ctx context.Context) (*relaydto.RelaySourcesResponse, error) {
	store, ok := s.repo.(relayAccessStore)
	if !ok {
		return &relaydto.RelaySourcesResponse{Sources: nil}, nil
	}
	access, err := store.GetRelayAccess(ctx)
	if err != nil {
		return nil, err
	}
	if access == nil || access.URL == "" {
		return &relaydto.RelaySourcesResponse{Sources: nil}, nil
	}
	body, err := relayGET(access, "/api/v1/sources")
	if err != nil {
		return nil, &validate.Error{Msg: err.Error()}
	}
	var out relaydto.RelaySourcesResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func relayGET(access *model.RelayAccess, path string) ([]byte, error) {
	base := strings.TrimRight(access.URL, "/")
	req, err := http.NewRequest(http.MethodGet, base+path, nil)
	if err != nil {
		return nil, err
	}
	if access.Token != "" {
		req.Header.Set("Authorization", "Bearer "+access.Token)
	}
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("relay unreachable: %v", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("relay rejected the access token")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("relay returned %d", resp.StatusCode)
	}
	return data, nil
}

func pingRelayHealth(base, token string) error {
	req, err := http.NewRequest(http.MethodGet, strings.TrimRight(base, "/")+"/api/v1/health", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("relay health check failed — is the service running?")
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("relay rejected the access token")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("relay health check failed — is the service running?")
	}
	var probe struct {
		Service string `json:"service"`
		Status  string `json:"status"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4096)).Decode(&probe); err != nil {
		return fmt.Errorf("relay health check failed — is the service running?")
	}
	if probe.Status == "ok" && (probe.Service == "relay" || probe.Service == "") {
		return nil
	}
	return fmt.Errorf("relay health check failed — is the service running?")
}

func normalizeRelayURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("relay address is required")
	}
	if !strings.Contains(raw, "://") {
		if strings.HasPrefix(raw, "127.0.0.1") || strings.HasPrefix(raw, "localhost") || strings.HasPrefix(raw, "[::1]") {
			raw = "http://" + raw
		} else {
			raw = "https://" + raw
		}
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid relay address")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("use http or https")
	}
	host := strings.ToLower(u.Hostname())
	if u.Scheme == "http" && !isLoopbackHost(host) {
		return "", fmt.Errorf("http is only allowed for localhost (use https for remote relay)")
	}
	if host == "" {
		return "", fmt.Errorf("invalid relay host")
	}
	u.Path = strings.TrimSuffix(u.Path, "/")
	u.RawQuery = ""
	u.Fragment = ""
	return strings.TrimRight(u.String(), "/"), nil
}

func isLoopbackHost(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
