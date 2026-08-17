package core

import (
	"encoding/json"
	"log"
	"sync"
	"time"
	pricev1 "tracker/gen/price/v1"
	"tracker/internal/core/domain"
	"tracker/internal/messaging"

	"github.com/google/uuid"
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
				log.Printf("[PriceHub] RabbitMQ consume error: %v, retrying...", err)
				continue
			}

			log.Println("[PriceHub] Listening for price updates from RabbitMQ...")

			for {
				select {
				case err := <-closeChan:
					log.Printf("[PriceHub] Channel closed: %v. Reconnecting...", err)
					// breaks inner loop to trigger re-consume
					break
				case d, ok := <-deliveries:
					if !ok {
						break
					}
					var prices []domain.TokenPrice
					if err := json.Unmarshal(d.Body, &prices); err != nil {
						log.Printf("[PriceHub] Failed to unmarshal price update: %v", err)
						_ = d.Nack(false, false)
						continue
					}
					_ = d.Ack(false)
					h.broadcast(&pricev1.PriceUpdate{
						Symbol:    "TEST",
						PriceUsd:  "1234",
						Timestamp: time.Now().UnixMilli(),
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
			// buffer full - drop update for slow readers so we don't stall the hub
		}
	}
}
