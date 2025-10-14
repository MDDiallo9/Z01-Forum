package app

import (
	"forum/internal/models"
	"log"
)

type Application struct {
	ErrorLog *log.Logger
	InfoLog *log.Logger
	Users *models.UsersModel
	
}

func NewApplication(info, errLog *log.Logger, users *models.UsersModel) *Application {
    return &Application{InfoLog: info, ErrorLog: errLog, Users: users}
}

