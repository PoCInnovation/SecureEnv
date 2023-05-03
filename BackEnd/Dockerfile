# Utilise l'image officielle de Go en version 1.16
FROM golang:1.20.2

# Définit le répertoire de travail
WORKDIR /app

# Copie les fichiers de l'application dans le conteneur
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . .

# Build l'application
RUN go build -o main .

# Expose le port 8080
EXPOSE 8080

# Démarre l'application
CMD [ "./main" ]
