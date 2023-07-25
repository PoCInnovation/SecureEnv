package main

func env_exists(dict map[string]string, key string) bool {
	_, exists := dict[key]
	return exists
}

// fonction qui en fonction de la
func get_auth_headers(env map[string]string) map[string]string {
	var auth_headers map[string]string

	if env_exists(env, "SECURE_ENV_AUTH_GITHUB_TOKEN") {
		auth_headers["X-auth-type"] = "GitHub"
		auth_headers["X-auth-github-token"] = env["SECURE_ENV_AUTH_GITHUB_TOKEN"]
	} else if env_exists(env, "SECURE_ENV_AUTH_JWT_TOKEN") {
		auth_headers["X-auth-type"] = "JWT"
		auth_headers["X-auth-jwt-token"] = env["SECURE_ENV_AUTH_JWT_TOKEN"]
	} else if env_exists(env, "SECURE_ENV_AUTH_USER") && env_exists(env, "SECURE_ENV_AUTH_PASS") {
		auth_headers["X-auth-type"] = "UserPass"
		auth_headers["X-auth-user"] = env["SECURE_ENV_AUTH_USER"]
		auth_headers["X-auth-pass"] = env["SECURE_ENV_AUTH_PASS"]
	}

	return auth_headers
}
