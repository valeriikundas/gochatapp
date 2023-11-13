package main

type UnauthorizedUserError struct{}

func (e *UnauthorizedUserError) Error() string {
	return "user is unauthorized"
}
