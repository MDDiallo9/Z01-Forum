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
	Categories   *models.CategoriesModel
	Attachments  *models.AttachmentsModel
	Reports      *models.ReportsModel
	Likes        *models.LikesModel
	Comments     *models.CommentsModel
	Sessions     *services.SessionManager
	TemplateData TemplateData
}

type TemplateData struct {
	Form            any
	Reports         []*models.Report
	IsAuthenticated bool
	CurrentUserID   string
	IsAdmin         bool
}

func NewApplication(
	info *log.Logger,
	errLog *log.Logger,
	users *models.UsersModel,
	posts *models.PostsModel,
	categories *models.CategoriesModel,
	attachments *models.AttachmentsModel,
	reports *models.ReportsModel,
	likes *models.LikesModel,
	comments *models.CommentsModel,
	sessions *services.SessionManager,
) *Application {
	return &Application{
		InfoLog:     info,
		ErrorLog:    errLog,
		Users:       users,
		Posts:       posts,
		Categories:  categories,
		Attachments: attachments,
		Reports:     reports,
		Likes:       likes,
		Comments:    comments,
		Sessions:    sessions,
	}
}
