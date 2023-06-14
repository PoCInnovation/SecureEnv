package routes

import (
	"api/controllers"
	"encoding/json"

	"github.com/gin-gonic/gin"
)

func project_list(c *gin.Context) {
	response := controllers.List_projects()
	var data map[string]interface{}
	_ = json.Unmarshal([]byte(response), &data)
	c.JSON(200, data)
}
