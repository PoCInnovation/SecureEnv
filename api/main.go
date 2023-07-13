package main

import (
	"api/middlewares"
	"api/routes"
		"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.Use(middlewares.CorsMiddleware())
	router.Use(middlewares.AuthMiddleware())
	routes.ApplyRoutes(router)
	router.Run(":8080")
	return
}
