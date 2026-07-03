package dto

type RegisterRequest struct {
	Email       string  `json:"email"`
	Password    string  `json:"password"`
	FullName    string  `json:"full_name"`
	FullSurname string  `json:"full_surname"`
	Username    string  `json:"username"`
	NickName    string  `json:"nick_name,omitempty"`
	Post        *string `json:"post,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type AuthResponse struct {
	User UserResponse `json:"user"`
}

type UserResponse struct {
	ID          string  `json:"id"`
	Email       string  `json:"email"`
	FullName    string  `json:"full_name"`
	FullSurname string  `json:"full_surname"`
	Username    string  `json:"username"`
	Role        string  `json:"role"`
	Post        *string `json:"post,omitempty"`
	Phone       *string `json:"phone,omitempty"`
	Timezone    *string `json:"timezone,omitempty"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
	CreatedAt   string  `json:"created_at"`
}

type RegisterResponse struct {
	User UserResponse `json:"user"`
}

type UpdateUsernameRequest struct {
	Username string `json:"username"`
}

type UpdateProfileRequest struct {
	FullName    *string `json:"full_name"`
	FullSurname *string `json:"full_surname"`
	Post        *string `json:"post"`
	Phone       *string `json:"phone"`
	Timezone    *string `json:"timezone"`
}

type AvatarResponse struct {
	AvatarURL string `json:"avatar_url"`
	UpdatedAt string `json:"updated_at"`
}

type PreferencesDateRange struct {
	From *string `json:"from,omitempty"`
	To   *string `json:"to,omitempty"`
}

type UserPreferencesResponse struct {
	ActiveCompanyUUID *string              `json:"active_company_uuid"`
	Theme             string               `json:"theme"`
	DateRange         PreferencesDateRange `json:"date_range"`
}

type UpdatePreferencesRequest struct {
	ActiveCompanyUUID *string               `json:"active_company_uuid"`
	Theme             *string               `json:"theme"`
	DateRange         *PreferencesDateRange `json:"date_range"`
}
