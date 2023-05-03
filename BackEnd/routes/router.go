package routes

import (
	"github.com/gin-gonic/gin"
)

func ApplyRoutes(router *gin.Engine) {
	router.GET("/api/send/:message1/:message2", message)
	router.GET("/api/see/:message1/:message2", seemessage)
}
