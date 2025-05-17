package storage

import "errors"

var (
	ErrUserExists           = errors.New("person already exists")
	ErrSubscriptionExists   = errors.New("subscription with that number already exists")
	ErrPersonNotFound       = errors.New("person not found")
	ErrSubscriptionNotFound = errors.New("subscription not found")
	ErrAppNotFound          = errors.New("app not found")
)
