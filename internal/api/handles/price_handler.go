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
		h.log.WithError(err).Warn("failed to get coins")
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
	id := c.Param("id")
	if id == "" {
		h.log.Warning("empty id")
		c.JSON(http.StatusBadRequest, nil)
	}
	coin, err := h.priceService.GetCoinSnapshot(c, id)
	if err != nil {
		h.log.WithError(err).Warn("failed to get coin")
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	resp := dto.ToCoinResponse(*coin)
	c.JSON(http.StatusOK, resp)
}

func (h *PriceHandler) GetPrices(c *gin.Context) {
	symbolsParam := c.Query("symbols")
	var symbols []string
	if symbolsParam != "" {
		symbols = strings.Split(strings.ToUpper(symbolsParam), ",")
	}
	prices, err := h.priceService.GetPrices(c.Request.Context(), symbols)
	if err != nil {
		h.log.WithError(err).Warn("failed to get prices")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_server_error"})
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
		h.log.Warning("empty id")
		c.JSON(http.StatusBadRequest, nil)
	}
	prices, err := h.priceService.GetPrices(c.Request.Context(), []string{id})
	if err != nil {
		h.log.WithError(err).Warn("failed to get price")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_server_error"})
		return
	}
	if len(prices) == 0 {
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}
	resp := dto.ToPriceResponse(prices[0])
	c.JSON(http.StatusOK, resp)
}
