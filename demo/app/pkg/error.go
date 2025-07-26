package pkg

import (
	"fmt"
	"strings"
)

type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return e.Message
}

type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	messages := make([]string, len(e))
	for i, err := range e {
		messages[i] = err.Message
	}
	return strings.Join(messages, "\n\n")
}

type InternalError struct {
	Message string
	Err     error
}

func (e InternalError) Error() string {
	return fmt.Sprintf("%s: %s", e.Message, e.Err)
}

type BadRequestError struct {
	Message string
	Err     error
}
func (e BadRequestError) Error() string {
	return fmt.Sprintf("%s: %s", e.Message, e.Err)
}

type NotFoundError struct {
	Message string
	Err     error
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("%s: %s", e.Message, e.Err)
}

type UnauthorizedError struct {
	Err error
}

func (e UnauthorizedError) Error() string {
	return e.Err.Error()
}

type ForbiddenError struct{
	Err error
}

func (e ForbiddenError) Error() string {
	return e.Err.Error()
}
