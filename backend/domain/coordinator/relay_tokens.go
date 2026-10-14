package coordinator

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strings"
	"time"

	"coordinator/domain/coordinator/dto"
	"coordinator/model"

	"github.com/google/uuid"
)

type relayIssuedTokenStore interface {
	GetRelayIssuedTokens(ctx context.Context) (*model.RelayIssuedTokenFile, error)
	SaveRelayIssuedTokens(ctx context.Context, file *model.RelayIssuedTokenFile) error
}

func (s *Service) ListRelayIssuedTokens(ctx context.Context) (*dto.RelayIssuedTokenListResponse, error) {
	if _, err := s.requireRelayAdmin(ctx); err != nil {
		return nil, err
	}
	store, ok := s.repo.(relayIssuedTokenStore)
	if !ok {
		return &dto.RelayIssuedTokenListResponse{Tokens: nil}, nil
	}
	file, err := store.GetRelayIssuedTokens(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]dto.RelayIssuedTokenResponse, 0, len(file.Tokens))
	for _, t := range file.Tokens {
		out = append(out, mapIssuedToken(t))
	}
	return &dto.RelayIssuedTokenListResponse{Tokens: out}, nil
}

func (s *Service) IssueRelayToken(ctx context.Context, req dto.RelayIssueTokenRequest) (*dto.RelayIssueTokenResponse, error) {
	adminAlias, err := s.requireRelayAdmin(ctx)
	if err != nil {
		return nil, err
	}
	store, ok := s.repo.(relayIssuedTokenStore)
	if !ok {
		return nil, &ValidationError{Msg: "relay token store is not available"}
	}
	alias := strings.ToUpper(strings.TrimSpace(req.Alias))
	if alias == "" {
		alias = adminAlias
	}
	team, err := s.repo.GetTeam(ctx)
	if err != nil {
		return nil, err
	}
	if !teamHasAlias(team, alias) {
		return nil, &ValidationError{Msg: "unknown team member alias"}
	}
	secret, err := newRelayClientSecret()
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256([]byte(secret))
	entry := model.RelayIssuedToken{
		ID:          uuid.NewString(),
		Alias:       alias,
		Label:       strings.TrimSpace(req.Label),
		TokenSHA256: hex.EncodeToString(sum[:]),
		CreatedAt:   time.Now().UTC(),
		CreatedBy:   adminAlias,
	}
	file, err := store.GetRelayIssuedTokens(ctx)
	if err != nil {
		return nil, err
	}
	file.Tokens = append(file.Tokens, entry)
	if err := store.SaveRelayIssuedTokens(ctx, file); err != nil {
		return nil, err
	}
	return &dto.RelayIssueTokenResponse{
		Token:  mapIssuedToken(entry),
		Secret: secret,
	}, nil
}

func (s *Service) RevokeRelayIssuedToken(ctx context.Context, id string) (*dto.RelayIssuedTokenListResponse, error) {
	if _, err := s.requireRelayAdmin(ctx); err != nil {
		return nil, err
	}
	store, ok := s.repo.(relayIssuedTokenStore)
	if !ok {
		return nil, &ValidationError{Msg: "relay token store is not available"}
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, &ValidationError{Msg: "token id is required"}
	}
	file, err := store.GetRelayIssuedTokens(ctx)
	if err != nil {
		return nil, err
	}
	next := make([]model.RelayIssuedToken, 0, len(file.Tokens))
	found := false
	for _, t := range file.Tokens {
		if t.ID == id {
			found = true
			continue
		}
		next = append(next, t)
	}
	if !found {
		return nil, &ValidationError{Msg: "token not found"}
	}
	file.Tokens = next
	if err := store.SaveRelayIssuedTokens(ctx, file); err != nil {
		return nil, err
	}
	return s.ListRelayIssuedTokens(ctx)
}

func mapIssuedToken(t model.RelayIssuedToken) dto.RelayIssuedTokenResponse {
	return dto.RelayIssuedTokenResponse{
		ID:        t.ID,
		Alias:     t.Alias,
		Label:     t.Label,
		CreatedAt: t.CreatedAt.Format(time.RFC3339),
		CreatedBy: t.CreatedBy,
	}
}

func teamHasAlias(team []model.TeamPerson, alias string) bool {
	for _, p := range team {
		if strings.EqualFold(p.Alias, alias) {
			return true
		}
	}
	return false
}

func newRelayClientSecret() (string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b[:]), nil
}
