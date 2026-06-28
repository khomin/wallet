package handlers

import (
	"tracker/internal/core"

	"github.com/gin-gonic/gin"
)

type PriceHandler struct {
	priceService *core.PriceService
}

func NewPriceHandler(priceService *core.PriceService) *PriceHandler {
	return &PriceHandler{
		priceService: priceService,
	}
}

func (h *PriceHandler) GetPrices(c *gin.Context) {
	// h.priceService.GetPrices()
	// 1. Parse HTTP request (symbols from query params)
	// 2. Call h.priceService.GetPrices()
	// 3. Format HTTP response
	// 4. Return JSON
}
