package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/schachte/websockets/internal/handlers"
)

func main() {
	routeList := GetRoutes()

	log.Println("Starting channel listener")
	go handlers.ListenToWsChannel()

	log.Println("Starting server on 8081")
	err := http.ListenAndServe(":8081", routeList)

	if err != nil {
		fmt.Println(err)
	}
}
