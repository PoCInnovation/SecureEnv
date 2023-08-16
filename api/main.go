package main

import (
	"api/config"
	"api/middlewares"
	"api/routes"
	"fmt"
	"github.com/gin-gonic/gin"
)

func main() {
	db, err := config.ConnectDB()
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	fmt.Println("connected to database")
	router := gin.Default()
	router.Use(middlewares.CorsMiddleware())
	router.Use(middlewares.AuthMiddleware())
	routes.ApplyRoutes(router)
	router.Run(":8080")
	return
}
