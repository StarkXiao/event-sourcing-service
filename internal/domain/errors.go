package domain

import (
	"errors"
	"fmt"
)

var (
	ErrInvalid       = errors.New("invalid command")
	ErrConflict      = errors.New("version conflict")
	ErrNotFound      = errors.New("not found")
	ErrInsufficient  = errors.New("insufficient funds")
	ErrProjectionLag = errors.New("projection lag")
)

func ProjectionLagError(required int64) error {
	return fmt.Errorf("%v: version %d required", ErrProjectionLag, required)
}
