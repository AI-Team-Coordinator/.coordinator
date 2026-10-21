package connector

import "errors"

var (
	ErrIgnored      = errors.New("connector: event ignored")
	ErrBadSignature = errors.New("connector: webhook signature mismatch")
)
