package handlers

import (
	"forum/internal/app"
	"net/http"
)

func SessionTest(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := f.Sessions.GetUserFromRequest(r)
		if err != nil || userID == "" {
			w.Write([]byte("No active sessions or invalid session"))
			return
		}

		w.Write([]byte("Active session for userID : " + userID + "/n"))
	}
}
