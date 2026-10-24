package service

import (
	"context"

	identityv1 "github.com/example/scaffold/services/api/identity/v1"
	"github.com/example/scaffold/services/internal/biz"
	authpkg "github.com/example/scaffold/services/internal/pkg/auth"
	"github.com/example/scaffold/services/internal/service/assembler"
)

// AuthService adapts auth usecases to the generated transport interfaces.
type AuthService struct {
	identityv1.UnimplementedAuthServiceServer
	uc *biz.AuthUsecase
}

// NewAuthService creates the auth service.
func NewAuthService(uc *biz.AuthUsecase) *AuthService {
	return &AuthService{uc: uc}
}

func (s *AuthService) Login(ctx context.Context, req *identityv1.LoginRequest) (*identityv1.LoginResponse, error) {
	result, err := s.uc.Login(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		return nil, err
	}
	return &identityv1.LoginResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		TokenType:    result.TokenType,
		ExpiresIn:    result.ExpiresIn,
		User:         assembler.UserToProto(result.User),
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, req *identityv1.RefreshRequest) (*identityv1.RefreshResponse, error) {
	result, err := s.uc.Refresh(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, err
	}
	return &identityv1.RefreshResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		TokenType:    result.TokenType,
		ExpiresIn:    result.ExpiresIn,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, req *identityv1.LogoutRequest) (*identityv1.LogoutResponse, error) {
	if err := s.uc.Logout(ctx, req.GetRefreshToken()); err != nil {
		return nil, err
	}
	return &identityv1.LogoutResponse{}, nil
}

func (s *AuthService) GetCurrentUser(ctx context.Context, _ *identityv1.GetCurrentUserRequest) (*identityv1.GetCurrentUserResponse, error) {
	principal, ok := authpkg.PrincipalFromContext(ctx)
	if !ok {
		return nil, biz.ErrUnauthorized()
	}
	user, err := s.uc.CurrentUser(ctx, principal.UserID)
	if err != nil {
		return nil, err
	}
	permissions, err := s.uc.Permissions(ctx, principal.UserID)
	if err != nil {
		return nil, err
	}
	return &identityv1.GetCurrentUserResponse{
		User:        assembler.UserToProto(user),
		Permissions: permissions,
	}, nil
}

func (s *AuthService) ChangePassword(ctx context.Context, req *identityv1.ChangePasswordRequest) (*identityv1.ChangePasswordResponse, error) {
	principal, ok := authpkg.PrincipalFromContext(ctx)
	if !ok {
		return nil, biz.ErrUnauthorized()
	}
	if err := s.uc.ChangePassword(ctx, principal.UserID, req.GetOldPassword(), req.GetNewPassword()); err != nil {
		return nil, err
	}
	return &identityv1.ChangePasswordResponse{}, nil
}
