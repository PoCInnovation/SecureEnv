package middlewares

import "os"

var adr_vault string = os.Getenv("SECURE_ENV_HOST")

var token string = os.Getenv("SECURE_ENV_TOKEN")
