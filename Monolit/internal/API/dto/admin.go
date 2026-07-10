package dto

type AdminCapabilitiesResponse struct {
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}
