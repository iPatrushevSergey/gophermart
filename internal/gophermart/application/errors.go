package application

import "errors"

var (
	// ErrNotFound — resource not found.
	ErrNotFound = errors.New("not found")

	// ErrAlreadyExists — resource already exists.
	ErrAlreadyExists = errors.New("already exists")

	// ErrConflict — operation conflicts with current state.
	ErrConflict = errors.New("conflict")

	// ErrInvalidCredentials — wrong login or password.
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrInsufficientBalance — not enough points on the balance account.
	ErrInsufficientBalance = errors.New("insufficient balance")

	// ErrInvalidOrderNumber — order number failed validation (e.g. Luhn check).
	ErrInvalidOrderNumber = errors.New("invalid order number")

	// ErrOptimisticLock — concurrent modification detected, operation should be retried.
	ErrOptimisticLock = errors.New("optimistic lock conflict")
)
