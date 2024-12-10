package example

import (
	"net/http"

	"github.com/CyberhavenInc/api2"
)

func GetRoutes(s IEchoService) []api2.Route {
	return []api2.Route{
		{Method: http.MethodPost, Path: "/hello", Handler: api2.Method(&s, "Hello")},
		{Method: http.MethodPost, Path: "/echo/:user", Handler: api2.Method(&s, "Echo")},
		{Method: http.MethodPost, Path: "/since", Handler: api2.Method(&s, "Since")},
		{Method: http.MethodPut, Path: "/stream", Handler: api2.Method(&s, "Stream")},
		{Method: http.MethodGet, Path: "/redirect", Handler: api2.Method(&s, "Redirect")},
		{Method: http.MethodPost, Path: "/raw", Handler: api2.Method(&s, "Raw")},
	}
}
