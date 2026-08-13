package apperror

import "net/http"

type Error struct {
	Status int
	Detail string
}

func (e Error) Error() string {
	return e.Detail
}

func New(status int, detail string) Error {
	return Error{Status: status, Detail: detail}
}

func BadRequest(detail string) Error {
	return New(http.StatusBadRequest, detail)
}

func Unauthorized(detail string) Error {
	return New(http.StatusUnauthorized, detail)
}

func Forbidden(detail string) Error {
	return New(http.StatusForbidden, detail)
}

func Internal(detail string) Error {
	return New(http.StatusInternalServerError, detail)
}
