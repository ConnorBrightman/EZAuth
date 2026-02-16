package auth

import (
	"crypto/rand"
	"encoding/base64"
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
	Email        string
	Password     string
	RefreshToken string
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
func (s *Service) GenerateAccessToken(user User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(), // token expires in 24h
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// GenerateRefreshToken creates a random refresh token string
func GenerateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// LoginWithTokens logs in user and returns both access and refresh tokens
func (s *Service) LoginWithTokens(input LoginInput) (accessToken string, refreshToken string, err error) {
	user, err := s.Login(input)
	if err != nil {
		return "", "", err
	}

	// Generate access token
	accessToken, err = s.GenerateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	// Generate refresh token
	refreshToken, err = GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}

	// Store hashed refresh token in repo
	hashedRefresh, err := HashPassword(refreshToken)
	if err != nil {
		return "", "", err
	}

	user.RefreshToken = hashedRefresh
	if err := s.repo.Update(user); err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// RefreshAccessToken verifies refresh token and returns new access token
func (s *Service) RefreshAccessToken(userEmail, refreshToken string) (string, error) {
	user, err := s.repo.FindByEmail(userEmail)
	if err != nil {
		return "", errors.New("user not found")
	}

	if err := CheckPassword(refreshToken, user.RefreshToken); err != nil {
		return "", errors.New("invalid refresh token")
	}

	return s.GenerateAccessToken(user)
}

// RefreshTokens verifies the old refresh token and returns new access + refresh tokens
func (s *Service) RefreshTokens(userEmail, oldRefreshToken string) (newAccessToken, newRefreshToken string, err error) {
	user, err := s.repo.FindByEmail(userEmail)
	if err != nil {
		return "", "", errors.New("user not found")
	}

	// Check old refresh token
	if err := CheckPassword(oldRefreshToken, user.RefreshToken); err != nil {
		return "", "", errors.New("invalid refresh token")
	}

	// Generate new access token
	newAccessToken, err = s.GenerateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	// Generate new refresh token
	newRefreshToken, err = GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}

	// Store hashed new refresh token
	hashedRefresh, err := HashPassword(newRefreshToken)
	if err != nil {
		return "", "", err
	}
	user.RefreshToken = hashedRefresh
	if err := s.repo.Update(user); err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}
