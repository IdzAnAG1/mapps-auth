package validation

import (
	"regexp"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,50}$`)
)

func ValidateRegisterRequest(email, password, username string) error {
	if !emailRegex.MatchString(email) {
		return status.Error(codes.InvalidArgument, "invalid email format")
	}
	if len(password) < 8 {
		return status.Error(codes.InvalidArgument, "password must be at least 8 characters")
	}
	if !usernameRegex.MatchString(username) {
		return status.Error(codes.InvalidArgument, "username must be 3-50 characters, only letters, digits and underscores")
	}
	return nil
}

func ValidateLoginRequest(email, password string) error {
	if email == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}
	if password == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}
	return nil
}
