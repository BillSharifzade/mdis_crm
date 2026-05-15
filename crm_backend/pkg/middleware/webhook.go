package middleware

import (
	"net/http"
	"os"
)

// WebhookSecretMiddleware validates the X-Webhook-Secret header
// against the WEBHOOK_SECRET environment variable.
// This prevents unauthorized parties from spoofing webhook calls.
func WebhookSecretMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secret := os.Getenv("WEBHOOK_SECRET")
		if secret == "" {
			// If no secret is configured, allow all requests (dev mode)
			next.ServeHTTP(w, r)
			return
		}

		headerSecret := r.Header.Get("X-Webhook-Secret")
		if headerSecret != secret {
			http.Error(w, `{"error":"Invalid webhook secret"}`, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
