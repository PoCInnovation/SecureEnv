package routes

import (
	"robix-backend/controllers"

	"github.com/gin-gonic/gin"
)

func seemessage(c *gin.Context) {
	message1 := c.Param("message1")
	message2 := c.Param("message2")
	response := controllers.LookVault(message1, message2)
	c.JSON(200, gin.H{
		"message": response,
	})
}
