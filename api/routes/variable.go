package routes

import (
	"api/controllers"
	data "api/models"
	"encoding/json"

	"github.com/gin-gonic/gin"
)

func var_list(c *gin.Context) {
	name_project := c.Param("project")
	response := controllers.List_vars(name_project)
	var data map[string]interface{}
	_ = json.Unmarshal([]byte(response), &data)
	c.JSON(200, data)
}

func var_add(c *gin.Context) {
	var myVar data.Vardata
	c.ShouldBindJSON(&myVar)
	name_project := c.Param("project")
	name_var := c.Param("variable")
	response := controllers.Add_vars(name_project, name_var, myVar.Value)
	c.JSON(200, gin.H{
		"message": response,
	})
}

//func var_edit(c *gin.Context) {
//	name_project := c.Param("name")
//	response := controllers.Edit_vars(name_project)
//	c.JSON(200, gin.H{
//		"message": response,
//	})
//}
//
//func var_del(c *gin.Context) {
//	name_project := c.Param("name")
//	response := controllers.Del_vars(name_project)
//	c.JSON(200, gin.H{
//		"message": response,
//	})
//}
//
