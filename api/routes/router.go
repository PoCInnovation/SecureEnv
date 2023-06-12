package routes

import (
	"github.com/gin-gonic/gin"
)

func ApplyRoutes(router *gin.Engine) {
	router.GET("/project/:project/var/", var_list)
	router.POST("/project/:project/var/:variable/", var_add)
	//router.PATCH("/project/:name/var/", var_edit)
	//router.DELETE("/project/:name/var/", var_del)
}
