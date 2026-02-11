package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Service contains business logic for authentication
type Service struct {
	repo        UserRepository
	jwtSecret   []byte
	tokenExpiry time.Duration
}

// LoginInput represents login request data
type LoginInput struct {
	Email    string
	Password string
}

// NewService creates a new Service instance with a secret key for JWT
func NewService(repo UserRepository, jwtSecret []byte, tokenExpiry time.Duration) *Service {
	return &Service{
		repo:        repo,
		jwtSecret:   jwtSecret,
		tokenExpiry: tokenExpiry,
	}
}

// Register a new user
func (s *Service) Register(email, password string) error {
	if email == "" || password == "" {
		return errors.New("email and password are required")
	}

	hash, err := HashPassword(password)
	if err != nil {
		return err
	}

	user := User{
		ID:       uuid.NewString(),
		Email:    email,
		Password: hash,
	}

	return s.repo.Create(user)
}

// Login checks credentials and returns the user if successful
func (s *Service) Login(input LoginInput) (authUser User, err error) {
	if input.Email == "" || input.Password == "" {
		return User{}, errors.New("email and password are required")
	}

	user, err := s.repo.FindByEmail(input.Email)
	if err != nil {
		return User{}, errors.New("invalid credentials")
	}

	if err := CheckPassword(input.Password, user.Password); err != nil {
		return User{}, errors.New("invalid credentials")
	}

	return user, nil
}

// GenerateToken creates a JWT token for a user
func (s *Service) GenerateToken(user User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(), // token expires in 24h
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
