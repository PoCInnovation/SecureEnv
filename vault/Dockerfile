FROM ubuntu:22.04

WORKDIR /vault

RUN apt update -y && \
    # GPG is required for the package signing key
    apt install wget gpg -y && \
    # Download the signing key to a new keyring
    wget -O- https://apt.releases.hashicorp.com/gpg | gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg && \
    # Verify the key's fingerprint
    gpg --no-default-keyring --keyring /usr/share/keyrings/hashicorp-archive-keyring.gpg --fingerprint && \
    # Add the HashiCorp repo
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com focal main" | tee /etc/apt/sources.list.d/hashicorp.list && \
    # Update
    apt update -y && \
    # Install vault
    apt install vault jq -y

VOLUME config/
VOLUME certs/
VOLUME data/

COPY config.hcl config/
COPY start.sh .
COPY key.pem certs/
COPY cert.pem certs/

EXPOSE 8200

CMD vault server -config=/vault/config/config.hcl