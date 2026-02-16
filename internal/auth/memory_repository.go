package auth

import "sync"

type MemoryUserRepository struct {
	mu    sync.RWMutex
	users map[string]User
}

func NewMemoryUserRepository() *MemoryUserRepository {
	return &MemoryUserRepository{
		users: make(map[string]User),
	}
}

func (r *MemoryUserRepository) Create(user User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.Email]; exists {
		return ErrUserExists
	}

	r.users[user.Email] = user
	return nil
}

func (r *MemoryUserRepository) FindByEmail(email string) (User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[email]
	if !exists {
		return User{}, ErrUserNotFound
	}

	return user, nil
}

func (r *MemoryUserRepository) Update(user User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.Email]; !exists {
		return ErrUserNotFound
	}

	r.users[user.Email] = user
	return nil
}
