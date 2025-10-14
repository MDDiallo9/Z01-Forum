package app

import (
	"net/http"
)

// Create http.Server, timeouts, graceful shutdown, health check

func Server(f *Application,handler http.Handler) *http.Server{
	return &http.Server{
		Addr: ":8000",
		ErrorLog: ErrorLog,
		Handler: handler,
	}
}