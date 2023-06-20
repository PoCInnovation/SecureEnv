package routes

import (
	"api/controllers"
	"api/middlewares"
	data "api/models"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

func var_list(c *gin.Context) {
	name_project := c.Param("project")

	response, statusCode := controllers.List_vars(middlewares.GetClient(c), name_project)
	if statusCode != http.StatusOK {
		c.JSON(statusCode, gin.H{"error": response})
		return
	}

	var data map[string]interface{}
	_ = json.Unmarshal([]byte(response), &data)
	c.JSON(http.StatusOK, data)
}

func var_add(c *gin.Context) {
	var myVar data.Vardata
	c.ShouldBindJSON(&myVar)
	name_project := c.Param("project")
	name_var := c.Param("variable")

	response, statusCode := controllers.Add_vars(middlewares.GetClient(c), name_project, name_var, myVar.Value)

	c.JSON(statusCode, gin.H{
		"message": response,
	})
}

func var_edit(c *gin.Context) {
	var myVar data.Vardata
	c.ShouldBindJSON(&myVar)
	name_project := c.Param("project")
	name_var := c.Param("variable")

	response, statusCode := controllers.Edit_vars(middlewares.GetClient(c), name_project, name_var, myVar.Value)

	c.JSON(statusCode, gin.H{
		"message": response,
	})
}

func var_del(c *gin.Context) {
	name_project := c.Param("project")
	name_var := c.Param("variable")

	response, statusCode := controllers.Del_vars(middlewares.GetClient(c), name_project, name_var)

	c.JSON(statusCode, gin.H{
		"message": response,
	})
}
