package dto

// RegisterRequest is the HTTP request body for registration.
type RegisterRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginRequest is the HTTP request body for login.
type LoginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}
