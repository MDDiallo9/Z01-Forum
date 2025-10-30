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
	Reports      *models.ReportsModel
	Sessions     *services.SessionManager
	TemplateData TemplateData
}

type TemplateData struct {
	Form    any
	Reports []*models.Report
}

func NewApplication(
	info *log.Logger,
	errLog *log.Logger,
	users *models.UsersModel,
	posts *models.PostsModel,
	attachments *models.AttachmentsModel,
	reports *models.ReportsModel,
	sessions *services.SessionManager,
) *Application {
	return &Application{
		InfoLog:     info,
		ErrorLog:    errLog,
		Users:       users,
		Posts:       posts,
		Attachments: attachments,
		Reports:     reports,
		Sessions:    sessions,
	}
}
