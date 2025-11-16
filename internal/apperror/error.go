package apperror

import "errors"

type Code string

type Error struct {
	Code Code
	Err  error
}

func New(code Code, err error) *Error {
	if err == nil {
		err = errors.New(string(code))
	}
	return &Error{Code: code, Err: err}
}

func (e *Error) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return string(e.Code)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func CodeOf(err error) (Code, bool) {
	var target *Error
	if errors.As(err, &target) {
		return target.Code, true
	}
	return "", false
}

func IsCode(err error, code Code) bool {
	c, ok := CodeOf(err)
	return ok && c == code
}
