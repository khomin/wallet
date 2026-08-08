package handlers

import (
	"context"
	userv1 "tracker/gen/user/v1"
	"tracker/internal/api/middleware"
	"tracker/internal/core"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UserHandler struct {
	userv1.UnimplementedUserServiceServer
	userRepo core.UserRepo
}

func NewUserHandler(repo core.UserRepo) *UserHandler {
	return &UserHandler{
		userRepo: repo,
	}
}

func (p *UserHandler) GetUser(ctx context.Context, req *emptypb.Empty) (*userv1.GetCurrentUserResp, error) {
	user, ok := middleware.GetOAUTH(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	return &userv1.GetCurrentUserResp{
		Id:   user.Subject,
		Name: user.Name,
	}, nil
}
