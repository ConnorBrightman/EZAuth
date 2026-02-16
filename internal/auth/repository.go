package auth

import "errors"

var ErrUserNotFound = errors.New("user not found")
var ErrUserExists = errors.New("user already exists")

type UserRepository interface {
	Create(user User) error
	FindByEmail(email string) (User, error)
	Update(user User) error
}
