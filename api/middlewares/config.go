package middlewares

import "os"

var adr_vault string = os.Getenv("SECURE_ENV_HOST")
