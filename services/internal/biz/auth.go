package biz

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"

	authpkg "github.com/example/scaffold/services/internal/pkg/auth"
	"github.com/example/scaffold/services/internal/pkg/id"
	"github.com/example/scaffold/services/internal/pkg/rbac"
)

// LoginResult is the token payload returned by login and refresh.
type LoginResult struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresIn    int32
	User         *User
}

// AuthUsecase owns login, token rotation, and logout.
type AuthUsecase struct {
	users   UserRepo
	tokens  RefreshTokenRepo
	tm      *authpkg.TokenManager
	checker rbac.Checker
}

// NewAuthUsecase wires the auth dependencies together.
func NewAuthUsecase(users UserRepo, tokens RefreshTokenRepo, tm *authpkg.TokenManager, checker rbac.Checker) *AuthUsecase {
	return &AuthUsecase{users: users, tokens: tokens, tm: tm, checker: checker}
}

func (uc *AuthUsecase) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	user, err := uc.users.GetByUsername(ctx, username)
	if err != nil {
		return nil, ErrInvalidCredentials()
	}
	if !authpkg.CheckPassword(user.PasswordHash, password) {
		return nil, ErrInvalidCredentials()
	}
	if user.Status != "" && user.Status != "active" {
		return nil, ErrForbidden()
	}
	return uc.issue(ctx, user)
}

func (uc *AuthUsecase) Refresh(ctx context.Context, rawToken string) (*LoginResult, error) {
	claims, err := uc.tm.ParseRefresh(rawToken)
	if err != nil {
		if errors.Is(err, authpkg.ErrTokenExpired) {
			return nil, ErrTokenExpired()
		}
		return nil, ErrTokenInvalid()
	}
	hash := hashToken(rawToken)
	stored, err := uc.tokens.GetByHash(ctx, hash)
	if err != nil || stored.Revoked {
		return nil, ErrTokenInvalid()
	}
	if stored.ExpiresAt > 0 && stored.ExpiresAt < NowMillis() {
		return nil, ErrTokenExpired()
	}
	if err := uc.tokens.RevokeByHash(ctx, hash); err != nil {
		return nil, ErrInternal("failed to rotate refresh token")
	}
	user, err := uc.users.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, ErrTokenInvalid()
	}
	return uc.issue(ctx, user)
}

func (uc *AuthUsecase) Logout(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return nil
	}
	return uc.tokens.RevokeByHash(ctx, hashToken(rawToken))
}

func (uc *AuthUsecase) CurrentUser(ctx context.Context, userID string) (*User, error) {
	return uc.users.GetByID(ctx, userID)
}

// Permissions returns the caller's permission codes.
func (uc *AuthUsecase) Permissions(ctx context.Context, userID string) ([]string, error) {
	return uc.checker.Permissions(ctx, userID)
}

// IssueForUser issues a token pair for an already authenticated user. It is
// used by OAuth and other external identity flows.
func (uc *AuthUsecase) IssueForUser(ctx context.Context, user *User) (*LoginResult, error) {
	return uc.issue(ctx, user)
}

// ChangePassword rotates the caller's password and clears the forced-rotation
// flag.
func (uc *AuthUsecase) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	if len(newPassword) < 8 {
		return ErrInvalidArgument("new password must be at least 8 characters")
	}
	user, err := uc.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if !authpkg.CheckPassword(user.PasswordHash, oldPassword) {
		return ErrInvalidCredentials()
	}
	hash, err := authpkg.HashPassword(newPassword)
	if err != nil {
		return ErrInternal("failed to hash password")
	}
	user.PasswordHash = hash
	user.MustChangePassword = false
	if _, err := uc.users.Update(ctx, user); err != nil {
		return err
	}
	return nil
}

func (uc *AuthUsecase) issue(ctx context.Context, user *User) (*LoginResult, error) {
	sessionID := id.New()
	access, refresh, expiresIn, err := uc.tm.Issue(user.ID, user.Username, sessionID, user.IsAdmin, user.MustChangePassword)
	if err != nil {
		return nil, ErrInternal("failed to issue tokens")
	}
	if err := uc.tokens.Create(ctx, &RefreshToken{
		ID:        id.New(),
		UserID:    user.ID,
		TokenHash: hashToken(refresh),
		SessionID: sessionID,
		ExpiresAt: NowMillis() + uc.tm.RefreshTTL().Milliseconds(),
	}); err != nil {
		return nil, ErrInternal("failed to persist refresh token")
	}
	return &LoginResult{
		AccessToken:  access,
		RefreshToken: refresh,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
		User:         user,
	}, nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
