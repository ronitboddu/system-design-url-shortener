FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN go build -o url-shortener-server ./cmd/api/main.go


FROM debian:stable-slim

WORKDIR /app

COPY --from=builder /app/url-shortener-server .

EXPOSE 8080

CMD ["./url-shortener-server"]