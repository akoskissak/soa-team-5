FROM golang:1.24.5-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o api-gateway

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/api-gateway .
EXPOSE 8080
ENTRYPOINT [ "./api-gateway" ]