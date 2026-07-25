package model

type USER struct {
	UserName string `json:"username" validate:"required,alphanum,min=3,max=10"`
	Password string `json:"password" validate:"required,min=3,max=20"`
	Email    string `json:"email"    validate:"required,email"`
}

type LoginRequest struct {
	UserName string `json:"username" validate:"required,alphanum,min=3,max=1go0"`
	Password string `json:"password" validate:"required,min=3,max=20"`
}

type LoginResponse struct {
	Token   string `json:"token"`
	Message string `json:"message"`
}

type ChangePassword struct {
	OldPassword string `json:"old_password" validate:"required,min=3,max=10"`
	NewPassword string `json:"new_password" validate:"required,min=3,max=20"`
}
