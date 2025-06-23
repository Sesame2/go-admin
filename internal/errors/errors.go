package errors

import "errors"

var (
	ErrUserNotFound        = errors.New("用户不存在")
	ErrInvalidCredentials  = errors.New("无效的用户名或密码")
	ErrParseToken          = errors.New("解析令牌失败")
	ErrInvalidToken        = errors.New("无效的令牌")
	ErrInvalidUserIDFormat = errors.New("无效的用户ID格式")
)
