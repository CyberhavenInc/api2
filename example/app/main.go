package main

import (
	"log"
	"net/http"

	"github.com/CyberhavenInc/api2"
	"github.com/CyberhavenInc/api2/example"
)

func main() {
	service := example.NewEchoService(example.NewEchoRepository())
	routes := example.GetRoutes(service)
	api2.BindRoutes(http.DefaultServeMux, routes)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
