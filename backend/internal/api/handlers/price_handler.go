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
		dto.InternallError(c)
		return
	}
	c.JSON(http.StatusOK, dto.ToCoinsResponse(coins))
}

func (h *PriceHandler) GetCoin(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		dto.InvalidParametersMessage(c, "id parameter is required")
		return
	}
	coin, err := h.priceService.GetCoin(c, id)
	if err != nil {
		dto.InternallError(c)
		return
	}
	c.JSON(http.StatusOK, dto.ToCoinResponse(coin))
}

func (h *PriceHandler) GetPrices(c *gin.Context) {
	symbolsParam := c.Query("symbols")
	symbols := strings.Split(strings.ToLower(symbolsParam), ",")
	if len(symbols) == 0 {
		dto.InvalidParametersMessage(c, "symbols parameter is required")
		return
	}
	prices, err := h.priceService.GetPrices(c.Request.Context(), symbols)
	if err != nil {
		dto.InternallError(c)
		return
	}
	c.JSON(http.StatusOK, dto.ToPricesResponse(prices))
}

func (h *PriceHandler) GetPrice(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		dto.InvalidParametersMessage(c, "id parameter is required")
		return
	}
	price, err := h.priceService.GetPrice(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, core.ErrPriceNotFound) {
			dto.NotFoundErrorMessage(c, "requested price not found")
			return
		}
		dto.InternallError(c)
		return
	}
	resp := dto.ToPriceResponse(price)
	c.JSON(http.StatusOK, resp)
}
