package handlers

import (
	"context"
	"tracker/internal/api/middleware"
	"tracker/internal/core"

	pricev1 "tracker/gen/price/v1"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PriceGrpcHandler struct {
	pricev1.UnimplementedPriceServiceServer
	priceService *core.PriceService
	priceHub     *core.PriceHub
	log          *logrus.Entry
}

func NewPriceGrpcHandler(service *core.PriceService, priceHub *core.PriceHub) pricev1.PriceServiceServer {
	return &PriceGrpcHandler{
		priceService: service,
		priceHub:     priceHub,
		log:          logrus.WithField("component", "PriceGrpcHandler"),
	}
}

func (s *PriceGrpcHandler) ListCoins(ctx context.Context, req *pricev1.ListCoinsRequest) (*pricev1.ListCoinsResponse, error) {
	coins, err := s.priceService.GetCoins(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*pricev1.Token, 0, len(coins))
	for _, i := range coins {
		chains := []*pricev1.Chain{}
		for _, i := range i.Chains {
			chains = append(chains, &pricev1.Chain{
				Symbol:   i.Symbol,
				Name:     i.Name,
				ImageUrl: i.ImageUrl,
			})
		}
		out = append(out, &pricev1.Token{
			Symbol:   i.Symbol,
			Name:     i.Name,
			Chains:   chains,
			Addrs:    i.Addrs,
			IsNative: i.IsNative,
			ImageUrl: i.ImageURL,
		})
	}
	return &pricev1.ListCoinsResponse{
		Total: int32(len(coins)),
		Token: out,
	}, nil
}

func (s *PriceGrpcHandler) GetCoin(ctx context.Context, req *pricev1.GetCoinRequest) (*pricev1.GetCoinResponse, error) {
	coin, err := s.priceService.GetCoin(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	chains := []*pricev1.Chain{}
	for _, i := range coin.Chains {
		chains = append(chains, &pricev1.Chain{
			Symbol:   i.Symbol,
			Name:     i.Name,
			ImageUrl: i.ImageUrl,
		})
	}
	return &pricev1.GetCoinResponse{
		Token: &pricev1.Token{
			Symbol:   coin.Symbol,
			Name:     coin.Name,
			Chains:   chains,
			Addrs:    coin.Addrs,
			IsNative: coin.IsNative,
			ImageUrl: coin.ImageURL,
		},
	}, nil
}

func (s *PriceGrpcHandler) GetPrices(ctx context.Context, req *pricev1.GetPricesRequest) (*pricev1.GetPricesResponse, error) {
	prices, err := s.priceService.GetPrices(ctx, req.Symbols)
	if err != nil {
		return nil, err
	}
	out := make([]*pricev1.Price, 0, len(prices))
	for _, i := range prices {
		out = append(out, i.ToGrpc())
	}
	return &pricev1.GetPricesResponse{
		Total: int32(len(prices)),
		Price: out,
	}, nil
}

func (s *PriceGrpcHandler) GetPrice(ctx context.Context, req *pricev1.GetPriceRequest) (*pricev1.GetPriceResponse, error) {
	price, err := s.priceService.GetPrice(ctx, req.Symbol)
	if err != nil {
		return nil, err
	}
	return &pricev1.GetPriceResponse{
		Price: price.ToGrpc(),
	}, nil
}

func (s *PriceGrpcHandler) SearchCoin(ctx context.Context, req *pricev1.SearchCoinsRequest) (*pricev1.SearchCoinsResponse, error) {
	tokens, err := s.priceService.SearchCoins(ctx, req.Text)
	if err != nil {
		return nil, err
	}
	out := make([]*pricev1.Token, 0, len(tokens))
	for _, i := range tokens {
		chains := []*pricev1.Chain{}
		for _, i := range i.Chains {
			chains = append(chains, &pricev1.Chain{
				Symbol:   i.Symbol,
				Name:     i.Name,
				ImageUrl: i.ImageUrl,
			})
		}
		out = append(out, &pricev1.Token{
			Symbol:   i.Symbol,
			Name:     i.Name,
			Chains:   chains,
			Addrs:    i.Addrs,
			IsNative: i.IsNative,
			ImageUrl: i.ImageURL,
		})
	}
	return &pricev1.SearchCoinsResponse{
		Total: int32(len(tokens)),
		Token: out,
	}, nil
}

func (s *PriceGrpcHandler) StreamPrices(
	req *pricev1.StreamPricesRequest,
	stream grpc.ServerStreamingServer[pricev1.PriceUpdate],
) error {
	ctx := stream.Context()
	_, ok := middleware.GetOAUTH(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "unauthorized")
	}
	subID, priceChan := s.priceHub.Subscribe()
	defer s.priceHub.Unsubscribe(subID)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case update, open := <-priceChan:
			if !open {
				return nil
			}
			if err := stream.Send(update); err != nil {
				return err
			}
		}
	}
}
