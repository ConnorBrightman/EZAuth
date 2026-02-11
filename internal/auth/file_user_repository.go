package auth

import (
	"encoding/json"
	"os"
	"sync"
)

type FileUserRepository struct {
	mu       sync.RWMutex
	users    map[string]User
	filePath string
}

// NewFileUserRepository creates a new file-backed user repository
func NewFileUserRepository(filePath string) (*FileUserRepository, error) {
	repo := &FileUserRepository{
		users:    make(map[string]User),
		filePath: filePath,
	}

	// Load existing users from file, if it exists
	if _, err := os.Stat(filePath); err == nil {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		if len(data) > 0 {
			if err := json.Unmarshal(data, &repo.users); err != nil {
				return nil, err
			}
		}
	}

	return repo, nil
}

// save writes the current users map to the file
func (r *FileUserRepository) save() error {
	data, err := json.MarshalIndent(r.users, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(r.filePath, data, 0644)
}

// Create adds a new user
func (r *FileUserRepository) Create(user User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.Email]; exists {
		return ErrUserExists
	}

	r.users[user.Email] = user
	return r.save()
}

// FindByEmail retrieves a user by email
func (r *FileUserRepository) FindByEmail(email string) (User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[email]
	if !exists {
		return User{}, ErrUserNotFound
	}

	return user, nil
}
