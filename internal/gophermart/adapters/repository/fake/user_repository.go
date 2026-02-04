package fake

import (
	"context"
	"sync"

	"gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/domain/entity"
	"gophermart/internal/gophermart/domain/vo"
)

// UserRepository is an in-memory user repository.
type UserRepository struct {
	mu      sync.RWMutex
	byID    map[vo.UserID]*entity.User
	byLogin map[string]*entity.User
	nextID  int64
}

// NewUserRepository returns a new in-memory user repository.
func NewUserRepository() *UserRepository {
	return &UserRepository{
		byID:    make(map[vo.UserID]*entity.User),
		byLogin: make(map[string]*entity.User),
		nextID:  1,
	}
}

// Create persists a new user.
func (r *UserRepository) Create(ctx context.Context, u *entity.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if u.ID == 0 {
		u.ID = vo.UserID(r.nextID)
	}
	clone := *u
	r.byID[u.ID] = &clone
	r.byLogin[u.Login] = &clone
	return nil
}

// FindByID returns the user by ID or nil if not found.
func (r *UserRepository) FindByID(ctx context.Context, id vo.UserID) (*entity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.byID[id]
	if !ok {
		return nil, nil
	}
	clone := *u
	return &clone, nil
}

// FindByLogin returns the user by login or nil if not found.
func (r *UserRepository) FindByLogin(ctx context.Context, login string) (*entity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.byLogin[login]
	if !ok {
		return nil, nil
	}
	clone := *u
	return &clone, nil
}

var _ port.UserRepository = (*UserRepository)(nil)
