package domain

import "errors"

var (
	ErrorNotFound            = errors.New("Not found")
	ErrorWalletAlreadyExists = errors.New("Already exists")
	ErrorInternalError       = errors.New("Internal error")
	ErrInvalidArgument       = errors.New("Invalid argument")
	ErrNotAllowedInDemoMode  = errors.New("Not allowed in demo mode")
)
