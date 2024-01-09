package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"ueba-profiling/config"
	"ueba-profiling/controller"
	"ueba-profiling/manager"
)

func main() {
	log.Println("+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-")
	appConfig, err := config.LoadFile(config.DefaultConfigFilePath)
	if err != nil {
		log.Fatalf("failed to parse configuration file %s: %v", config.DefaultConfigFilePath, err)
	}
	config.AppConfig = appConfig

	jobManager := manager.NewJobManager()
	jobManager.Run()

	route := gin.Default()
	apiGroup := route.Group("/api/v1")
	jobHandler := controller.NewJobHandler(jobManager)
	jobHandler.MakeHandler(apiGroup)

	err = route.Run(fmt.Sprintf("%s:%d", config.AppConfig.Service.Host, config.AppConfig.Service.Port))
	if err != nil {
		log.Fatalf("failed to start the server: %v", err)
	}
}
