package main

import (
	"fmt"
	"generator/service"
	"github.com/gin-gonic/gin"
	"html/template"
	"net/http"
	"time"
)

func handler(s *service.Service) http.Handler {

	gin.SetMode(gin.DebugMode)
	handler := gin.Default()
	handler.SetHTMLTemplate(setupTemplates())

	handler.Static("/static", "./static/")
	handler.Static("/css", "./templates/css/")
	handler.Static("/js", "./templates/js/")
	handler.Static("/images", "./templates/images/")

	handler.GET("/ws", s.WsHandleConnections)

	front := handler.Group("/") //s.ResponseHtmlWriter

	front.GET("/", s.Index)
	front.GET("/organization/create", s.CreateOrganization)
	front.GET("/organization/:name", s.Organization)

	api := handler.Group("/api")
	// Создание организации и подтягивание проектов
	api.GET("/list-organization", s.ListOrganizationApi)

	// Создание структуры для организации и клонирование прото и спек
	api.POST("/generate-organization-struct", s.CreateOrganizationStructApi)

	// Создание структуры для проекта и клонирование прото и спек
	api.GET("/create-project", s.CreateProject)

	// Клонирование репозитория и переработка его в струкуру проекта для дальнейшей обработки
	api.GET("/clone-repository", s.CloneRepositoryApi)

	// Клонирование гетвея и переработка его в струкуру проекта для дальнейшей обработки
	api.GET("/clone-gateway", s.CloneGatewayApi)

	// создание структуры проекта папок для репозитория
	api.GET("/generate-path-project", s.GeneratePathProjectRepositoryApi)

	// Генератор файлов entity
	api.GET("/generate-entity", s.GenerateEntityApi)

	// Генератор файлов migration
	api.GET("/generate-migration", s.GenerateMigrationApi)
	// Генератор файлов migration
	api.GET("/generate-service", s.GenerateServiceFileApi)
	// Генератор файлов gateway
	api.GET("/generate-gateway", s.GenerateGatewayFileApi)
	// Генератор тестовых файлов
	api.GET("/generate-test", s.GenerateServiceTestFileApi)
	// Генератор  основных файлов
	api.GET("/generate-general", s.GenerateGeneralFileApi)

	//
	api.POST("/save-file", s.SaveFileApi)

	ssh := handler.Group("/ssh")

	ssh.GET("/update-and-tidy-modules", s.SshUpdateAndTidyModules)
	ssh.GET("/generate-swagger-server", s.GenerateSwaggerServer)
	ssh.GET("/generate-protobuf", s.GenerateProtobuf)

	return handler

}

func StartWebServer(s *service.Service) error {

	srv := http.Server{
		//Addr:              ":" + os.Getenv("HTTP_PORT"),
		Addr:              ":8090",
		IdleTimeout:       60 * time.Second,
		WriteTimeout:      60 * time.Second,
		ReadHeaderTimeout: 60 * time.Second,
		ReadTimeout:       60 * time.Second,
		MaxHeaderBytes:    8192,
		Handler:           handler(s),
	}

	return srv.ListenAndServe()
}

func setupTemplates() *template.Template {
	funcMap := template.FuncMap{
		"timeAgo": timeAgo,
	}

	tmpl := template.New("").Funcs(funcMap)
	tmpl, err := tmpl.ParseGlob("templates/*.html")
	if err != nil {
		panic(err)
	}

	return tmpl
}

func timeAgo(t time.Time) string {
	duration := time.Since(t)
	seconds := int64(duration.Seconds())

	switch {
	case seconds < 60:
		return fmt.Sprintf("%d seconds ago", seconds)
	case seconds < 3600:
		return fmt.Sprintf("%d minutes ago", seconds/60)
	case seconds < 86400:
		return fmt.Sprintf("%d hours ago", seconds/3600)
	case seconds > 100*86400:
		return ""
	default:
		return fmt.Sprintf("%d days ago", seconds/86400)
	}
}
