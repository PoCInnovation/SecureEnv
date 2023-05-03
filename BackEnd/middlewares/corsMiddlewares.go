package middlewares

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CorsMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowMethods:           []string{"*"},
		AllowHeaders:           []string{"*"},
		AllowAllOrigins:        true,
		AllowCredentials:       true,
		AllowWildcard:          true,
		AllowBrowserExtensions: true,
	})
}
