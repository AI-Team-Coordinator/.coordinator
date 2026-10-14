package dto

type RelayIssuedTokenResponse struct {
	ID        string `json:"id"`
	Alias     string `json:"alias"`
	Label     string `json:"label,omitempty"`
	CreatedAt string `json:"created_at"`
	CreatedBy string `json:"created_by"`
}

type RelayIssuedTokenListResponse struct {
	Tokens []RelayIssuedTokenResponse `json:"tokens"`
}

type RelayIssueTokenRequest struct {
	Alias string `json:"alias"`
	Label string `json:"label,omitempty"`
}

type RelayIssueTokenResponse struct {
	Token RelayIssuedTokenResponse `json:"token"`
	// Plaintext secret — only in this response.
	Secret string `json:"secret"`
}
