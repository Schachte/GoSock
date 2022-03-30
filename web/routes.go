package main

import (
	"net/http"

	"github.com/bmizerany/pat"
	"github.com/schachte/websockets/internal/handlers"
)

func GetRoutes() http.Handler {
	mux := pat.New()

	mux.Get("/", http.HandlerFunc(handlers.Home))
	mux.Get("/ws", http.HandlerFunc(handlers.WsEndpoint))

	return mux
}
