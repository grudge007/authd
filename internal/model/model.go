package model

type USER struct {
	UserName string `json:"username" validate:"required,alphanum,min=3,max=20"`
	Password string `json:"password" validate:"required,min=3,max=10"`
	Email    string `json:"email"    validate:"required,email"`
}

type LoginRequest struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token   string `json:"token"`
	Message string `json:"message"`
}
