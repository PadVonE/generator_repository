package main

import (
	"generator/service"
	"generator/usecase"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func handler(s *service.Service) http.Handler {

	gin.SetMode(gin.DebugMode)
	handler := gin.Default()


	s.Templates = usecase.GetTemplate()

	front := handler.Group("/",
		s.ResponseHtmlWriter,
	)

	handler.Static("/css", "./assets/css")
	handler.Static("/js", "./assets/js")
	handler.Static("/img", "./assets/img")
	handler.Static("/plugins", "./assets/plugins")
	handler.StaticFile("/favicon.ico", "./assets/favicon.ico")

	front.GET("/", s.Index)




	api := handler.Group("/api",
		)

	api.GET("/git-clone", s.GitClone)
	//handler.GET("/test",s.ForData)

	return handler

}

func StartWebServer(s *service.Service) error {

	srv := http.Server{
		Addr:              ":8080",
		IdleTimeout:       3 * time.Second,
		WriteTimeout:      5 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       3 * time.Second,
		MaxHeaderBytes:    8192,
		Handler:           handler(s),
	}

	return srv.ListenAndServe()
}
