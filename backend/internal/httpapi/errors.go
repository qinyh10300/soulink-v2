package httpapi

import (
	"errors"
	"net/http"
)

const (
	CodeSuccess                     = 0
	CodeInvalidParams               = 4001
	CodeInvalidVerificationCode     = 4002
	CodeVerificationCodeExpired     = 4003
	CodeVerificationCodeTooFrequent = 4004
	CodeInvalidEmail                = 4005
	CodeInvalidPassword             = 4006
	CodePasswordMismatch            = 4007
	CodeUnauthorized                = 4011
	CodeInvalidToken                = 4012
	CodeInvalidCredentials          = 4013
	CodeForbidden                   = 4031
	CodeUserNotFound                = 4041
	CodeQuestionnaireNotFound       = 4042
	CodeNoCurrentMatch              = 4043
	CodeContactNotFound             = 4044
	CodeEmailRegistered             = 4091
	CodeConflict                    = 4092
	CodeInternalError               = 5001
)

type AppError struct {
	HTTPStatus int
	Code       int
	Message    string
	Err        error
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// NewError 创建一个带业务码的应用错误。
func NewError(status, code int, message string) *AppError {
	return &AppError{
		HTTPStatus: status,
		Code:       code,
		Message:    message,
	}
}

// Wrap 将底层错误包装成统一的应用错误。
func Wrap(err error, status, code int, message string) *AppError {
	return &AppError{
		HTTPStatus: status,
		Code:       code,
		Message:    message,
		Err:        err,
	}
}

// InternalError 将未知错误转换成通用服务器错误。
func InternalError(err error) *AppError {
	return Wrap(err, http.StatusInternalServerError, CodeInternalError, "服务器内部错误")
}

// FeatureNotReady 返回功能尚未实现时的统一错误。
func FeatureNotReady(feature string) *AppError {
	return NewError(http.StatusNotImplemented, CodeInternalError, feature+" not implemented yet")
}

// AsAppError 将任意错误归一化为应用错误。
func AsAppError(err error) *AppError {
	if err == nil {
		return nil
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	return InternalError(err)
}
