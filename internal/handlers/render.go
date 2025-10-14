package handlers

// Template loading, layout rendering, error/flash handling, status helpers (400/401/403/404/500)

import (
	"fmt"
	"forum/internal/app"
	"html/template"
	"net/http"
)

func render(w http.ResponseWriter, r *http.Request, f *app.Application, page string) {
	files := []string{
		fmt.Sprintf("./ui/templates/pages/%s", page),
		"./ui/templates/layouts/base.layout.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		f.ErrorLog.Println(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = ts.ExecuteTemplate(w, "base", nil)
	if err != nil {
		f.ErrorLog.Println(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
