package handlers

import (
	"errors"
	"net/http"
	"strings"
	"tracker/internal/api/dto"
	"tracker/internal/core"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type PriceHandler struct {
	priceService *core.PriceService
	log          *logrus.Entry
}

func NewPriceHandler(priceService *core.PriceService) *PriceHandler {
	return &PriceHandler{
		priceService: priceService,
		log:          logrus.WithField("component", "PriceHandler"),
	}
}

func (h *PriceHandler) GetCoins(c *gin.Context) {
	coins, err := h.priceService.GetCoins(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "unexpected error",
		})
		return
	}
	resp := dto.ToCoinsResponse(coins)
	c.JSON(http.StatusOK, gin.H{
		"coins": resp,
		"total": len(resp),
	})
}

func (h *PriceHandler) GetCoin(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "id parameter is required",
		})
		return
	}
	coin, err := h.priceService.GetCoin(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "unexpected error",
		})
		return
	}
	resp := dto.ToCoinResponse(*coin)
	c.JSON(http.StatusOK, resp)
}

func (h *PriceHandler) GetPrices(c *gin.Context) {
	symbolsParam := c.Query("symbols")
	symbols := strings.Split(strings.ToLower(symbolsParam), ",")
	if len(symbols) == 0 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "symbols parameter is required",
		})
		return
	}
	prices, err := h.priceService.GetPrices(c.Request.Context(), symbols)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "unexpected error",
		})
		return
	}
	resp := dto.ToPricesResponse(prices)
	c.JSON(http.StatusOK, gin.H{
		"prices": resp,
		"total":  len(resp),
	})
}

func (h *PriceHandler) GetPrice(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "id parameter is required",
		})
		return
	}
	price, err := h.priceService.GetPrice(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, core.ErrPriceNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Code:    "NOT_FOUND",
				Message: "requested price not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "unexpected error",
		})
		return
	}
	resp := dto.ToPriceResponse(*price)
	c.JSON(http.StatusOK, resp)
}
