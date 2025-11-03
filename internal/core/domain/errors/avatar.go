package errors

import "errors"

var (
	ErrAvatarNotFound    = errors.New("avatar not found")
	ErrInvalidAvatarData = errors.New("invalid avatar data")
	ErrAvatarUploadFail  = errors.New("failed to upload avatar")
)
