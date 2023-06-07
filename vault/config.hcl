storage "file" {
  path = "/vault/data/"
}

listener "tcp" {
 address     = "0.0.0.0:8200"
 tls_cert_file = "/vault/certs/cert.pem"
 tls_key_file = "/vault/certs/key.pem"
}

ui = true
disable_mlock = true
api_addr = "https://secure-env.poc-innovation.com:8200/"