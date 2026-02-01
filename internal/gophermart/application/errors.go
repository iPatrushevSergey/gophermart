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
)
