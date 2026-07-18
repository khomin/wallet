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
	walletService *core.WalletService
	log           *logrus.Entry
}

func NewWalletHandler(
	walletService *core.WalletService,
) *WalletHandler {
	return &WalletHandler{
		walletService: walletService,
		log:           logrus.WithField("component", "WalletHandler"),
	}
}

func (h *WalletHandler) ListWallets(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		dto.UnauthorizedError(c)
		return
	}
	wallet, err := h.walletService.ListWallets(c.Request.Context(), userID)
	if err != nil {
		dto.InternallError(c)
		return
	}
	c.JSON(http.StatusOK, dto.ToWalletResponses(wallet))
}

func (h *WalletHandler) AddWallet(c *gin.Context) {
	var req dto.CreateWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.InvalidParameters(c)
		return
	}
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		dto.UnauthorizedError(c)
		return
	}
	wallet, err := h.walletService.AddWallet(c.Request.Context(), userID, req.Chain, req.Address, req.Symbol, req.Label)
	if err != nil {
		dto.InternallError(c)
		return
	}
	c.JSON(http.StatusCreated, dto.ToWalletResponse(wallet))
}

func (h *WalletHandler) DeleteWallet(c *gin.Context) {
	var req dto.DeleteWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.InvalidParameters(c)
		return
	}
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		dto.UnauthorizedError(c)
		return
	}
	err := h.walletService.DeleteWallet(c.Request.Context(), userID, req.ID)
	if err != nil {
		if errors.Is(err, core.ErrWalletNotFound) {
			dto.NotFoundErrorMessage(c, "wallet not found")
			return
		}
		dto.InternallError(c)
		return
	}
	c.JSON(http.StatusOK, dto.DeleteWalletResponse{
		ID: req.ID,
	})
}

func (h *WalletHandler) GetWalletBalance(c *gin.Context) {
	var req dto.GetWalletBalanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.InvalidParameters(c)
		return
	}
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		dto.UnauthorizedError(c)
		return
	}
	wallet, err := h.walletService.GetWallet(c.Request.Context(), userID, req.ID)
	if err != nil {
		dto.InternallError(c)
		return
	}
	c.JSON(http.StatusOK, dto.ToGetWalletBalanceResponse(wallet))
}
