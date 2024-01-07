package middleware

import "net/http"

type MiddlewareFunc func(http.Handler) http.Handler

func xMiddware() MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}
}
