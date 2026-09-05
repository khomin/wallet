package handlers

import (
	"context"
	"errors"
	"time"
	walletv1 "tracker/gen/wallet/v1"
	"tracker/internal/api/middleware"
	"tracker/internal/core"
	"tracker/internal/core/domain"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type WalletGrpcHandler struct {
	walletService *core.WalletService
	walletWorker  *core.WalletWorker
}

func NewWalletGrpcHandler(walletService *core.WalletService, walletWorker *core.WalletWorker) walletv1.WalletServiceServer {
	return &WalletGrpcHandler{
		walletService: walletService,
		walletWorker:  walletWorker,
	}
}

func (s *WalletGrpcHandler) ListWallets(ctx context.Context, req *walletv1.ListWalletsRequest) (*walletv1.ListWalletsResponse, error) {
	user, ok := middleware.GetUser(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	wallets, err := s.walletService.ListWallets(ctx, user)
	if err != nil {
		return nil, err
	}
	out := make([]*walletv1.Wallet, 0, len(wallets))
	for _, i := range wallets {
		out = append(out, &walletv1.Wallet{
			Id:            i.Wallet.ID,
			Address:       i.Wallet.Address,
			Chain:         i.Wallet.Chain,
			TokenSymbol:   i.Wallet.Symbol,
			Label:         i.Wallet.Label,
			BalanceCrypto: i.Balance,
			BalanceUsd:    i.BalanceUSD,
			HasError:      i.HasError,
			ErrorMsg:      i.ErrorMsg,
			Price:         i.Price.ToGrpc(),
		})
	}
	return &walletv1.ListWalletsResponse{
		Total:  int32(len(wallets)),
		Wallet: out,
	}, nil
}

func (s *WalletGrpcHandler) GetWallet(ctx context.Context, req *walletv1.GetWalletRequest) (*walletv1.GetWalletResponse, error) {
	user, ok := middleware.GetUser(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	uuid, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "id parameter is required")
	}
	wallet, err := s.walletService.GetWallet(ctx, user, uuid)
	if err != nil {
		if errors.Is(err, domain.ErrorNotFound) {
			return nil, status.Error(codes.NotFound, "wallet not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &walletv1.GetWalletResponse{
		Wallet: wallet.ToGrpc(),
	}, nil
}

func (s *WalletGrpcHandler) CreateWallet(ctx context.Context, req *walletv1.CreateWalletRequest) (*walletv1.CreateWalletResponse, error) {
	user, ok := middleware.GetUser(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	err := s.walletService.CreateWallet(ctx, user, req.Chain, req.Address, req.TokenSymbol, req.Label)
	if err != nil {
		if errors.Is(err, domain.ErrorNotFound) {
			return nil, status.Error(codes.NotFound, "wallet not found")
		} else if errors.Is(err, domain.ErrorWWalletAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "wallet already exists")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &walletv1.CreateWalletResponse{}, nil
}

func (s *WalletGrpcHandler) UpdateWallet(ctx context.Context, req *walletv1.UpdateWalletRequest) (*walletv1.UpdateWalletResponse, error) {
	user, ok := middleware.GetUser(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	uuid, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "id parameter is required")
	}
	wallet, err := s.walletService.UpdateWallet(ctx, user, uuid, req.Label)
	if err != nil {
		if errors.Is(err, domain.ErrorNotFound) {
			return nil, status.Error(codes.NotFound, "wallet not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &walletv1.UpdateWalletResponse{
		Id:    wallet.ID,
		Label: wallet.Label,
	}, nil
}

func (s *WalletGrpcHandler) DeleteWallet(ctx context.Context, req *walletv1.DeleteWalletRequest) (*walletv1.DeleteWalletResponse, error) {
	user, ok := middleware.GetUser(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	uuid, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "id parameter is required")
	}
	err = s.walletService.DeleteWallet(ctx, user, uuid)
	if err != nil {
		if errors.Is(err, domain.ErrorNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &walletv1.DeleteWalletResponse{
		DeletedId: uuid.String(),
	}, nil
}

func (s *WalletGrpcHandler) ListWalletBalances(ctx context.Context, req *walletv1.ListWalletBalancesRequest) (*walletv1.ListWalletBalancesResponse, error) {
	user, ok := middleware.GetUser(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	uuid, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}
	if req.GetLimit() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "limit must be greater than zero")
	}
	now := time.Now()
	from, err := getBalancePeriodFrom(req.GetPeriod(), now)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid period")
	}
	snapshots, err := s.walletService.GetBalanceSnapshot(
		ctx,
		user,
		uuid,
		core.BalanceSnapshotFilter{
			From:  from,
			To:    now,
			Limit: int(req.GetLimit()),
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get wallet balances")
	}
	balances := make([]*walletv1.WalletBalance, 0, len(snapshots))
	for _, snapshot := range snapshots {
		balances = append(balances, &walletv1.WalletBalance{
			BalanceCrypto: snapshot.Balance,
			BalanceUsd:    snapshot.BalanceUSD,
			Time:          timestamppb.New(snapshot.Time),
		})
	}
	return &walletv1.ListWalletBalancesResponse{
		Balance: balances,
	}, nil
}

func (s *WalletGrpcHandler) StreamWallet(
	req *walletv1.StreamWalletRequest,
	stream grpc.ServerStreamingServer[walletv1.WalletUpdate],
) error {
	ctx := stream.Context()
	user, ok := middleware.GetUser(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "unauthorized")
	}
	channel, streamID := s.walletWorker.Subscribe(user.ID)
	defer s.walletWorker.UnSubscribe(user.ID, streamID)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event := <-channel:
			logrus.Infof("received event: %v", event)
			if err := stream.Send(event); err != nil {
				return err
			}
		}
	}
}

func getBalancePeriodFrom(period walletv1.BalancePeriod, now time.Time) (time.Time, error) {
	switch period {
	case walletv1.BalancePeriod_BALANCE_PERIOD_1D:
		return now.Add(-24 * time.Hour), nil
	case walletv1.BalancePeriod_BALANCE_PERIOD_1W:
		return now.AddDate(0, 0, -7), nil
	case walletv1.BalancePeriod_BALANCE_PERIOD_1M:
		return now.AddDate(0, -1, 0), nil
	case walletv1.BalancePeriod_BALANCE_PERIOD_6M:
		return now.AddDate(0, -6, 0), nil
	case walletv1.BalancePeriod_BALANCE_PERIOD_1Y:
		return now.AddDate(-1, 0, 0), nil
	case walletv1.BalancePeriod_BALANCE_PERIOD_5Y:
		return now.AddDate(-5, 0, 0), nil
	case walletv1.BalancePeriod_BALANCE_PERIOD_ALL:
		return time.Time{}, nil
	default:
		return time.Time{}, errors.New("invalid balance period")
	}
}
