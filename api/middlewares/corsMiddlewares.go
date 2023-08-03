package middlewares

import (
	"fmt"
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

		for name, values := range c.Request.Header {
			fmt.Printf("%s: %s\n", name, values)
		}

		client_token := ""
		auth_type := c.GetHeader("X-auth-type")

		if auth_type == "root-token" {
			client_token = c.GetHeader("X-auth-root-token")
		}

		config := vault.DefaultConfig()

		config.Address = "https://secure-env.poc-innovation.com:8200/"

		client, err := vault.NewClient(config)
		if err != nil {
			log.Fatalf("unable to initialize Vault client: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "client auth"})
			return
		}
		fmt.Printf(client_token)
		client.SetToken(client_token)

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
