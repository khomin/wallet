package handlers

import (
	"context"
	"time"
	alertv1 "tracker/gen/alert/v1"
	"tracker/internal/api/middleware"
	"tracker/internal/core"
	"tracker/internal/core/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AlertHandler struct {
	alertRepo core.AlertRepository
	userRepo  core.UserRepo
}

func NewAlertHandler(repo core.AlertRepository, userRepo core.UserRepo) alertv1.AlertServiceServer {
	return &AlertHandler{
		alertRepo: repo,
		userRepo:  userRepo,
	}
}

func (a *AlertHandler) CreateAlert(ctx context.Context, req *alertv1.CreateAlertRequest) (*alertv1.Alert, error) {
	if req.CoinId == "" || req.Condition == alertv1.Condition_CONDITION_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "missing or invalid required arguments")
	}
	user, ok := middleware.GetOAUTH(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	if err := a.userRepo.EnsureExists(ctx, domain.User{
		ID:    user.Subject,
		Name:  user.Name,
		Email: user.Email,
	}); err != nil {
		return nil, err
	}
	result, err := a.alertRepo.Create(ctx, domain.Alert{
		UserID:     user.Subject,
		CoinSymbol: req.CoinId,
		Condition:  conditionString(req.Condition),
		Price:      req.Price,
		Enabled:    true,
		CreatedAt:  time.Now(),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	alert := alertToGrpc(result)
	return &alert, nil
}

func (a *AlertHandler) UpdateAlert(ctx context.Context, req *alertv1.UpdateAlertRequest) (*alertv1.Alert, error) {
	user, ok := middleware.GetOAUTH(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	result, err := a.alertRepo.Update(ctx, user.Subject, req.Id, domain.AlertUpdate{
		Condition: conditionString(req.Condition),
		Price:     req.Price,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	alert := alertToGrpc(result)
	return &alert, nil
}

func (a *AlertHandler) PauseAlert(ctx context.Context, req *alertv1.PauseAlertRequest) (*alertv1.Alert, error) {
	user, ok := middleware.GetOAUTH(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	result, err := a.alertRepo.Pause(ctx, user.Subject, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	alert := alertToGrpc(result)
	return &alert, nil
}

func (a *AlertHandler) ResumeAlert(ctx context.Context, req *alertv1.ResumeAlertRequest) (*alertv1.Alert, error) {
	user, ok := middleware.GetOAUTH(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	result, err := a.alertRepo.Enable(ctx, user.Subject, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	alert := alertToGrpc(result)
	return &alert, nil
}

func (a *AlertHandler) DeleteAlert(ctx context.Context, req *alertv1.DeleteAlertRequest) (*alertv1.DeleteAlertResponse, error) {
	user, ok := middleware.GetOAUTH(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	if err := a.alertRepo.Delete(ctx, user.Subject, req.Id); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &alertv1.DeleteAlertResponse{}, nil
}

func (a *AlertHandler) ListAlerts(ctx context.Context, req *alertv1.ListAlertsRequest) (*alertv1.ListAlertsResponse, error) {
	user, ok := middleware.GetOAUTH(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	alerts, err := a.alertRepo.ListByUser(ctx, user.Subject)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	out := []*alertv1.Alert{}
	for _, i := range alerts {
		alert := alertToGrpc(&i)
		out = append(out, &alert)
	}
	return &alertv1.ListAlertsResponse{
		Alerts: out,
	}, nil
}

func conditionToRpc(s string) alertv1.Condition {
	switch s {
	case "above":
		return alertv1.Condition_CONDITION_ABOVE
	case "below":
		return alertv1.Condition_CONDITION_BELOW
	}
	return alertv1.Condition_CONDITION_UNSPECIFIED
}

func conditionString(s alertv1.Condition) string {
	switch s {
	case alertv1.Condition_CONDITION_ABOVE:
		return "above"
	case alertv1.Condition_CONDITION_BELOW:
		return "below"
	}
	return "unspecified"
}

func alertToGrpc(in *domain.Alert) alertv1.Alert {
	var triggeredAt *timestamppb.Timestamp
	if in.TriggeredAt != nil {
		triggeredAt = timestamppb.New(*in.TriggeredAt)
	}
	return alertv1.Alert{
		Id:          in.ID,
		UserId:      in.UserID,
		CoinSymbol:  in.CoinSymbol,
		Condition:   conditionToRpc(in.Condition),
		Price:       in.Price,
		Enabled:     in.Enabled,
		CreatedAt:   timestamppb.New(in.CreatedAt),
		TriggeredAt: triggeredAt,
		UpdatedAt:   timestamppb.New(in.UpdatedAt),
	}
}
