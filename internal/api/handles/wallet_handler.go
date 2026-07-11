package handlers

import (
	"net/http"
	"tracker/internal/api/dto"
	"tracker/internal/core"
	"tracker/internal/db/models"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type WalletHandler struct {
	walletService *core.WalletService
	log           *logrus.Entry
}

func NewWalletHandler(walletService *core.WalletService) *WalletHandler {
	return &WalletHandler{
		walletService: walletService,
		log:           logrus.WithField("component", "WalletHandler"),
	}
}

func (h *WalletHandler) ListWallets(c *gin.Context) {
	wallets, err := h.walletService.ListWallets(c.Request.Context())
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
	wallet := models.Wallet{
		Address: req.Address,
		Chain:   req.Chain,
		Label:   req.Label,
		UserID:  req.UserID,
	}
	createdWallet, err := h.walletService.AddWallet(c.Request.Context(), wallet)
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
