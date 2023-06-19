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
		// Auth -> Vault

		config := vault.DefaultConfig()

		config.Address = adr_vault

		client, err := vault.NewClient(config)
		if err != nil {
			log.Fatalf("unable to initialize Vault client: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "client auth"})
			return
		}

		client.SetToken(token)

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
