package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"coordinator/domain/validate"
	"coordinator/infra"
	relaydto "coordinator/service/relay/dto"
)

func (s *Service) GetRelayConnectorActivity(ctx context.Context, connectorID string, limit, offset int) (*relaydto.RelayConnectorActivityResponse, error) {
	store, ok := s.repo.(relayAccessStore)
	if !ok {
		return &relaydto.RelayConnectorActivityResponse{Entries: nil, Total: 0, Limit: limit, Offset: offset}, nil
	}
	access, err := store.GetRelayAccess(ctx)
	if err != nil {
		return nil, err
	}
	if access == nil || access.URL == "" {
		return &relaydto.RelayConnectorActivityResponse{Entries: nil, Total: 0, Limit: limit, Offset: offset}, nil
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	path := fmt.Sprintf("/api/v1/connectors/activity?limit=%d&offset=%d", limit, offset)
	if strings.TrimSpace(connectorID) != "" {
		path += "&connector_id=" + strings.TrimSpace(connectorID)
	}
	data, code, err := relayRequest(access, http.MethodGet, path, nil)
	if err != nil {
		return nil, &validate.Error{Msg: err.Error()}
	}
	if code == http.StatusUnauthorized {
		return nil, &validate.Error{Msg: "relay rejected the access token"}
	}
	if code != http.StatusOK {
		return nil, &validate.Error{Msg: fmt.Sprintf("relay returned %d", code)}
	}
	var out relaydto.RelayConnectorActivityResponse
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *Service) GetRelayMarketplace(ctx context.Context) (*relaydto.RelayMarketplaceResponse, error) {
	store, ok := s.repo.(relayAccessStore)
	if !ok {
		return &relaydto.RelayMarketplaceResponse{Connectors: nil}, nil
	}
	access, err := store.GetRelayAccess(ctx)
	if err != nil {
		return nil, err
	}
	if access == nil || access.URL == "" {
		return &relaydto.RelayMarketplaceResponse{Connectors: nil}, nil
	}
	data, code, err := relayRequest(access, http.MethodGet, "/api/v1/connectors/marketplace", nil)
	if err != nil {
		return nil, &validate.Error{Msg: err.Error()}
	}
	if code == http.StatusUnauthorized {
		return nil, &validate.Error{Msg: "relay rejected the access token"}
	}
	if code != http.StatusOK {
		return nil, &validate.Error{Msg: fmt.Sprintf("relay returned %d", code)}
	}
	var out relaydto.RelayMarketplaceResponse
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *Service) GetRelayConnectorConfig(ctx context.Context, id string) (*relaydto.RelayConnectorConfigResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, &validate.Error{Msg: "connector id is required"}
	}
	store, ok := s.repo.(relayAccessStore)
	if !ok {
		return nil, &validate.Error{Msg: "relay is not connected"}
	}
	access, err := store.GetRelayAccess(ctx)
	if err != nil {
		return nil, err
	}
	if access == nil || access.URL == "" {
		return nil, &validate.Error{Msg: "relay is not connected"}
	}
	data, code, err := relayConfigRequest(access, http.MethodGet, "/api/v1/connectors/"+id+"/config", nil)
	if err != nil {
		return nil, &validate.Error{Msg: err.Error()}
	}
	if code != http.StatusOK {
		return nil, &validate.Error{Msg: fmt.Sprintf("relay returned %d", code)}
	}
	var out relaydto.RelayConnectorConfigResponse
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *Service) PutRelayConnectorConfig(ctx context.Context, id string, req relaydto.RelayPutConnectorConfigRequest) (*relaydto.RelayConnectorConfigResponse, error) {
	if _, err := s.requireRelayAdmin(ctx); err != nil {
		return nil, err
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, &validate.Error{Msg: "connector id is required"}
	}
	store, ok := s.repo.(relayAccessStore)
	if !ok {
		return nil, &validate.Error{Msg: "relay pairing is not available"}
	}
	access, err := store.GetRelayAccess(ctx)
	if err != nil {
		return nil, err
	}
	if access == nil || access.URL == "" {
		return nil, &validate.Error{Msg: "relay is not connected"}
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	start := time.Now()
	data, code, err := relayConfigRequest(access, http.MethodPut, "/api/v1/connectors/"+id+"/config", body)
	if err != nil {
		return nil, &validate.Error{Msg: err.Error()}
	}
	if d := time.Since(start); d > 5*time.Second {
		infra.LogInfo("relay PUT config %s took %s", id, d.Round(time.Millisecond))
	}
	if code == http.StatusForbidden {
		return nil, &validate.Error{Msg: "relay config requires the master bearer token — reconnect with the token from .relay_bearer_token"}
	}
	if code != http.StatusOK {
		return nil, &validate.Error{Msg: fmt.Sprintf("relay returned %d", code)}
	}
	var out relaydto.RelayConnectorConfigResponse
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *Service) TriggerRelayConnectorAction(ctx context.Context, id, action string) error {
	if _, err := s.requireRelayAdmin(ctx); err != nil {
		return err
	}
	id = strings.TrimSpace(id)
	action = strings.TrimSpace(action)
	if id == "" {
		return &validate.Error{Msg: "connector id is required"}
	}
	if action == "" {
		return &validate.Error{Msg: "action is required"}
	}
	store, ok := s.repo.(relayAccessStore)
	if !ok {
		return &validate.Error{Msg: "relay pairing is not available"}
	}
	access, err := store.GetRelayAccess(ctx)
	if err != nil {
		return err
	}
	if access == nil || access.URL == "" {
		return &validate.Error{Msg: "relay is not connected"}
	}
	path := "/api/v1/connectors/" + id + "/actions/" + action
	data, code, err := relayRequestWithTimeout(access, http.MethodPost, path, nil, 130*time.Second)
	if err != nil {
		return &validate.Error{Msg: err.Error()}
	}
	if code == http.StatusForbidden {
		return &validate.Error{Msg: "relay action requires the master bearer token — reconnect with the token from .relay_bearer_token"}
	}
	if code == http.StatusNotImplemented {
		return &validate.Error{Msg: "this connector does not support that action"}
	}
	if code != http.StatusOK {
		msg := strings.TrimSpace(string(data))
		if msg == "" {
			msg = fmt.Sprintf("relay returned %d", code)
		}
		return &validate.Error{Msg: msg}
	}
	return nil
}

func (s *Service) DisableRelayConnector(ctx context.Context, id string, clearSecrets bool) (*relaydto.RelayConnectorConfigResponse, error) {
	if _, err := s.requireRelayAdmin(ctx); err != nil {
		return nil, err
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, &validate.Error{Msg: "connector id is required"}
	}
	store, ok := s.repo.(relayAccessStore)
	if !ok {
		return nil, &validate.Error{Msg: "relay pairing is not available"}
	}
	access, err := store.GetRelayAccess(ctx)
	if err != nil {
		return nil, err
	}
	if access == nil || access.URL == "" {
		return nil, &validate.Error{Msg: "relay is not connected"}
	}
	body, _ := json.Marshal(map[string]bool{"clear_secrets": clearSecrets})
	data, code, err := relayConfigRequest(access, http.MethodPost, "/api/v1/connectors/"+id+"/disable", body)
	if err != nil {
		return nil, &validate.Error{Msg: err.Error()}
	}
	if code == http.StatusForbidden {
		return nil, &validate.Error{Msg: "relay config requires the master bearer token — reconnect with the token from .relay_bearer_token"}
	}
	if code != http.StatusOK {
		return nil, &validate.Error{Msg: fmt.Sprintf("relay returned %d", code)}
	}
	var out relaydto.RelayConnectorConfigResponse
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
