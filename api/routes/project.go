package routes

import (
	"api/controllers"
	"api/middlewares"
	data "api/models"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

func project_list(c *gin.Context, db *sql.DB) {
	response, statusCode := controllers.List_projects(middlewares.GetClient(c))
	if statusCode >= http.StatusBadRequest {
		c.JSON(statusCode, gin.H{
			"error": response,
		})
		return
	}

	//db.Exec("INSERT INTO projects (name) VALUES (?)", response)

	var data map[string]interface{}
	_ = json.Unmarshal([]byte(response), &data)
	c.JSON(http.StatusOK, data)
}

func project_create(c *gin.Context, db *sql.DB) {
	var myVar data.Vardata
	c.ShouldBindJSON(&myVar)
	ip_adress := c.Request.Header.Get("X-Forwarded-For")
	action_description := ip_adress + ": Created project " + myVar.Value

	_, err := db.Exec("INSERT INTO projects (name) VALUES (?)", myVar.Value)
	if err != nil {
		println(err.Error())
		println("Error in INSERT INTO projects (name) VALUES (?)" + myVar.Value)
		return
	}

	var project_id int
	err = db.QueryRow("SELECT id FROM projects WHERE name = ?", myVar.Value).Scan(&project_id)
	if err != nil {
		println(err.Error())
		println("Error in SELECT id FROM projects WHERE name = ?" + myVar.Value)
		return
	}
	_, err = db.Exec("INSERT INTO logs (ip_adress, action, project_id) VALUES (?, ?, ?)", ip_adress, action_description, project_id)
	if err != nil {
		println(err.Error())
		println("ip:" + ip_adress)
		println("Error in INSERT INTO logs (ip_adress, description, project_id) VALUES (?, ?)")
		return
	}

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

func project_edit(c *gin.Context, db *sql.DB) {
	name_project := c.Param("project")
	var myVar data.Vardata
	c.ShouldBindJSON(&myVar)
	ip_adress := c.ClientIP()
	action_description := ip_adress + ": Edited project " + name_project

	var project_id int
	err := db.QueryRow("SELECT id FROM projects WHERE name = ?", name_project).Scan(&project_id)
	if err != nil {
		println(err.Error())
		return
	}
	_, err = db.Exec("INSERT INTO logs (ip_adress, action, project_id) VALUES (?, ?, ?)", ip_adress, action_description, project_id)
	if err != nil {
		println(err.Error())
		return
	}

	_, err = db.Exec("UPDATE projects SET name = ? WHERE name = ?", myVar.Value, name_project)
	if err != nil {
		println(err.Error())
		return
	}

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

func project_del(c *gin.Context, db *sql.DB) {
	name_project := c.Param("project")
	ip_adress := c.ClientIP()
	action_description := ip_adress + ": Deleted project " + name_project

	var project_id int
	err := db.QueryRow("SELECT id FROM projects WHERE name = ?", name_project).Scan(&project_id)
	if err != nil {
		println(err.Error())
		return
	}

	_, err = db.Exec("INSERT INTO logs (ip_adress, action, project_id) VALUES (?, ?, ?)", ip_adress, action_description, project_id)
	if err != nil {
		println(err.Error())
		return
	}
	_, err = db.Exec("DELETE FROM projects WHERE name = ?", name_project)
	if err != nil {
		println(err.Error())
		return
	}

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

	forcePush := false
	pushHeader := c.GetHeader("push-force")
	if pushHeader == "true" {
		forcePush = true
	}
	response, statusCode := controllers.Update_project(middlewares.GetClient(c), name_project, newData, forcePush)
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
