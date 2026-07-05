package handlers

import (
	"context"
	"net/http"
	"strings"
	"tracker/internal/api/dto"
	"tracker/internal/core"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type PriceHandler struct {
	priceService *core.PriceService
}

func NewPriceHandler(priceService *core.PriceService) *PriceHandler {
	return &PriceHandler{
		priceService: priceService,
	}
}

func (h *PriceHandler) GetCoins(c *gin.Context) {
	logrus.WithField("PriceHandler", "GetCoins")
	coins, err := h.priceService.GetCoins(c)
	if err != nil {
		logrus.WithError(err).Warn("failed to get coins")
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	resp := dto.ToCoinResponse(coins)
	c.JSON(http.StatusOK, gin.H{
		"coins": resp,
		"total": len(resp),
	})
}

func (h *PriceHandler) GetPrices(c *gin.Context) {
	logrus.WithField("PriceHandler", "GetPrices")
	symbolsParam := c.Query("symbols")
	var symbols []string
	if symbolsParam != "" {
		symbols = strings.Split(strings.ToUpper(symbolsParam), ",")
	}
	prices, err := h.priceService.GetPrices(context.Background(), symbols)
	if err != nil {
		logrus.WithError(err).Warn("failed to get prices")
		return
	}
	resp := dto.ToPriceResponse(prices)
	c.JSON(http.StatusOK, gin.H{
		"prices": resp,
		"total":  len(resp),
	})
}
