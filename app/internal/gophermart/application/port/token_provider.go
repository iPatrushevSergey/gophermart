package port

import "gophermart/internal/gophermart/domain/vo"

// TokenProvider issues and validates auth tokens.
type TokenProvider interface {
	Issue(userID vo.UserID) (token string, err error)
	Validate(token string) (userID vo.UserID, err error)
}
