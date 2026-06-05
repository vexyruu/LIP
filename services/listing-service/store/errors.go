package store

import "errors"

var ErrNotUnderReview = errors.New("listing is not under review")
var ErrUserNotFound = errors.New("user not found")
var ErrAlreadyBanned = errors.New("user is already banned")
