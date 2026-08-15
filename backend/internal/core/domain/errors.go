package domain

import "errors"

var (
	ErrWalletNotFound      = errors.New("not found")
	ErrWalletAlreadyExists = errors.New("already exists")
	ErrWalletInternalError = errors.New("internal error")
	ErrPriceNotFound       = errors.New("not found")
)
