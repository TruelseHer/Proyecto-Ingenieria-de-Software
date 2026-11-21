package handler

import (
	"context"
	"net"
	"net/http"
	"strings"
	"trotamundos/internal/platform"
	"trotamundos/internal/repository"
)

type ctxKey string

const (
	ctxKeyUserID   ctxKey = "user_id"
	ctxKeyClientIP ctxKey = "client_ip"

	SessionCookieName = "trotamundos_session"
)

func UserIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKeyUserID).(string); ok {
		return v
	}
	return ""
}

func ClientIPFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKeyClientIP).(string); ok && v != "" {
		return v
	}
	return "127.0.0.1"
}

func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

func RateLimitMiddleware(limiter *platform.RateLimiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/static/") || r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}
		ip := extractClientIP(r)
		if limiter != nil && !limiter.Allow(ip) {
			w.Header().Set("Retry-After", "60")
			http.Error(w, "Demasiadas solicitudes. Por favor aguarda unos segundos.", http.StatusTooManyRequests)
			return
		}
		ctx := context.WithValue(r.Context(), ctxKeyClientIP, ip)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func isPublicPath(path string) bool {
	if path == "/health" || path == "/login" || path == "/register" || path == "/favicon.ico" {
		return true
	}
	if strings.HasPrefix(path, "/static/") {
		return true
	}
	return false
}

func isAPIPath(path string) bool {
	return strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/mobile/")
}

// SessionMiddleware validates the authenticated user's session cookie or Bearer token.
// Unauthenticated requests to protected pages are redirected to /login.
func SessionMiddleware(tokens *platform.TokenService, store *repository.Store, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var userID string

		// 1. Check Authorization Bearer header
		if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
			rawToken := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
			claims, err := tokens.Validate(rawToken)
			if err != nil {
				writeJSONError(w, http.StatusUnauthorized, "Token de sesión inválido o revocado")
				return
			}
			if _, err := store.GetUser(claims.UserID); err != nil {
				writeJSONError(w, http.StatusUnauthorized, "Usuario no encontrado")
				return
			}
			userID = claims.UserID
		}

		// 2. Check signed HttpOnly session cookie
		if userID == "" {
			if c, err := r.Cookie(SessionCookieName); err == nil && c.Value != "" {
				if claims, err := tokens.Validate(c.Value); err == nil {
					if _, uErr := store.GetUser(claims.UserID); uErr == nil {
						userID = claims.UserID
					}
				}
			}
		}

		if userID == "" {
			if isPublicPath(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}
			if isAPIPath(r.URL.Path) {
				writeJSONError(w, http.StatusUnauthorized, "Debes iniciar sesión para acceder a este recurso")
				return
			}
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", "/login")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), ctxKeyUserID, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	if r.RemoteAddr != "" {
		return r.RemoteAddr
	}
	return "127.0.0.1"
}
