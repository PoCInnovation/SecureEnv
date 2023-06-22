package routes

import (
	"api/controllers"
	"api/middlewares"
	data "api/models"
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

func project_create(c *gin.Context) {
	var myVar data.Vardata
	c.ShouldBindJSON(&myVar)

	response, statusCode := controllers.Create_project(middlewares.GetClient(c), myVar.Value)
	if statusCode >= http.StatusBadRequest {
		c.JSON(statusCode, gin.H{
			"error": response,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": response,
	})
}

func project_edit(c *gin.Context) {
	name_project := c.Param("project")
	var myVar data.Vardata
	c.ShouldBindJSON(&myVar)

	response, statusCode := controllers.Edit_project(middlewares.GetClient(c), name_project, myVar.Value)
	if statusCode >= http.StatusBadRequest {
		c.JSON(statusCode, gin.H{
			"error": response,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": response,
	})
}

func project_del(c *gin.Context) {
	name_project := c.Param("project")

	response, statusCode := controllers.Del_project(middlewares.GetClient(c), name_project)
	if statusCode >= http.StatusBadRequest {
		c.JSON(statusCode, gin.H{
			"error": response,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": response,
	})
}

func project_get(c *gin.Context) {
	name_project := c.Param("project")

	response, statusCode := controllers.Get_project(middlewares.GetClient(c), name_project)
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

func project_update(c *gin.Context) {
	name_project := c.Param("project")
	var newData map[string]interface{}
	c.ShouldBindJSON(&newData)

	response, statusCode := controllers.Update_project(middlewares.GetClient(c), name_project, newData)
	if statusCode >= http.StatusBadRequest {
		c.JSON(statusCode, gin.H{
			"error": response,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": response,
	})
}
