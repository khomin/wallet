package handlers

// type PriceHandler struct {
// 	priceService *core.PriceService
// 	log          *logrus.Entry
// }

// func NewPriceHandler(
// 	service *core.PriceService,
// ) *PriceGrpcHandler {
// 	return &PriceGrpcHandler{
// 		priceService: service,
// 		log:          logrus.WithField("component", "PriceHandler"),
// 	}
// }

// func (h *PriceGrpcHandler) GetCoins(c *gin.Context) {
// 	coins, err := h.priceService.GetCoins(c)
// 	if err != nil {
// 		dto.InternallError(c)
// 		return
// 	}
// 	c.JSON(http.StatusOK, dto.CoinsResponse{
// 		Total: len(coins),
// 		Coins: coins,
// 	})
// }

// func (h *PriceGrpcHandler) GetCoin(c *gin.Context) {
// 	id := c.Param("id")
// 	if id == "" {
// 		dto.InvalidParametersMessage(c, "id parameter is required")
// 		return
// 	}
// 	coin, err := h.priceService.GetCoin(c, id)
// 	if err != nil {
// 		dto.InternallError(c)
// 		return
// 	}
// 	c.JSON(http.StatusOK, coin)
// }

// func (h *PriceGrpcHandler) SearchCoin(c *gin.Context) {
// 	var req dto.SearchCoins
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		dto.InvalidParameters(c)
// 		return
// 	}
// 	tokens, err := h.priceService.SearchCoins(c.Request.Context(), req.Text)
// 	if err != nil {
// 		dto.InternallError(c)
// 		return
// 	}
// 	c.JSON(http.StatusOK, tokens)
// }

// func (h *PriceGrpcHandler) GetPrices(c *gin.Context) {
// 	symbolsParam := c.Query("symbols")
// 	symbols := strings.Split(strings.ToUpper(symbolsParam), ",")
// 	if len(symbols) == 0 {
// 		dto.InvalidParametersMessage(c, "symbols parameter is required")
// 		return
// 	}
// 	prices, err := h.priceService.GetPrices(c.Request.Context(), symbols)
// 	if err != nil {
// 		dto.InternallError(c)
// 		return
// 	}
// 	c.JSON(http.StatusOK, dto.ToPricesResponse(prices))
// }

// func (h *PriceGrpcHandler) GetPrice(c *gin.Context) {
// 	id := c.Param("id")
// 	if id == "" {
// 		dto.InvalidParametersMessage(c, "id parameter is required")
// 		return
// 	}
// 	price, err := h.priceService.GetPrice(c.Request.Context(), id)
// 	if err != nil {
// 		if errors.Is(err, core.ErrPriceNotFound) {
// 			dto.NotFoundErrorMessage(c, "requested price not found")
// 			return
// 		}
// 		dto.InternallError(c)
// 		return
// 	}
// 	resp := dto.ToPriceResponse(price)
// 	c.JSON(http.StatusOK, resp)
// }
