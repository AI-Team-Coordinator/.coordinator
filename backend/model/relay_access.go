package model

// RelayAccess is stored in Common/data/.relay_access (gitignored on each machine).
type RelayAccess struct {
	URL   string `json:"url"`
	Token string `json:"token,omitempty"`
	Scope string `json:"scope,omitempty"`
}
