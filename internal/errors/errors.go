package errors

import "errors"

var (
	ErrUserNotFound       = errors.New("用户不存在")
	ErrInvalidCredentials = errors.New("无效的用户名或密码")
)
