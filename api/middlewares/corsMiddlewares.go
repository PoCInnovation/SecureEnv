package middlewares

import (
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	vault "github.com/hashicorp/vault/api"
)

func GetClient(c *gin.Context) *vault.Client {
	client, _ := c.Get("vaultClient")
	vaultClient := client.(*vault.Client)
	return vaultClient
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		getEnvFile()
		// Auth -> Vault

		config := vault.DefaultConfig()

		config.Address = adr_vault

		client, err := vault.NewClient(config)
		if err != nil {
			log.Fatalf("unable to initialize Vault client: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "client auth"})
			return
		}

		// authHeader := c.GetHeader("Authorization")
		// if authHeader == "" {
		// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
		// 	c.Abort()
		// 	return
		// }

		// var token string
		// _, err = fmt.Sscanf(authHeader, "Bearer %s", &token)
		// if err != nil {
		// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
		// 	c.Abort()
		// 	return
		// }

		client.SetToken("hvs.NWAqB95aaCepzz7GXVrO43KW")

		c.Set("vaultClient", client)

		c.Next()
	}
}

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
