package player

import (
	"errors"
)

var (
	ErrNotFound        = errors.New("player not found")
	ErrConflict        = errors.New("player with such email already exists")
	ErrIDMismatch      = errors.New("id mismatch")
	ErrVersionMismatch = errors.New("version mismatch")
	ErrEmptyRequest    = errors.New("request is empty")
)
