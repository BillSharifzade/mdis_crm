package middleware

import (
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

// SecurityHeaders проставляет базовые защитные заголовки.
// HSTS включается только если ENV != "dev".
func SecurityHeaders(next http.Handler) http.Handler {
	prod := os.Getenv("APP_ENV") == "production"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		// CSP без unsafe-inline для скриптов; стили — да, т.к. лиц-в-инлайн style attributes в UI.
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; img-src 'self' data: https:; style-src 'self' 'unsafe-inline'; "+
				"script-src 'self'; connect-src 'self' https: wss: http://localhost:* http://127.0.0.1:*; "+
				"font-src 'self' data:; frame-ancestors 'none'")
		if prod {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		next.ServeHTTP(w, r)
	})
}

// BodySizeLimit ограничивает размер тела запроса, чтобы предотвратить DoS большими payload.
// Multipart (импорт) разрешает до 30MB через свой ParseMultipartForm; здесь общий лимит 5MB.
func BodySizeLimit(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Не вмешиваемся в multipart upload (только под /leads/import & /leads-import).
			if ct := r.Header.Get("Content-Type"); len(ct) >= 19 && ct[:19] == "multipart/form-data" {
				next.ServeHTTP(w, r)
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

// LoginRateLimiter — простой in-memory rate-limit для /auth/login.
// Не подходит для multi-instance, для одного pod достаточно.
type LoginRateLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
	max      int
	window   time.Duration
}

func NewLoginRateLimiter(max int, window time.Duration) *LoginRateLimiter {
	return &LoginRateLimiter{attempts: make(map[string][]time.Time), max: max, window: window}
}

func (l *LoginRateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}
		ip := clientIP(r)
		now := time.Now()
		l.mu.Lock()
		// очищаем устаревшие
		cutoff := now.Add(-l.window)
		filtered := l.attempts[ip][:0]
		for _, t := range l.attempts[ip] {
			if t.After(cutoff) {
				filtered = append(filtered, t)
			}
		}
		if len(filtered) >= l.max {
			l.attempts[ip] = filtered
			l.mu.Unlock()
			w.Header().Set("Retry-After", "60")
			http.Error(w, `{"error":"Too many login attempts. Try again later."}`, http.StatusTooManyRequests)
			return
		}
		filtered = append(filtered, now)
		l.attempts[ip] = filtered
		l.mu.Unlock()
		next.ServeHTTP(w, r)
	})
}

func clientIP(r *http.Request) string {
	if v := r.Header.Get("X-Real-IP"); v != "" {
		return v
	}
	if v := r.Header.Get("X-Forwarded-For"); v != "" {
		// Берём первый IP из цепочки X-Forwarded-For.
		for i := 0; i < len(v); i++ {
			if v[i] == ',' {
				return v[:i]
			}
		}
		return v
	}
	// RemoteAddr приходит как "ip:port" — отрезаем порт.
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
