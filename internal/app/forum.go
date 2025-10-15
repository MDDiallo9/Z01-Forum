package app

import (
	"forum/internal/models"
	"log"
)

type Application struct {
	ErrorLog *log.Logger
	InfoLog *log.Logger
	Users *models.UsersModel
	Posts *models.PostsModel
	TemplateData  TemplateData
}

type TemplateData struct {
	Form any
}

func NewApplication(info, errLog *log.Logger, users *models.UsersModel,posts *models.PostsModel) *Application {
    return &Application{InfoLog: info, ErrorLog: errLog, Users: users,Posts:posts}
}

