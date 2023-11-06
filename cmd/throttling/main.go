package main

import "net/http"

type LimitHandler struct {
	Next       http.Handler
	connection chan struct{}
}

func NewLimitHandler(next http.Handler, maxConnection int) *LimitHandler {
	cons := make(chan struct{}, maxConnection)
	for i := 0; i < maxConnection; i++ {
		cons <- struct{}{}
	}

	return &LimitHandler{
		Next:       next,
		connection: cons,
	}
}

func (h *LimitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	select {
	case <-h.connection:
		h.Next.ServeHTTP(w, r)
		h.connection <- struct{}{}
	default:
		http.Error(w, "busy", http.StatusTooManyRequests)
	}
}
