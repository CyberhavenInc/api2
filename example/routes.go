package example

import (
	"net/http"

	"github.com/starius/api2"
)

func GetRoutes(s IEchoService) []api2.Route {
	return []api2.Route{
		{Method: http.MethodPost, Path: "/hello", Handler: api2.Method(&s, "Hello")},
		{Method: http.MethodPost, Path: "/echo/:user", Handler: api2.Method(&s, "Echo")},
		{Method: http.MethodPost, Path: "/since", Handler: api2.Method(&s, "Since")},
		{Method: http.MethodPut, Path: "/stream", Handler: api2.Method(&s, "Stream")},
		{Method: http.MethodGet, Path: "/redirect", Handler: api2.Method(&s, "Redirect")},
		{Method: http.MethodPost, Path: "/raw", Handler: api2.Method(&s, "Raw")},
		{Method: http.MethodPost, Path: "/wildcard/:param_a/static-part/:param_b*/static-part/:param_c/static_part/:param_d*/static_part/:param_e/static_part/:param_f*/static_part", Handler: api2.Method(&s, "AdvancedWildcard")},
		{Method: http.MethodPost, Path: "/wildcard/:param_a/static-part/:param_b*", Handler: api2.Method(&s, "BasicWildcard")},
	}
}
