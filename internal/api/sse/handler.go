package sse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/deleema/homelabwatch/internal/domain"
)

type EventSource interface {
	Subscribe(int) chan domain.EventEnvelope
	Unsubscribe(chan domain.EventEnvelope)
}

type Handler struct {
	source EventSource
}

func NewHandler(source EventSource) *Handler {
	return &Handler{source: source}
}

func (h *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	flusher, ok := writer.(http.Flusher)
	if !ok {
		http.Error(writer, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "text/event-stream")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")

	channel := h.source.Subscribe(32)
	defer h.source.Unsubscribe(channel)

	keepAlive := time.NewTicker(20 * time.Second)
	defer keepAlive.Stop()

	for {
		select {
		case <-request.Context().Done():
			return
		case <-keepAlive.C:
			fmt.Fprint(writer, ": keepalive\n\n")
			flusher.Flush()
		case event, ok := <-channel:
			if !ok {
				return
			}
			payload, err := json.Marshal(event)
			if err != nil {
				continue
			}
			fmt.Fprintf(writer, "event: %s\n", event.Type)
			fmt.Fprintf(writer, "data: %s\n\n", payload)
			flusher.Flush()
		}
	}
}
