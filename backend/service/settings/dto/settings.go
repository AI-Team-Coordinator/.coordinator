package settingsdto

type UserSettingsResponse struct {
	Alias     string `json:"alias"`
	Name      string `json:"name"`
	Role      string `json:"role,omitempty"`
	Access    string `json:"access,omitempty"`
	Language  string `json:"language"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

type SaveUserSettingsRequest struct {
	Name     string `json:"name"`
	Language string `json:"language"`
}
