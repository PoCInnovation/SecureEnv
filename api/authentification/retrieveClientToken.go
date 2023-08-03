package authentification

import (
	"fmt"

	vault "github.com/hashicorp/vault/api"
)

func getVaultTokenWithGitHubAuth(vaultAddress, githubToken string) (string, error) {
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

func getVaultTokenWithJWT(vaultAddress, role, jwtToken string) (string, error) {
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
