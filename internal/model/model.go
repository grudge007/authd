package model

type USER struct {
	UserName string `json:"username" validate:"required,alphanum,min=3,max=20"`
	Password string `json:"password" validate:"required,min=3,max=10"`
	Email    string `json:"email"    validate:"required,email"`
}

type LoginRequest struct {
	UserName string `json:"username" validate:"required,alphanum,min=3,max=20"`
	Password string `json:"password" validate:"required,min=3,max=10"`
}

type LoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Message      string `json:"message"`
}

type ChangePassword struct {
	OldPassword string `json:"old_password" validate:"required,min=3,max=10"`
	NewPassword string `json:"new_password" validate:"required,min=3,max=10"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
