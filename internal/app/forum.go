package app

import (
	"forum/internal/models"
	"forum/internal/services"
	"log"
)

type Application struct {
	InfoLog      *log.Logger
	ErrorLog     *log.Logger
	Users        *models.UsersModel
	Posts        *models.PostsModel
	Attachments  *models.AttachmentsModel
	Sessions     *services.SessionManager
	TemplateData TemplateData
}

type TemplateData struct {
	Form any
}

func NewApplication(
	info *log.Logger,
	errLog *log.Logger,
	users *models.UsersModel,
	posts *models.PostsModel,
	attachments *models.AttachmentsModel,
	sessions *services.SessionManager,
) *Application {
	return &Application{
		InfoLog:     info,
		ErrorLog:    errLog,
		Users:       users,
		Posts:       posts,
		Attachments: attachments,
		Sessions:    sessions,
	}
}
