package store

import "errors"

var ErrNotUnderReview = errors.New("listing is not under review")
var ErrUserNotFound = errors.New("user not found")
var ErrAlreadyBanned = errors.New("user is already banned")
var ErrAlreadyAssigned = errors.New("listing is already claimed by another moderator")
var ErrNotAssignedToYou = errors.New("listing is not claimed by you")
