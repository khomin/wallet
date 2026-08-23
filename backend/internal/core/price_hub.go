package core

import (
	"encoding/json"
	"sync"
	pricev1 "tracker/gen/price/v1"
	"tracker/internal/core/domain"
	"tracker/internal/messaging"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
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
	log := logrus.WithField("EmailWorker", "Start")
	go func() {
		for {
			deliveries, closeChan, err := h.consumer.Consume()
			if err != nil {
				log.Debugf("consume error: %v, retrying...", err)
				continue
			}
		loop:
			for {
				select {
				case err := <-closeChan:
					log.Debugf("Channel closed: %v. Reconnecting...", err)
					break loop
				case d, ok := <-deliveries:
					if !ok {
						break
					}
					var event []domain.TokenPrice
					if err := json.Unmarshal(d.Body, &event); err != nil {
						log.Debugf("Failed to unmarshal event: %v", err)
						_ = d.Nack(false, false)
						continue
					}
					_ = d.Ack(false)

					prices := []*pricev1.Price{}
					for _, i := range event {
						prices = append(prices, i.ToGrpc())
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
