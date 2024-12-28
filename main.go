package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/imafaz/B9CA/api"
	"github.com/imafaz/B9CA/controller"
)

func main() {

	bindManager := controller.NewBindManager(`/etc/bind/zones`, `/etc/bind/named.conf.local`)
	router := gin.Default()

	api := api.NewAPI(bindManager)
	api.SetupRoutes(router)

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
