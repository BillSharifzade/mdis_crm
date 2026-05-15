package middleware

import (
	"net/http"
	"os"
	"strings"
)

// CORSMiddleware — ограничивает источники списком ALLOWED_ORIGINS (через запятую).
// Пустой ALLOWED_ORIGINS → разрешаем * (dev-режим). В таком режиме credentials
// (cookie/Authorization) браузером всё равно отсекаются по спецификации.
func CORSMiddleware(next http.Handler) http.Handler {
	raw := os.Getenv("ALLOWED_ORIGINS")
	allowed := map[string]bool{}
	if raw != "" {
		for _, o := range strings.Split(raw, ",") {
			allowed[strings.TrimSpace(o)] = true
		}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if len(allowed) == 0 {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else if origin != "" && allowed[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Webhook-Secret")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
