package app

import (
	"forum/internal/models"
	"forum/internal/services"
	"log"
)

type Application struct {
	ErrorLog     *log.Logger
	InfoLog      *log.Logger
	Users        *models.UsersModel
	Sessions     *services.SessionManager // Session Manager for the session
	TemplateData TemplateData
}

type TemplateData struct {
	Form any
}

func NewApplication(info, errLog *log.Logger, users *models.UsersModel, sessions *services.SessionManager) *Application {
	return &Application{InfoLog: info, ErrorLog: errLog, Users: users, Sessions: sessions}
}
