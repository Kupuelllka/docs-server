package service

import "errors"

// Вынести в отдельный middleware
var ErrInvalidAdminToken = errors.New("invalid token")
