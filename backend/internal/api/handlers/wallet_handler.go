package handlers

import (
	"context"
	"errors"
	pricev1 "tracker/gen/price/v1"
	walletv1 "tracker/gen/wallet/v1"
	"tracker/internal/api/middleware"
	"tracker/internal/core"

	"github.com/google/uuid"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type WalletGrpcHandler struct {
	walletv1.UnimplementedWalletServiceServer
	walletService *core.WalletService
}

func NewWalletGrpcHandler(walletService *core.WalletService) *WalletGrpcHandler {
	return &WalletGrpcHandler{
		walletService: walletService,
	}
}

func (s *WalletGrpcHandler) GetWallets(ctx context.Context, req *walletv1.GetWalletsReq) (*walletv1.GetWalletsResp, error) {
	user, ok := middleware.GetOAUTH(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	wallets, err := s.walletService.ListWallets(ctx, user.Subject)
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
	return &walletv1.GetWalletsResp{
		Total:  int32(len(wallets)),
		Wallet: out,
	}, nil
}

func (s *WalletGrpcHandler) GetWallet(ctx context.Context, req *walletv1.GetWalletReq) (*walletv1.GetWalletResp, error) {
	user, ok := middleware.GetOAUTH(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	uuid, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "id parameter is required")
	}
	wallet, err := s.walletService.GetWallet(ctx, user.Subject, uuid)
	if err != nil {
		if errors.Is(err, core.ErrWalletNotFound) {
			return nil, status.Error(codes.NotFound, "wallet not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &walletv1.GetWalletResp{
		Wallet: &walletv1.Wallet{
			Id:            wallet.Wallet.ID,
			Address:       wallet.Wallet.Address,
			Chain:         wallet.Wallet.Chain,
			TokenSymbol:   wallet.Wallet.Symbol,
			Label:         wallet.Wallet.Label,
			BalanceCrypto: float32(wallet.Balance),
			BalanceUsd:    float32(wallet.BalanceUSD),
			HasError:      wallet.HasError,
			ErrorMsg:      wallet.ErrorMsg,
			Price: &pricev1.Price{
				Symbol:                        wallet.Symbol,
				Name:                          wallet.Price.Name,
				PriceUsd:                      float32(wallet.Price.CurrentPrice),
				MarketCap:                     float32(wallet.Price.MarketCap),
				TotalVolume:                   float32(wallet.Price.TotalVolume),
				High_24H:                      float32(wallet.Price.High_24h),
				Low_24H:                       float32(wallet.Price.Low_24h),
				PriceChange_24H:               float32(wallet.Price.PriceChange_24h),
				PriceChangePercentage_24H:     float32(wallet.Price.PriceChangePercentage_24h),
				MarketCapChange_24H:           float32(wallet.Price.MarketCapChange_24h),
				MarketCapChangePercentage_24H: float32(wallet.Price.MarketCapChange_percentage_24h),
				UpdatedAt:                     timestamppb.New(wallet.Price.UpdatedAt),
			},
		},
	}, nil
}

func (s *WalletGrpcHandler) AddWallet(ctx context.Context, req *walletv1.AddWalletReq) (*emptypb.Empty, error) {
	user, ok := middleware.GetOAUTH(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	err := s.walletService.AddWallet(ctx, user.Subject, req.Chain, req.Address, req.TokenSymbol, req.Label)
	if err != nil {
		if errors.Is(err, core.ErrWalletNotFound) {
			return nil, status.Error(codes.NotFound, "wallet not found")
		} else if errors.Is(err, core.ErrWalletAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "wallet already exists")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	// return &walletv1.AddWalletResp{
	// 	Wallet: &walletv1.Wallet{
	// 		Id:                wallet.Wallet.ID,
	// 		Address:           wallet.Wallet.Address,
	// 		Chain:             wallet.Wallet.Chain,
	// 		TokenSymbol:       wallet.Wallet.Symbol,
	// 		Label:             wallet.Wallet.Label,
	// 		BalanceCrypto:     float32(wallet.Balance),
	// 		BalanceUsd:        float32(wallet.BalanceUSD),
	// 		Change_24HPercent: float32(wallet.Price.Change_24h),
	// 		HasError:          wallet.HasError,
	// 		ErrorMsg:          wallet.ErrorMsg,
	// 	},
	// }, nil
	return &emptypb.Empty{}, nil
}

func (s *WalletGrpcHandler) EditWallet(ctx context.Context, req *walletv1.EditWalletReq) (*walletv1.EditWalletResp, error) {
	user, ok := middleware.GetOAUTH(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	uuid, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "id parameter is required")
	}
	wallet, err := s.walletService.EditWallet(ctx, user.Subject, uuid, req.Label)
	if err != nil {
		if errors.Is(err, core.ErrWalletNotFound) {
			return nil, status.Error(codes.NotFound, "wallet not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &walletv1.EditWalletResp{
		Id:    wallet.ID,
		Label: wallet.Label,
	}, nil
}

func (s *WalletGrpcHandler) DeleteWallet(ctx context.Context, req *walletv1.DeleteWalletReq) (*walletv1.DeleteWalletResp, error) {
	user, ok := middleware.GetOAUTH(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	uuid, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "id parameter is required")
	}
	err = s.walletService.DeleteWallet(ctx, user.Subject, uuid)
	if err != nil {
		if errors.Is(err, core.ErrWalletNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &walletv1.DeleteWalletResp{
		DeletedId: uuid.String(),
	}, nil
}
