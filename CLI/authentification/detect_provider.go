package authentification

import (
	"os"
)

func get_auth_headers() map[string]string {
	var auth_headers map[string]string

	if val, present := os.LookupEnv("SECURE_ENV_AUTH_GITHUB_TOKEN"); present == true {
		auth_headers["X-auth-type"] = "GitHub"
		auth_headers["X-auth-github-token"] = val
	} else if val, present := os.LookupEnv("SECURE_ENV_AUTH_JWT_TOKEN"); present == true {
		auth_headers["X-auth-type"] = "JWT"
		auth_headers["X-auth-jwt-token"] = val
	} else if val, present := os.LookupEnv("SECURE_ENV_AUTH_ROOT_TOKEN"); present == true {
		auth_headers["X-auth-type"] = "root-token"
		auth_headers["X-auth-root-token"] = val
	}

	return auth_headers
}
