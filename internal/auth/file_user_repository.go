package auth

import (
	"encoding/json"
	"os"
	"sync"
)

type FileUserRepository struct {
	mu       sync.Mutex
	filePath string
}

// NewFileUserRepository creates a new file-backed user repository
func NewFileUserRepository(filePath string) (*FileUserRepository, error) {
	// Ensure file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err := os.WriteFile(filePath, []byte("{}"), 0644); err != nil {
			return nil, err
		}
	}

	return &FileUserRepository{
		filePath: filePath,
	}, nil
}

// readAll loads all users from the file
func (r *FileUserRepository) readAll() (map[string]User, error) {
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		return nil, err
	}

	users := make(map[string]User)
	if len(data) > 0 {
		if err := json.Unmarshal(data, &users); err != nil {
			return nil, err
		}
	}

	return users, nil
}

// writeAll writes all users to the file
func (r *FileUserRepository) writeAll(users map[string]User) error {
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(r.filePath, data, 0644)
}

// Create adds a new user
func (r *FileUserRepository) Create(user User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	users, err := r.readAll()
	if err != nil {
		return err
	}

	if _, exists := users[user.Email]; exists {
		return ErrUserExists
	}

	users[user.Email] = user
	return r.writeAll(users)
}

// FindByEmail retrieves a user by email
func (r *FileUserRepository) FindByEmail(email string) (User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	users, err := r.readAll()
	if err != nil {
		return User{}, err
	}

	user, exists := users[email]
	if !exists {
		return User{}, ErrUserNotFound
	}

	return user, nil
}

// Update modifies an existing user
func (r *FileUserRepository) Update(user User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	users, err := r.readAll()
	if err != nil {
		return err
	}

	if _, exists := users[user.Email]; !exists {
		return ErrUserNotFound
	}

	users[user.Email] = user
	return r.writeAll(users)
}

// Delete removes a user
func (r *FileUserRepository) Delete(email string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	users, err := r.readAll()
	if err != nil {
		return err
	}

	if _, exists := users[email]; !exists {
		return ErrUserNotFound
	}

	delete(users, email)
	return r.writeAll(users)
}
