package biz

import (
	"context"
	"regexp"
	"strings"

	authpkg "github.com/example/scaffold/services/internal/pkg/auth"
	"github.com/example/scaffold/services/internal/pkg/id"
	"github.com/example/scaffold/services/internal/pkg/pagination"
)

var usernamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{2,31}$`)

// UserUsecase contains account management rules.
type UserUsecase struct {
	repo UserRepo
}

// NewUserUsecase wires the user repository into the usecase.
func NewUserUsecase(repo UserRepo) *UserUsecase {
	return &UserUsecase{repo: repo}
}

func (uc *UserUsecase) Create(ctx context.Context, user *User, password string) (*User, error) {
	user.Username = strings.ToLower(strings.TrimSpace(user.Username))
	if !usernamePattern.MatchString(user.Username) {
		return nil, ErrInvalidArgument("username must match ^[a-z0-9][a-z0-9._-]{2,31}$")
	}
	if len(password) < 6 {
		return nil, ErrInvalidArgument("password must be at least 6 characters")
	}
	hash, err := authpkg.HashPassword(password)
	if err != nil {
		return nil, ErrInternal("failed to hash password")
	}
	user.ID = id.New()
	user.PasswordHash = hash
	if user.Status == "" {
		user.Status = "active"
	}
	return uc.repo.Create(ctx, user)
}

func (uc *UserUsecase) Get(ctx context.Context, userID string) (*User, error) {
	return uc.repo.GetByID(ctx, userID)
}

func (uc *UserUsecase) List(ctx context.Context, opts ListOptions) ([]*User, int, string, error) {
	opts.PageSize = pagination.Normalize(opts.PageSize)
	items, total, err := uc.repo.List(ctx, opts)
	if err != nil {
		return nil, 0, "", err
	}
	offset, err := pagination.DecodeOffset(opts.PageToken)
	if err != nil {
		return nil, 0, "", ErrInvalidArgument("invalid page token")
	}
	next := ""
	if offset+len(items) < total {
		next = pagination.EncodeOffset(offset + len(items))
	}
	return items, total, next, nil
}

func (uc *UserUsecase) Update(ctx context.Context, user *User, password string) (*User, error) {
	if password != "" {
		if len(password) < 6 {
			return nil, ErrInvalidArgument("password must be at least 6 characters")
		}
		hash, err := authpkg.HashPassword(password)
		if err != nil {
			return nil, ErrInternal("failed to hash password")
		}
		user.PasswordHash = hash
	}
	return uc.repo.Update(ctx, user)
}

func (uc *UserUsecase) Delete(ctx context.Context, userID string) error {
	return uc.repo.Delete(ctx, userID)
}
