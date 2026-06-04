package dto

type RegisterRequest struct {
	Email       string  `json:"email"`
	Password    string  `json:"password"`
	FullName    string  `json:"full_name"`
	FullSurname string  `json:"full_surname"`
	NickName    string  `json:"nick_name"`
	Post        *string `json:"post,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	AccessToken string       `json:"access_token"`
	User        UserResponse `json:"user"`
}

type UserResponse struct {
	ID          string  `json:"id"`
	Email       string  `json:"email"`
	FullName    string  `json:"full_name"`
	FullSurname string  `json:"full_surname"`
	NickName    string  `json:"nick_name"`
	Role        string  `json:"role"`
	Post        *string `json:"post,omitempty"`
	CreatedAt   string  `json:"created_at"`
}

type RegisterResponse struct {
	User UserResponse `json:"user"`
}
