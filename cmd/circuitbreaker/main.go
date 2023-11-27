package main

import (
	"net/http"

	"github.com/Kolo7/project-king/interval/handler"
)

func main() {
	normal := handler.NewHandler()

	http.Handle("/hello", normal)
	http.ListenAndServe(":80", nil)
}

type Middler struct {
	next http.Handler
}

func (m *Middler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	m.next.ServeHTTP(w, r)
}
