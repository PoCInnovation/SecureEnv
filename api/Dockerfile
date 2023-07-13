FROM golang:1.20

WORKDIR /api

COPY . .

RUN go get

EXPOSE 8080

CMD go run main.go