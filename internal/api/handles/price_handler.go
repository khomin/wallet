package handlers

import (
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
	resp := dto.ToCoinsResponse(coins)
	c.JSON(http.StatusOK, gin.H{
		"coins": resp,
		"total": len(resp),
	})
}

func (h *PriceHandler) GetCoin(c *gin.Context) {
	logrus.WithField("PriceHandler", "GetCoin")
	id := c.Param("id")
	if id == "" {
		logrus.Warning("empty id")
		c.JSON(http.StatusBadRequest, nil)
	}
	coin, err := h.priceService.GetCoin(c, id)
	if err != nil {
		logrus.WithError(err).Warn("failed to get coin")
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	resp := dto.ToCoinResponse(*coin)
	c.JSON(http.StatusOK, resp)
}

func (h *PriceHandler) GetPrices(c *gin.Context) {
	log := logrus.WithField("PriceHandler", "GetPrices")
	symbolsParam := c.Query("symbols")
	var symbols []string
	if symbolsParam != "" {
		symbols = strings.Split(strings.ToUpper(symbolsParam), ",")
	}
	prices, err := h.priceService.GetPrices(c.Request.Context(), symbols)
	if err != nil {
		log.WithError(err).Warn("failed to get prices")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_server_error"})
		return
	}
	resp := dto.ToPriceResponse(prices)
	c.JSON(http.StatusOK, gin.H{
		"prices": resp,
		"total":  len(resp),
	})
}
