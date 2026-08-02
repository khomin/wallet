package handlers

import (
	"context"
	walletv1 "tracker/gen/wallet/v1"
	"tracker/internal/api/middleware"
	"tracker/internal/core"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

// func (s *WalletGrpcHandler) GetBalance(ctx context.Context, req *walletv1.GetBalanceRequest) (*walletv1.GetBalanceResponse, error) {
// 	// Call your core service logic
// 	// balance, err := s.walletService.GetBalance(ctx, req.Chain, req.Address, req.Symbol)
// 	// if err != nil {
// 	// 	return &walletv1.GetBalanceResponse{
// 	// 		Chain:    req.Chain,
// 	// 		Address:  req.Address,
// 	// 		HasError: true,
// 	// 		ErrorMsg: err.Error(),
// 	// 	}, nil
// 	// }

// 	return &walletv1.GetBalanceResponse{
// 		// Chain:         balance.Chain,
// 		// Address:       balance.Address,
// 		// BalanceCrypto: balance.Balance,
// 		// HasError:      false,
// 	}, nil
// }

func (s *WalletGrpcHandler) GetWallets(ctx context.Context, req *walletv1.GetWalletsReq) (*walletv1.GetWalletsResp, error) {
	// user, ok := middleware.GetOAUTH(c)
	// if !ok {
	// 	dto.UnauthorizedError(c)
	// 	return
	// }
	user, ok := ctx.Value("user").(*middleware.JwtCustomClaims)
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
			Id:                i.Wallet.ID,
			Address:           i.Wallet.Address,
			Chain:             i.Wallet.Chain,
			TokenSymbol:       i.Wallet.Symbol,
			Label:             i.Wallet.Label,
			BalanceCrypto:     float32(i.Balance),
			BalanceUsd:        float32(i.Balance),
			Change_24HPercent: float32(i.Price.Change_24h),
			HasError:          i.HasError,
			ErrorMsg:          i.ErrorMsg,
		})
	}
	return &walletv1.GetWalletsResp{
		Total:  int64(len(wallets)),
		Wallet: out,
	}, nil
}

func (s *WalletGrpcHandler) GetWallet(ctx context.Context, req *walletv1.GetWalletReq) (*walletv1.GetWalletResp, error) {
	return nil, status.Error(codes.Unimplemented, "method GetWallet not implemented")
}
func (s *WalletGrpcHandler) EditWallet(ctx context.Context, req *walletv1.EditWalletReq) (*walletv1.EditWalletResp, error) {
	return nil, status.Error(codes.Unimplemented, "method EditWallet not implemented")
}
func (s *WalletGrpcHandler) DeleteWallet(ctx context.Context, req *walletv1.DeleteWalletReq) (*walletv1.DeleteWalletResp, error) {
	return nil, status.Error(codes.Unimplemented, "method DeleteWallet not implemented")
}
func (s *WalletGrpcHandler) AddWallet(ctx context.Context, req *walletv1.AddWalletReq) (*walletv1.AddWalletResp, error) {
	return nil, status.Error(codes.Unimplemented, "method AddWallet not implemented")
}
