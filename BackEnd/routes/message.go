package routes

import (
	"robix-backend/controllers"

	"github.com/gin-gonic/gin"
)

func message(c *gin.Context) {
	message1 := c.Param("message1")
	message2 := c.Param("message2")
	response := controllers.CallVault(message1, message2)
	c.JSON(200, gin.H{
		"message": response,
	})
}
