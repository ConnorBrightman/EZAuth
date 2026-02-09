package auth

import "errors"

type LoginInput struct {
	Email    string
	Password string
}

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Login(input LoginInput) error {
	if input.Email == "" || input.Password == "" {
		return errors.New("email and password are required")
	}

	// Later: lookup user, compare hash, etc
	return nil
}
