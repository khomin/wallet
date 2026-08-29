package domain

import "errors"

var (
	ErrorNotFound             = errors.New("Not found")
	ErrorWWalletAlreadyExists = errors.New("Already exists")
	ErrorWWalletInternalError = errors.New("Internal error")
	ErrInvalidArgument        = errors.New("Invalid argument")
	ErrNotAllowedInDemoMode   = errors.New("Cannot work in demo mode")
)
