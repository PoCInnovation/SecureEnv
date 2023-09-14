package routes

import (
	"database/sql"
	"github.com/gin-gonic/gin"
)

func ApplyRoutes(router *gin.Engine, db *sql.DB) {
	router.GET("/project/", func(c *gin.Context) {
		project_list(c, db)
	})
	router.POST("/project/", func(c *gin.Context) {
		project_create(c, db)
	})
	router.PUT("/project/:project/", func(c *gin.Context) {
		project_edit(c, db)
	})
	router.DELETE("/project/:project/", func(c *gin.Context) {
		project_del(c, db)
	})
	router.GET("/project/:project/", project_get)
	router.PATCH("/project/:project/", func(c *gin.Context) {
		project_update(c, db)
	})
	router.GET("/project/:project/var/", var_list)
	router.POST("/project/:project/var/:variable", func(c *gin.Context) {
		var_add(c, db)
	})
	router.PATCH("/project/:project/var/:variable", func(c *gin.Context) {
		var_edit(c, db)
	})
	router.DELETE("/project/:project/var/:variable", func(c *gin.Context) {
		var_del(c, db)
	})
}
