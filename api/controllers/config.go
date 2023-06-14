package controllers

import "os"

const adr_vault = "https://secure-env.poc-innovation.com:8200"

var token string = os.Getenv("TOKEN_VPS_VAULT")
