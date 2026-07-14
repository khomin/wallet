package handlers

import (
	"errors"
	"net/http"
	"tracker/internal/api/dto"
	"tracker/internal/api/middleware"
	"tracker/internal/core"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type WalletHandler struct {
	walletService     *core.WalletService
	blockchainService *core.BlockchainService
	log               *logrus.Entry
}

func NewWalletHandler(
	walletService *core.WalletService,
	blockchainService *core.BlockchainService,
) *WalletHandler {
	return &WalletHandler{
		walletService:     walletService,
		blockchainService: blockchainService,
		log:               logrus.WithField("component", "WalletHandler"),
	}
}

func (h *WalletHandler) ListWallets(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "unauthorized",
		})
		return
	}
	wallets, err := h.walletService.ListWallets(c.Request.Context(), userID)
	if err != nil {
		h.log.WithError(err).Error("failed to list wallets")
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "unexpected error",
		})
		return
	}
	resp := dto.ToWalletResponses(wallets)
	c.JSON(http.StatusOK, gin.H{
		"wallets": resp,
		"total":   len(resp),
	})
}

func (h *WalletHandler) AddWallet(c *gin.Context) {
	var req dto.CreateWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "invalid request payload",
		})
		return
	}
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "unauthorized",
		})
		return
	}
	createdWallet, err := h.walletService.AddWallet(c.Request.Context(), userID, req.Chain, req.Address, req.Label)
	if err != nil {
		h.log.WithError(err).Error("failed to add wallet")
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "unexpected error",
		})
		return
	}
	c.JSON(http.StatusCreated, dto.ToWalletResponse(*createdWallet))
}

func (h *WalletHandler) DeleteWallet(c *gin.Context) {
	var req dto.DeleteWalletRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "invalid request payload",
		})
		return
	}
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "unauthorized",
		})
		return
	}
	err := h.walletService.DeleteWallet(c.Request.Context(), userID, req.ID)
	if err != nil {
		h.log.WithError(err).Error("failed to delete wallet")
		if errors.Is(err, core.ErrWalletNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Code:    "NOT_FOUND",
				Message: "wallet not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "unexpected error",
		})
		return
	}
	c.JSON(http.StatusOK, dto.DeleteWalletResponse{
		ID: req.ID,
	})
}

func (h *WalletHandler) GetWalletBalance(c *gin.Context) {
	var req dto.GetWalletBalanceRequest
	if err := c.ShouldBind(&req); err != nil {
		// chain := c.Param("chain")
		// address := c.Param("address")
		// if chain != "" && address != "" {
		// 	req = dto.GetWalletBalanceRequest{Address: address, Chain: chain}
		// } else {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "invalid request payload",
		})
		return
		// }
	}
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "unauthorized",
		})
		return
	}
	balance, err := h.blockchainService.GetBalance(c.Request.Context(), userID, req.ID)
	if err != nil {
		h.log.WithError(err).Error("failed to get wallet balance")
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "unexpected error",
		})
		return
	}
	c.JSON(http.StatusOK, dto.ToGetWalletBalanceResponse(balance))
}
