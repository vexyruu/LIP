package store

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

func isUniqueViolation(err error) bool {
	var pg *pgconn.PgError
	return errors.As(err, &pg) && pg.Code == "23505"
}

var ErrNotUnderReview = errors.New("listing is not under review")
var ErrUserNotFound = errors.New("user not found")
var ErrAlreadyBanned = errors.New("user is already banned")
var ErrAlreadyAssigned = errors.New("listing is already claimed by another moderator")
var ErrNotAssignedToYou = errors.New("listing is not claimed by you")
var ErrAssignedToOther = errors.New("listing is claimed by another moderator")
