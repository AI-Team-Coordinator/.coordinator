package dto

type HealthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	StartedAt string `json:"started_at"`
}
