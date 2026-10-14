package model

import "time"

type RelayIssuedToken struct {
	ID          string    `json:"id"`
	Alias       string    `json:"alias"`
	Label       string    `json:"label,omitempty"`
	TokenSHA256 string    `json:"token_sha256"`
	CreatedAt   time.Time `json:"created_at"`
	CreatedBy   string    `json:"created_by"`
}

type RelayIssuedTokenFile struct {
	Version int                `json:"version"`
	Tokens  []RelayIssuedToken `json:"tokens"`
}
