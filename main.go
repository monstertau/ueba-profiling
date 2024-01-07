package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"ueba-profiling/config"
	"ueba-profiling/controller"
	"ueba-profiling/manager"
)

func main() {
	log.Println("+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-")
	appConfig, err := config.NewAppConfigFrom(config.DefConfigFilePath)
	if err != nil {
		log.Fatalf("failed to parse configuration file %s: %v", config.DefConfigFilePath, err)
	}
	config.GlobalConfig = appConfig

	jobManager := manager.NewJobManager()
	jobManager.Run()

	route := gin.Default()
	apiGroup := route.Group("/api/v1")
	jobHandler := controller.NewJobHandler(jobManager)
	jobHandler.MakeHandler(apiGroup)

	err = route.Run(appConfig.GetServerAddr())
	if err != nil {
		log.Fatalf("failed to start the server: %v", err)
	}
}
