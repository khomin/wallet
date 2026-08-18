package core

import (
	"encoding/json"
	"sync"
	pricev1 "tracker/gen/price/v1"
	"tracker/internal/core/domain"
	"tracker/internal/messaging"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PriceHub struct {
	mu          sync.RWMutex
	subscribers map[string]chan *pricev1.PriceUpdate
	consumer    *messaging.Consumer
}

func NewPriceHub(consumer *messaging.Consumer) *PriceHub {
	return &PriceHub{
		subscribers: make(map[string]chan *pricev1.PriceUpdate),
		consumer:    consumer,
	}
}

func (h *PriceHub) Start() {
	go func() {
		for {
			deliveries, closeChan, err := h.consumer.Consume()
			if err != nil {
				logrus.Debugf("[PriceHub] RabbitMQ consume error: %v, retrying...", err)
				continue
			}
			logrus.Debugf("[PriceHub] Listening for price updates from RabbitMQ...")
		loop:
			for {
				select {
				case err := <-closeChan:
					logrus.Debugf("[PriceHub] Channel closed: %v. Reconnecting...", err)
					break loop
				case d, ok := <-deliveries:
					if !ok {
						break
					}
					var event []domain.TokenPrice
					if err := json.Unmarshal(d.Body, &event); err != nil {
						logrus.Debugf("[PriceHub] Failed to unmarshal price update: %v", err)
						_ = d.Nack(false, false)
						continue
					}
					_ = d.Ack(false)

					prices := []*pricev1.Price{}
					for _, i := range event {
						prices = append(prices, &pricev1.Price{
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
							UpdatedAt:                     timestamppb.New(i.UpdatedAt),
						})
					}
					h.broadcast(&pricev1.PriceUpdate{
						Price: prices,
					})
				}
			}
		}
	}()
}

func (h *PriceHub) Subscribe() (string, <-chan *pricev1.PriceUpdate) {
	h.mu.Lock()
	defer h.mu.Unlock()

	subID := uuid.NewString()
	ch := make(chan *pricev1.PriceUpdate, 20)
	h.subscribers[subID] = ch

	return subID, ch
}

func (h *PriceHub) Unsubscribe(subID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	ch, exists := h.subscribers[subID]
	if !exists {
		return
	}

	delete(h.subscribers, subID)
	close(ch)
}

func (h *PriceHub) broadcast(update *pricev1.PriceUpdate) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, ch := range h.subscribers {
		select {
		case ch <- update:
		default:
			select {
			case <-ch:
			default:
			}
			select {
			case ch <- update:
			default:
			}
		}
	}
}
