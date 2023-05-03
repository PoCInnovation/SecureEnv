FROM vault:latest

WORKDIR /app

ENV VAULT_DEV_ROOT_TOKEN_ID=dev-only-token
ENV VAULT_DEV_LISTEN_ADDRESS=0.0.0.0:8200

EXPOSE 8200
EXPOSE 8201

CMD ["vault", "server", "-dev"]
