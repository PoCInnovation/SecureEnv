package routes

import (
	"api/controllers"
	"api/middlewares"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

func project_list(c *gin.Context) {
	response, statusCode := controllers.List_projects(middlewares.GetClient(c))
	if statusCode >= http.StatusBadRequest {
		c.JSON(statusCode, gin.H{
			"error": response,
		})
		return
	}
	var data map[string]interface{}
	_ = json.Unmarshal([]byte(response), &data)
	c.JSON(http.StatusOK, data)
}
