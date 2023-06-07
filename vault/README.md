# Secure-Env Docker Image: Vault Production Deployment with HTTPS 

This Docker image enables you to deploy a HashiCorp Vault instance for production purposes with HTTPS support. The image is designed to be used in conjunction with the Secure-Env project.

## Prerequisites

Prior to launching your Vault, you'll need to place your own `key.pem` and `cert.pem` files within the Vault directory. These are critical for securing your Vault instance.

Next, validate that the following values within the listed files match your host:

- `api_addr` in the `config.hcl` file
- `VAULT_ADDR` in the `start.sh` script

## Building and Running the Docker Image

You can build the Docker image for the Secure-Env project by using the following command:

```bash
docker build -t secure-env .
```

This command will create an image tagged as 'secure-env'.

To run the Docker container, use:

```bash
docker run -d --name vault-instance -p8200:8200 secure-env
```

This will launch the Vault instance as a Docker container, named 'vault-instance', and map port 8200 from the container to the host.

## Initializing and Unsealing the Vault

To initialize and unseal your Vault instance each time you start or restart it, use the following command:

```bash
docker exec vault-instance /vault/start.sh
```

This command runs the `start.sh` script inside the running 'vault-instance' Docker container, performing necessary initializations and unsealing operations on the Vault.

Remember to follow these steps carefully to ensure the security and proper functioning of your Vault instance.
