FROM ubuntu:22.04

WORKDIR /vault

RUN apt update -y && apt install wget gpg -y && \
    wget -O- https://apt.releases.hashicorp.com/gpg | gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg && \
    gpg --no-default-keyring --keyring /usr/share/keyrings/hashicorp-archive-keyring.gpg --fingerprint && \
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com focal main" | tee /etc/apt/sources.list.d/hashicorp.list && \
    apt update -y && \
    apt install vault jq -y

VOLUME /vault/config/
VOLUME /vault/certs/
VOLUME /vault/data/

COPY config.hcl /vault/config/
COPY unseal-script.sh /vault/
COPY key.pem /vault/certs/
COPY cert.pem /vault/certs/

EXPOSE 8200

CMD vault server -config=/vault/config/config.hcl