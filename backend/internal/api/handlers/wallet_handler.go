package handlers

import (
	"context"
	"errors"
	pricev1 "tracker/gen/price/v1"
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
	walletv1.UnimplementedWalletServiceServer
}

func NewWalletGrpcHandler(walletService *core.WalletService, walletWorker *core.WalletWorker) walletv1.WalletServiceServer {
	return &WalletGrpcHandler{
		walletService: walletService,
		walletWorker:  walletWorker,
	}
}

func (s *WalletGrpcHandler) ListWallets(ctx context.Context, req *walletv1.ListWalletsRequest) (*walletv1.ListWalletsResponse, error) {
	user, ok := middleware.GetOAUTH(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	wallets, err := s.walletService.ListWallets(ctx, domain.User{ID: user.Subject, Name: user.Name, Email: user.Email})
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
			BalanceCrypto: float32(i.Balance),
			BalanceUsd:    float32(i.Balance),
			HasError:      i.HasError,
			ErrorMsg:      i.ErrorMsg,
			Price: &pricev1.Price{
				Symbol:                        i.Symbol,
				Name:                          i.Price.Name,
				PriceUsd:                      float32(i.Price.CurrentPrice),
				MarketCap:                     float32(i.Price.MarketCap),
				TotalVolume:                   float32(i.Price.TotalVolume),
				High_24H:                      float32(i.Price.High_24h),
				Low_24H:                       float32(i.Price.Low_24h),
				PriceChange_24H:               float32(i.Price.PriceChange_24h),
				PriceChangePercentage_24H:     float32(i.Price.PriceChangePercentage_24h),
				MarketCapChange_24H:           float32(i.Price.MarketCapChange_24h),
				MarketCapChangePercentage_24H: float32(i.Price.MarketCapChange_percentage_24h),
				UpdatedAt:                     timestamppb.New(i.Price.UpdatedAt),
			},
		})
	}
	return &walletv1.ListWalletsResponse{
		Total:  int32(len(wallets)),
		Wallet: out,
	}, nil
}

func (s *WalletGrpcHandler) GetWallet(ctx context.Context, req *walletv1.GetWalletRequest) (*walletv1.GetWalletResponse, error) {
	user, ok := middleware.GetOAUTH(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	uuid, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "id parameter is required")
	}
	wallet, err := s.walletService.GetWallet(ctx, domain.User{ID: user.Subject, Name: user.Name, Email: user.Email}, uuid)
	if err != nil {
		if errors.Is(err, domain.ErrWalletNotFound) {
			return nil, status.Error(codes.NotFound, "wallet not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &walletv1.GetWalletResponse{
		Wallet: wallet.ToGrpc(),
	}, nil
}

func (s *WalletGrpcHandler) CreateWallet(ctx context.Context, req *walletv1.CreateWalletRequest) (*walletv1.CreateWalletResponse, error) {
	user, ok := middleware.GetOAUTH(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	err := s.walletService.CreateWallet(ctx, domain.User{ID: user.Subject, Name: user.Name, Email: user.Email}, req.Chain, req.Address, req.TokenSymbol, req.Label)
	if err != nil {
		if errors.Is(err, domain.ErrWalletNotFound) {
			return nil, status.Error(codes.NotFound, "wallet not found")
		} else if errors.Is(err, domain.ErrWalletAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "wallet already exists")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &walletv1.CreateWalletResponse{}, nil
}

func (s *WalletGrpcHandler) EditWallet(ctx context.Context, req *walletv1.EditWalletRequest) (*walletv1.EditWalletResponse, error) {
	user, ok := middleware.GetOAUTH(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	uuid, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "id parameter is required")
	}
	wallet, err := s.walletService.EditWallet(ctx, domain.User{ID: user.Subject, Name: user.Name, Email: user.Email}, uuid, req.Label)
	if err != nil {
		if errors.Is(err, domain.ErrWalletNotFound) {
			return nil, status.Error(codes.NotFound, "wallet not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &walletv1.EditWalletResponse{
		Id:    wallet.ID,
		Label: wallet.Label,
	}, nil
}

func (s *WalletGrpcHandler) DeleteWallet(ctx context.Context, req *walletv1.DeleteWalletRequest) (*walletv1.DeleteWalletResponse, error) {
	user, ok := middleware.GetOAUTH(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	uuid, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "id parameter is required")
	}
	err = s.walletService.DeleteWallet(ctx, domain.User{ID: user.Subject, Name: user.Name, Email: user.Email}, uuid)
	if err != nil {
		if errors.Is(err, domain.ErrWalletNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &walletv1.DeleteWalletResponse{
		DeletedId: uuid.String(),
	}, nil
}

func (s *WalletGrpcHandler) StreamWallet(
	req *walletv1.StreamWalletRequest,
	stream grpc.ServerStreamingServer[walletv1.WalletUpdate],
) error {
	ctx := stream.Context()
	user, ok := middleware.GetOAUTH(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "unauthorized")
	}
	channel, streamID := s.walletWorker.Subscribe(user.Subject)
	defer s.walletWorker.UnSubscribe(user.Subject, streamID)
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
