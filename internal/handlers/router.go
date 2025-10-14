package handlers

import (
	"forum/internal/app"
	"net/http"
)

// Register routes and all handlerson ServeMux; mount static, compose middleware

func Routes(f *app.Application) *http.ServeMux {
	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/",http.StripPrefix("/static",fileServer))

	mux.HandleFunc("/",Home(f))
	return mux
}
