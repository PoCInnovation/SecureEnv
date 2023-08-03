package middlewares

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	vault "github.com/hashicorp/vault/api"
)

func getVaultTokenWithGitHubAuth(vaultAddress string, githubToken string) (string, error) {
	config := &vault.Config{
		Address: vaultAddress,
	}

	client, err := vault.NewClient(config)
	if err != nil {
		return "", fmt.Errorf("failed to create vault client: %v", err)
	}

	data := map[string]interface{}{
		"token": githubToken,
	}

	secret, err := client.Logical().Write("auth/github/login", data)
	if err != nil {
		return "", fmt.Errorf("failed to login with GitHub token: %v", err)
	}

	if secret == nil || secret.Auth == nil || secret.Auth.ClientToken == "" {
		return "", fmt.Errorf("failed to retrieve client token from Vault")
	}

	return secret.Auth.ClientToken, nil
}

func getVaultTokenWithJWT(vaultAddress string, role, jwtToken string) (string, error) {
	config := &vault.Config{
		Address: vaultAddress,
	}

	client, err := vault.NewClient(config)
	if err != nil {
		return "", fmt.Errorf("failed to create vault client: %v", err)
	}

	data := map[string]interface{}{
		"jwt":  jwtToken,
		"role": role,
	}

	secret, err := client.Logical().Write("auth/jwt/login", data)
	if err != nil {
		return "", fmt.Errorf("failed to login with JWT token: %v", err)
	}

	if secret == nil || secret.Auth == nil || secret.Auth.ClientToken == "" {
		return "", fmt.Errorf("failed to retrieve client token from Vault")
	}

	return secret.Auth.ClientToken, nil
}

func getVaultTokenWithUserPass(vaultAddress, username, password string) (string, error) {
	config := &vault.Config{
		Address: vaultAddress,
	}

	client, err := vault.NewClient(config)
	if err != nil {
		return "", fmt.Errorf("failed to create vault client: %v", err)
	}

	data := map[string]interface{}{
		"password": password,
	}

	secret, err := client.Logical().Write(fmt.Sprintf("auth/userpass/login/%s", username), data)
	if err != nil {
		return "", fmt.Errorf("failed to login with username and password: %v", err)
	}

	if secret == nil || secret.Auth == nil || secret.Auth.ClientToken == "" {
		return "", fmt.Errorf("failed to retrieve client token from Vault")
	}

	return secret.Auth.ClientToken, nil
}

func GetClient(c *gin.Context) *vault.Client {
	client, _ := c.Get("vaultClient")
	vaultClient := client.(*vault.Client)
	return vaultClient
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Auth -> Vault

		vault_adress := "https://secure-env.poc-innovation.com:8200/"

		for name, values := range c.Request.Header {
			fmt.Printf("%s: %s\n", name, values)
		}

		client_token := ""
		auth_type := c.GetHeader("X-auth-type")

		if auth_type == "root-token" {
			client_token = c.GetHeader("X-auth-root-token")
		} else if auth_type == "GitHub" {
			client_token, _ = getVaultTokenWithGitHubAuth(vault_adress, c.GetHeader("X-auth-github-token"))
		} else if auth_type == "JWT" {
			client_token, _ = getVaultTokenWithJWT(vault_adress, "", c.GetHeader("X-auth-jwt-token"))
		}

		config := vault.DefaultConfig()

		config.Address = vault_adress

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
