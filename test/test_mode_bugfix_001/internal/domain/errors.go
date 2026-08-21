package domain

import "errors"

var (
	ErrInvalid      = errors.New("invalid command")
	ErrConflict     = errors.New("version conflict")
	ErrNotFound     = errors.New("not found")
	ErrInsufficient = errors.New("insufficient funds")
)
