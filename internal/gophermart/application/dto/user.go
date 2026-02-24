package dto

// RegisterInput is the input for user registration.
type RegisterInput struct {
	Login    string
	Password string
}

// LoginInput is the input for login.
type LoginInput struct {
	Login    string
	Password string
}
