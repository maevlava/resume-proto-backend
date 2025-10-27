package httperror

import "net/http"

type AppError struct {
	Message    string `json:"message"`
	Code       string `json:"code"`
	Err        error  `json:"error"`
	StatusCode int    `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}
func (e *AppError) Unwrap() error {
	return e.Err
}

func BadRequest(message string, err error) *AppError {
	return &AppError{
		Message:    message,
		Code:       "Bad Request",
		StatusCode: http.StatusBadRequest,
		Err:        err,
	}
}

func Unauthorized(message string) *AppError {
	return &AppError{
		Message:    message,
		Code:       "Unauthorized",
		StatusCode: http.StatusUnauthorized,
	}
}

func InternalServerError(message string, err error) *AppError {
	return &AppError{
		Message:    message,
		Code:       "Internal Server Error",
		StatusCode: http.StatusInternalServerError,
		Err:        err,
	}
}
