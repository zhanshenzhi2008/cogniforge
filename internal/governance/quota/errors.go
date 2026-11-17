package quota

import "errors"

var (
	ErrRateLimited = errors.New("request rate limited")
	ErrExceeded    = errors.New("user quota exceeded")
	ErrUnavailable = errors.New("quota store unavailable")
)
