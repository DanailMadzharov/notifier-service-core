package http

import (
	"github.com/gorilla/mux"
	"net/http"
)

// RouteInfo represents information about an HTTP route
type RouteInfo struct {
	Name    string
	Pattern string
	Method  string
	Handler http.HandlerFunc
}

// SetupRoutes sets up the HTTP routes and returns information about each route
func SetupRoutes(handler *Handler) []RouteInfo {

	routes := []RouteInfo{
		{
			Name:    "SendNotification",
			Pattern: "/v1/notification/{channel}",
			Method:  http.MethodPost,
			Handler: handler.handleNotificationCreation,
		},
	}

	r := mux.NewRouter()

	for _, route := range routes {
		r.HandleFunc(route.Pattern, route.Handler).Methods(route.Method)
	}

	http.Handle("/", r)

	return routes
}
