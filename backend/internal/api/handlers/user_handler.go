package handlers

import (
	"context"
	userv1 "tracker/gen/user/v1"
	"tracker/internal/api/middleware"
	"tracker/internal/core"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserHandler struct {
	userv1.UnimplementedUserServiceServer
	userRepo core.UserRepo
}

func NewUserHandler(repo core.UserRepo) userv1.UserServiceServer {
	return &UserHandler{
		userRepo: repo,
	}
}

func (p *UserHandler) GetUser(ctx context.Context, req *userv1.GetCurrentUserRequest) (*userv1.GetCurrentUserResponse, error) {
	user, ok := middleware.GetUser(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	return &userv1.GetCurrentUserResponse{
		Id:   user.ID,
		Name: user.Name,
	}, nil
}
