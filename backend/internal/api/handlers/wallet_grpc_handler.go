package handlers

import (
	"context"
	walletv1 "tracker/gen/wallet/v1"
	"tracker/internal/core"
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

func (s *WalletGrpcHandler) GetBalance(ctx context.Context, req *walletv1.GetBalanceRequest) (*walletv1.GetBalanceResponse, error) {
	// Call your core service logic
	// balance, err := s.walletService.GetBalance(ctx, req.Chain, req.Address, req.Symbol)
	// if err != nil {
	// 	return &walletv1.GetBalanceResponse{
	// 		Chain:    req.Chain,
	// 		Address:  req.Address,
	// 		HasError: true,
	// 		ErrorMsg: err.Error(),
	// 	}, nil
	// }

	return &walletv1.GetBalanceResponse{
		// Chain:         balance.Chain,
		// Address:       balance.Address,
		// BalanceCrypto: balance.Balance,
		// HasError:      false,
	}, nil
}
