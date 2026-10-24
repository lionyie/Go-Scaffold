package biz

import (
	"context"
	"fmt"
	"strings"

	"github.com/example/scaffold/services/internal/auth/oauth"
	"github.com/example/scaffold/services/internal/conf"
	authpkg "github.com/example/scaffold/services/internal/pkg/auth"
	"github.com/example/scaffold/services/internal/pkg/id"
)

// OAuthAccount is a link between a local user and a provider identity.
type OAuthAccount struct {
	Provider       string
	ProviderUserID string
	Email          string
	EmailVerified  bool
	Name           string
	Avatar         string
	UserID         string
}

// OAuthAccountRepo persists provider links.
type OAuthAccountRepo interface {
	FindUserID(ctx context.Context, provider, providerUserID string) (string, error)
	Link(ctx context.Context, account *OAuthAccount) error
	ListByUser(ctx context.Context, userID string) ([]*OAuthAccount, error)
	Unlink(ctx context.Context, userID, provider string) error
}

// OAuthUsecase resolves a provider identity to a local session.
type OAuthUsecase struct {
	accounts     OAuthAccountRepo
	users        UserRepo
	auth         *AuthUsecase
	autoRegister bool
}

// NewOAuthUsecase wires the OAuth dependencies together.
func NewOAuthUsecase(accounts OAuthAccountRepo, users UserRepo, auth *AuthUsecase, bc *conf.Bootstrap) *OAuthUsecase {
	return &OAuthUsecase{
		accounts:     accounts,
		users:        users,
		auth:         auth,
		autoRegister: bc.GetAuth().GetOauth().GetAutoRegister(),
	}
}

// LoginOrRegister returns a session for a provider identity.
func (uc *OAuthUsecase) LoginOrRegister(ctx context.Context, provider string, u *oauth.OAuthUser) (*LoginResult, error) {
	if u == nil || u.ProviderUserID == "" {
		return nil, ErrInvalidArgument("provider did not return a user id")
	}

	account := &OAuthAccount{
		Provider:       provider,
		ProviderUserID: u.ProviderUserID,
		Email:          u.Email,
		EmailVerified:  u.EmailVerified,
		Name:           u.Name,
		Avatar:         u.Avatar,
	}

	userID, err := uc.accounts.FindUserID(ctx, provider, u.ProviderUserID)
	if err != nil {
		return nil, err
	}
	if userID == "" && oauth.CanLinkByEmail(u) {
		if existing, lookupErr := uc.users.GetByEmail(ctx, u.Email); lookupErr == nil {
			userID = existing.ID
		}
	}
	if userID == "" {
		if !uc.autoRegister {
			return nil, oauth.ErrNotLinked
		}
		created, createErr := uc.createUser(ctx, provider, u)
		if createErr != nil {
			return nil, createErr
		}
		userID = created.ID
	}

	account.UserID = userID
	if err := uc.accounts.Link(ctx, account); err != nil {
		return nil, err
	}
	user, err := uc.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return uc.auth.IssueForUser(ctx, user)
}

// LinkAccount binds a provider identity to an authenticated user.
func (uc *OAuthUsecase) LinkAccount(ctx context.Context, userID, provider string, u *oauth.OAuthUser) error {
	if u == nil || u.ProviderUserID == "" {
		return ErrInvalidArgument("provider did not return a user id")
	}
	return uc.accounts.Link(ctx, &OAuthAccount{
		Provider:       provider,
		ProviderUserID: u.ProviderUserID,
		Email:          u.Email,
		EmailVerified:  u.EmailVerified,
		Name:           u.Name,
		Avatar:         u.Avatar,
		UserID:         userID,
	})
}

// ListAccounts returns the provider links owned by a user.
func (uc *OAuthUsecase) ListAccounts(ctx context.Context, userID string) ([]*OAuthAccount, error) {
	return uc.accounts.ListByUser(ctx, userID)
}

// Unlink removes a provider link owned by a user.
func (uc *OAuthUsecase) Unlink(ctx context.Context, userID, provider string) error {
	return uc.accounts.Unlink(ctx, userID, provider)
}

func (uc *OAuthUsecase) createUser(ctx context.Context, provider string, u *oauth.OAuthUser) (*User, error) {
	suffix, err := oauth.RandomToken(6)
	if err != nil {
		return nil, ErrInternal("failed to generate a username")
	}
	password, err := oauth.RandomToken(24)
	if err != nil {
		return nil, ErrInternal("failed to generate a password")
	}
	hash, err := authpkg.HashPassword(password)
	if err != nil {
		return nil, ErrInternal("failed to hash password")
	}
	username := sanitizeUsername(provider + "_" + suffix)
	return uc.users.Create(ctx, &User{
		ID:           id.New(),
		Username:     username,
		DisplayName:  strings.TrimSpace(u.Name),
		Email:        strings.TrimSpace(u.Email),
		PasswordHash: hash,
		Status:       "active",
	})
}

func sanitizeUsername(value string) string {
	value = strings.ToLower(value)
	var b strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '.', r == '_', r == '-':
			b.WriteRune(r)
		}
	}
	out := b.String()
	if len(out) < 3 {
		return fmt.Sprintf("user_%s", out)
	}
	return out
}

// ErrOAuthNotLinked exposes the sentinel without leaking package details.
func ErrOAuthNotLinked() error { return oauth.ErrNotLinked }
