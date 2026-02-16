package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUserService_CreateUser(t *testing.T) {
	svc := UserService{}
	now := time.Now()

	user := svc.CreateUser("alice", "hash123", now)

	assert.Equal(t, "alice", user.Login)
	assert.Equal(t, "hash123", user.PasswordHash)
	assert.Equal(t, now, user.CreatedAt)
	assert.Equal(t, now, user.UpdatedAt)
}
