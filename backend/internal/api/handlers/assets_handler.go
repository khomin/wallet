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

type AssetsHandler struct {
	priceService *core.AssetService
	log          *logrus.Entry
}

func NewAssetsHandler(
	assetService *core.AssetService,
) *AssetsHandler {
	return &AssetsHandler{
		priceService: assetService,
		log:          logrus.WithField("component", "AssetsHandler"),
	}
}

func (h *AssetsHandler) GetCoins(c *gin.Context) {
	coins, err := h.priceService.GetCoins(c)
	if err != nil {
		dto.InternallError(c)
		return
	}
	c.JSON(http.StatusOK, dto.CoinsResponse{
		Total: len(coins),
		Coins: coins,
	})
}

func (h *AssetsHandler) GetCoin(c *gin.Context) {
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
	c.JSON(http.StatusOK, coin)
}

func (h *AssetsHandler) SearchCoin(c *gin.Context) {
	var req dto.SearchCoins
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.InvalidParameters(c)
		return
	}
	tokens, err := h.priceService.SearchCoins(c.Request.Context(), req.Text)
	if err != nil {
		dto.InternallError(c)
		return
	}
	c.JSON(http.StatusOK, tokens)
}

func (h *AssetsHandler) GetPrices(c *gin.Context) {
	symbolsParam := c.Query("symbols")
	symbols := strings.Split(strings.ToUpper(symbolsParam), ",")
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

func (h *AssetsHandler) GetPrice(c *gin.Context) {
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
