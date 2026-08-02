package handlers

import (
	"context"
	"tracker/internal/core"

	pricev1 "tracker/gen/price/v1"

	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PriceGrpcHandler struct {
	pricev1.UnimplementedPriceServiceServer
	priceService *core.PriceService
	log          *logrus.Entry
}

func NewPriceGrpcHandler(service *core.PriceService) *PriceGrpcHandler {
	return &PriceGrpcHandler{
		priceService: service,
		log:          logrus.WithField("component", "PriceGrpcHandler"),
	}
}

func (s *PriceGrpcHandler) GetCoin(ctx context.Context, req *pricev1.GetCoinRequest) (*pricev1.GetCoinResponse, error) {
	coin, err := s.priceService.GetCoin(ctx, req.Coin)
	if err != nil {
		return nil, err
	}
	return &pricev1.GetCoinResponse{
		Symbol:   coin.Symbol,
		Name:     coin.Name,
		Chains:   coin.Chains,
		Addrs:    coin.Addrs,
		IsNative: coin.IsNative,
		ImageUrl: coin.ImageURL,
	}, nil
}

func (s *PriceGrpcHandler) GetCoins(ctx context.Context, req *pricev1.GetCoinsReq) (*pricev1.GetCoinsResp, error) {
	coins, err := s.priceService.GetCoins(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*pricev1.Token, 0, len(coins))
	for _, i := range coins {
		out = append(out, &pricev1.Token{
			Symbol:   i.Symbol,
			Name:     i.Name,
			Chains:   i.Chains,
			Addrs:    i.Addrs,
			IsNative: i.IsNative,
			ImageUrl: i.ImageURL,
		})
	}
	return &pricev1.GetCoinsResp{
		Total: int32(len(coins)),
		Token: out,
	}, nil
}

func (s *PriceGrpcHandler) GetPrices(ctx context.Context, req *pricev1.GetPricesReq) (*pricev1.GetPricesResp, error) {
	prices, err := s.priceService.GetPrices(ctx, req.Symbols)
	if err != nil {
		return nil, err
	}
	out := make([]*pricev1.Price, 0, len(prices))
	for _, i := range prices {
		out = append(out, &pricev1.Price{
			Symbol:                        i.Symbol,
			Name:                          i.Name,
			PriceUsd:                      float32(i.CurrentPrice),
			MarketCap:                     float32(i.MarketCap),
			TotalVolume:                   float32(i.TotalVolume),
			High_24H:                      float32(i.High_24h),
			Low_24H:                       float32(i.Low_24h),
			PriceChange_24H:               float32(i.PriceChange_24h),
			PriceChangePercentage_24H:     float32(i.PriceChangePercentage_24h),
			MarketCapChange_24H:           float32(i.MarketCapChange_24h),
			MarketCapChangePercentage_24H: float32(i.MarketCapChange_percentage_24h),
			LastUpdated:                   timestamppb.New(i.LastUpdated),
		})
	}
	return &pricev1.GetPricesResp{
		Total: int32(len(prices)),
		Price: out,
	}, nil
}

func (s *PriceGrpcHandler) GetPrice(ctx context.Context, req *pricev1.GetPriceReq) (*pricev1.GetPriceResp, error) {
	price, err := s.priceService.GetPrice(ctx, req.Symbol)
	if err != nil {
		return nil, err
	}
	return &pricev1.GetPriceResp{
		Price: &pricev1.Price{
			Symbol:                        price.Symbol,
			Name:                          price.Name,
			PriceUsd:                      float32(price.CurrentPrice),
			MarketCap:                     float32(price.MarketCap),
			TotalVolume:                   float32(price.TotalVolume),
			High_24H:                      float32(price.High_24h),
			Low_24H:                       float32(price.Low_24h),
			PriceChange_24H:               float32(price.PriceChange_24h),
			PriceChangePercentage_24H:     float32(price.PriceChangePercentage_24h),
			MarketCapChange_24H:           float32(price.MarketCapChange_24h),
			MarketCapChangePercentage_24H: float32(price.MarketCapChange_percentage_24h),
			LastUpdated:                   timestamppb.New(price.LastUpdated),
		},
	}, nil
}

func (s *PriceGrpcHandler) SearchCoin(ctx context.Context, req *pricev1.SearchCoinsReq) (*pricev1.SearchCoinsResp, error) {
	tokens, err := s.priceService.SearchCoins(ctx, req.Text)
	if err != nil {
		return nil, err
	}
	out := make([]*pricev1.Token, 0, len(tokens))
	for _, i := range tokens {
		out = append(out, &pricev1.Token{
			Symbol:   i.Symbol,
			Name:     i.Name,
			Chains:   i.Chains,
			Addrs:    i.Addrs,
			IsNative: i.IsNative,
			ImageUrl: i.ImageURL,
		})
	}
	return &pricev1.SearchCoinsResp{
		Total: int32(len(tokens)),
		Token: out,
	}, nil
}
