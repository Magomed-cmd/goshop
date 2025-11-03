package errors

import "errors"

var (
	ErrInvalidRating          = errors.New("invalid rating of review")
	ErrInvalidReviewSortBy    = errors.New("invalid sort_by for review filter")
	ErrInvalidReviewSortOrder = errors.New("invalid sort_order for review filter")
	ErrInvalidReviewID        = errors.New("invalid id of review")
	ErrNothingToUpdate        = errors.New("not enough parameters for update")
	ErrInvalidComment         = errors.New("comment too long (max 1000 characters)")
	ErrReviewNotOwnedByUser   = errors.New("review does not belong to user")
)
