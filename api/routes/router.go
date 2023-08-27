package routes

import (
	"database/sql"
	"github.com/gin-gonic/gin"
)

func ApplyRoutes(router *gin.Engine, db *sql.DB) {
	router.GET("/project/", func(c *gin.Context) {
		project_list(c, db)
	})
	router.POST("/project/", project_create)
	router.PUT("/project/:project/", project_edit)
	router.DELETE("/project/:project/", project_del)
	router.GET("/project/:project/", project_get)
	router.PATCH("/project/:project/", project_update)
	router.GET("/project/:project/var/", var_list)
	router.POST("/project/:project/var/:variable", var_add)
	router.PATCH("/project/:project/var/:variable", var_edit)
	router.DELETE("/project/:project/var/:variable", var_del)
}
