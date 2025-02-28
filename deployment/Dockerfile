FROM golang:1.22-alpine AS build

WORKDIR /app

COPY ./ .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o shortener cmd/shortener/main.go

FROM alpine:latest

EXPOSE 8000

WORKDIR /app

COPY --from=build /shortener .

ENV CONFIG_PATH=config.yaml

CMD ["./app/shortener"]