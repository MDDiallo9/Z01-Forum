package app

import (
	"forum/internal/models"
	"forum/internal/services"
	"log"
)

type Application struct {
	ErrorLog *log.Logger
	InfoLog *log.Logger
	Users *models.UsersModel
	Posts *models.PostsModel
	Attachments *models.AttachmentsModel
	TemplateData  TemplateData
	ErrorLog     *log.Logger
	InfoLog      *log.Logger
	Users        *models.UsersModel
	Posts        *models.PostsModel
	Sessions     *services.SessionManager
	TemplateData TemplateData
}

type TemplateData struct {
	Form any
}

func NewApplication(info, errLog *log.Logger, users *models.UsersModel,posts *models.PostsModel,attachments *models.AttachmentsModel) *Application {
    return &Application{InfoLog: info, ErrorLog: errLog, Users: users,Posts:posts,Attachments: attachments}
func NewApplication(info, errLog *log.Logger, users *models.UsersModel, posts *models.PostsModel, sessions *services.SessionManager) *Application {
	return &Application{InfoLog: info, ErrorLog: errLog, Users: users, Posts: posts, Sessions: sessions}
}
