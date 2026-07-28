package handlers

import (
	"errors"
	"net/http"
	"tracker/internal/api/dto"
	"tracker/internal/api/middleware"
	"tracker/internal/core"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

func (h *WalletHandler) GetWallet(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		dto.InvalidParametersMessage(c, "id parameter is required")
		return
	}
	uuid, err := uuid.Parse(id)
	if err != nil {
		dto.InvalidParametersMessage(c, "id parameter is required")
		return
	}
	user, ok := middleware.GetOAUTH(c)
	if !ok {
		dto.UnauthorizedError(c)
		return
	}
	wallet, err := h.walletService.GetWallet(c.Request.Context(), user.Subject, uuid)
	if err != nil {
		if errors.Is(err, core.ErrWalletNotFound) {
			dto.NotFoundErrorMessage(c, "wallet not found")
			return
		}
		dto.InternallError(c)
		return
	}
	c.JSON(http.StatusOK, dto.ToWalletResponse(&wallet))
}

func (h *WalletHandler) ListWallets(c *gin.Context) {
	user, ok := middleware.GetOAUTH(c)
	if !ok {
		dto.UnauthorizedError(c)
		return
	}
	wallet, err := h.walletService.ListWallets(c.Request.Context(), user.Subject)
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
	user, ok := middleware.GetOAUTH(c)
	if !ok {
		dto.UnauthorizedError(c)
		return
	}
	wallet, err := h.walletService.AddWallet(c.Request.Context(), user.Subject, req.Chain, req.Address, req.TokenSymbol, req.Label)
	if err != nil {
		if errors.Is(err, core.ErrWalletAlreadyExists) {
			dto.AlreadyExistsError(c)
			return
		}
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
	user, ok := middleware.GetOAUTH(c)
	if !ok {
		dto.UnauthorizedError(c)
		return
	}
	err := h.walletService.DeleteWallet(c.Request.Context(), user.Subject, req.ID)
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

func (h *WalletHandler) EditWallet(c *gin.Context) {
	var req dto.EditWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.InvalidParameters(c)
		return
	}
	user, ok := middleware.GetOAUTH(c)
	if !ok {
		dto.UnauthorizedError(c)
		return
	}
	wallet, err := h.walletService.EditWallet(c.Request.Context(), user.Subject, req.ID, req.Label)
	if err != nil {
		if errors.Is(err, core.ErrWalletNotFound) {
			dto.NotFoundErrorMessage(c, "wallet not found")
			return
		}
		dto.InternallError(c)
		return
	}
	c.JSON(http.StatusOK, dto.EditWalletResponse{
		WalletResponse: dto.ToWalletResponse(wallet),
	})
}
