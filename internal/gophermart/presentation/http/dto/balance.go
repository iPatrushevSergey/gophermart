package dto

//go:generate easyjson -all $GOFILE

// BalanceResponse is the HTTP response body for the user balance.
type BalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

// WithdrawRequest is the HTTP request body for a withdrawal.
type WithdrawRequest struct {
	Order string  `json:"order" binding:"required"`
	Sum   float64 `json:"sum" binding:"required,gt=0"`
}
