package relaydto

type RelayAccessResponse struct {
	Connected bool   `json:"connected"`
	URL       string `json:"url,omitempty"`
	Scope     string `json:"scope,omitempty"`
}

type RelayConnectRequest struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}

type RelaySourceResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Detail  string `json:"detail"`
	Enabled bool   `json:"enabled"`
}

type RelaySourcesResponse struct {
	Sources []RelaySourceResponse `json:"sources"`
}
