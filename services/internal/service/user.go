package service

import (
	"context"

	v1 "github.com/example/scaffold/services/api/identity/v1"
	"github.com/example/scaffold/services/internal/biz"
	"github.com/example/scaffold/services/internal/service/assembler"
)

// UserService adapts user usecases to the generated transport interfaces.
type UserService struct {
	v1.UnimplementedUserServiceServer
	uc *biz.UserUsecase
}

// NewUserService creates the user service.
func NewUserService(uc *biz.UserUsecase) *UserService {
	return &UserService{uc: uc}
}

func (s *UserService) ListUsers(ctx context.Context, req *v1.ListUsersRequest) (*v1.ListUsersResponse, error) {
	items, total, next, err := s.uc.List(ctx, biz.ListOptions{
		PageSize:  int(req.GetPageSize()),
		PageToken: req.GetPageToken(),
		Keyword:   req.GetKeyword(),
	})
	if err != nil {
		return nil, err
	}
	return &v1.ListUsersResponse{
		Items:         assembler.UsersToProto(items),
		NextPageToken: next,
		TotalSize:     int32(total),
	}, nil
}

func (s *UserService) CreateUser(ctx context.Context, req *v1.CreateUserRequest) (*v1.CreateUserResponse, error) {
	user, err := s.uc.Create(ctx, assembler.UserFromProto(req.GetUser()), req.GetPassword())
	if err != nil {
		return nil, err
	}
	return &v1.CreateUserResponse{User: assembler.UserToProto(user)}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, req *v1.UpdateUserRequest) (*v1.UpdateUserResponse, error) {
	user := assembler.UserFromProto(req.GetUser())
	if user == nil {
		return nil, biz.ErrInvalidArgument("user is required")
	}
	user.ID = req.GetId()
	updated, err := s.uc.Update(ctx, user, req.GetPassword())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateUserResponse{User: assembler.UserToProto(updated)}, nil
}

func (s *UserService) DeleteUser(ctx context.Context, req *v1.DeleteUserRequest) (*v1.DeleteUserResponse, error) {
	if err := s.uc.Delete(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &v1.DeleteUserResponse{}, nil
}
