package routes

import (
	"github.com/gin-gonic/gin"
)

func ApplyRoutes(router *gin.Engine) {
	router.GET("/project/", project_list)
	router.GET("/project/:project/var/", var_list)
	router.POST("/project/:project/var/:variable", var_add)
	router.PATCH("/project/:project/var/:variable", var_edit)
	router.DELETE("/project/:project/var/:variable", var_del)
}
