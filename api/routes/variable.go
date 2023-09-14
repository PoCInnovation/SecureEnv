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

func var_list(c *gin.Context) {
	name_project := c.Param("project")

	response, statusCode := controllers.List_vars(middlewares.GetClient(c), name_project)
	if statusCode >= http.StatusBadRequest {
		c.JSON(statusCode, gin.H{"error": response})
		return
	}

	var data map[string]interface{}
	_ = json.Unmarshal([]byte(response), &data)
	c.JSON(http.StatusOK, data)
}

func var_add(c *gin.Context, db *sql.DB) {
	var myVar data.Vardata
	c.ShouldBindJSON(&myVar)

	name_project := c.Param("project")
	name_var := c.Param("variable")
	ip_adress := c.ClientIP()
	action_description := ip_adress + ": added variable" + name_var + " in project " + name_project

	response, statusCode := controllers.Add_vars(middlewares.GetClient(c), name_project, name_var, myVar.Value)
	if statusCode >= http.StatusBadRequest {
		c.JSON(statusCode, gin.H{
			"error": response,
		})
		return
	}
	var project_id int
	db.QueryRow("SELECT id FROM projects WHERE name = ?", name_project).Scan(&project_id)

	_, err := db.Exec("INSERT INTO logs (ip_adress, action, project_id) VALUES (?, ?, ?)", ip_adress, action_description, project_id)
	if err != nil {
		println(err.Error())
		return
	}

	c.JSON(statusCode, gin.H{
		"message": response,
	})
}

func var_edit(c *gin.Context, db *sql.DB) {
	var myVar data.Vardata
	c.ShouldBindJSON(&myVar)
	name_project := c.Param("project")
	name_var := c.Param("variable")
	ip_adress := c.ClientIP()
	action_description := ip_adress + ": edited variable " + name_var + " in project " + name_project

	response, statusCode := controllers.Edit_vars(middlewares.GetClient(c), name_project, name_var, myVar.Value)
	if statusCode >= http.StatusBadRequest {
		c.JSON(statusCode, gin.H{
			"error": response,
		})
		return
	}
	var project_id int
	db.QueryRow("SELECT id FROM projects WHERE name = ?", name_project).Scan(&project_id)

	_, err := db.Exec("INSERT INTO logs (ip_adress, action, project_id) VALUES (?, ?, ?)", ip_adress, action_description, project_id)
	if err != nil {
		println(err.Error())
		return
	}

	c.JSON(statusCode, gin.H{
		"message": response,
	})
}

func var_del(c *gin.Context, db *sql.DB) {
	name_project := c.Param("project")
	name_var := c.Param("variable")
	ip_adress := c.ClientIP()
	action_description := ip_adress + ": Deleted variable " + name_var + " in project " + name_project

	response, statusCode := controllers.Del_vars(middlewares.GetClient(c), name_project, name_var)
	if statusCode >= http.StatusBadRequest {
		c.JSON(statusCode, gin.H{
			"error": response,
		})
		return
	}
	var project_id int
	db.QueryRow("SELECT id FROM projects WHERE name = ?", name_project).Scan(&project_id)

	_, err := db.Exec("INSERT INTO logs (ip_adress, action, project_id) VALUES (?, ?, ?)", ip_adress, action_description, project_id)
	if err != nil {
		println(err.Error())
		return
	}

	c.JSON(statusCode, gin.H{
		"message": response,
	})
}
