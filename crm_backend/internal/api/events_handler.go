package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"crm_backend/pkg/events"
	"github.com/golang-jwt/jwt/v5"
)

type EventsHandler struct {
	bus *events.Bus
}

func NewEventsHandler(bus *events.Bus) *EventsHandler {
	return &EventsHandler{bus: bus}
}

// stream upgrades the connection to Server-Sent Events.
// EventSource в браузере не умеет ставить кастомные заголовки, поэтому
// JWT принимается либо через Authorization: Bearer, либо через ?token=...
// в query-string.
//
// @Summary Live events stream (SSE)
// @Description Holds an open connection and pushes JSON events on every CRM
// @Description mutation (lead created/updated/deleted, status changed,
// @Description interaction added, telegram message). Auth: Bearer in header OR
// @Description ?token=<jwt> query param.
// @Tags events
// @Produce text/event-stream
// @Success 200 {string} string "event stream"
// @Router /events [get]
func (h *EventsHandler) stream(w http.ResponseWriter, r *http.Request) {
	if !validateEventsAuth(r) {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Nginx: do not buffer
	w.WriteHeader(http.StatusOK)

	// Initial hello so the client can confirm the connection is live.
	fmt.Fprintf(w, "event: hello\ndata: {\"ok\":true}\n\n")
	flusher.Flush()

	ch, cancel := h.bus.Subscribe()
	defer cancel()

	keepalive := time.NewTicker(25 * time.Second)
	defer keepalive.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-keepalive.C:
			// SSE-комментарий — держит соединение через прокси.
			fmt.Fprintf(w, ": ping %d\n\n", time.Now().Unix())
			flusher.Flush()
		case evt, ok := <-ch:
			if !ok {
				return
			}
			body, err := json.Marshal(evt)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", evt.Type, body)
			flusher.Flush()
		}
	}
}

// Routes exposes the SSE endpoint. Auth is handled inline (header OR query).
func (h *EventsHandler) Route(w http.ResponseWriter, r *http.Request) {
	h.stream(w, r)
}

func validateEventsAuth(r *http.Request) bool {
	var tokenStr string
	if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		tokenStr = strings.TrimPrefix(auth, "Bearer ")
	} else {
		tokenStr = r.URL.Query().Get("token")
	}
	if tokenStr == "" {
		return false
	}
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "fallback_secret_for_dev_only"
	}
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, http.ErrAbortHandler
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return false
	}
	return true
}
