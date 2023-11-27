package main

import (
	"net/http"
	"time"

	"github.com/Kolo7/project-king/interval/handler"
	"github.com/Kolo7/project-king/pkg/middleware"
	"golang.org/x/time/rate"
)

func main() {
	normal := handler.NewHandler()
	limiter := rate.NewLimiter(rate.Every(time.Second), 1)
	normal = LimiterMiddleware(limiter)(normal)

	http.Handle("/hello", normal)
	http.ListenAndServe(":80", nil)
}

type Allow interface {
	Allow() bool
}

type Handler struct {
	limit Allow
	next  http.Handler
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !h.limit.Allow() {
		w.Write([]byte("limit"))
		return
	}
	h.next.ServeHTTP(w, r)
}

func LimiterMiddleware(limit Allow) middleware.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return &Handler{
			limit: limit,
			next:  next,
		}
	}
}
