package auth

type User struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	Password     string `json:"password"`      // hashed
	RefreshToken string `json:"refresh_token"` // hashed refresh token
}
