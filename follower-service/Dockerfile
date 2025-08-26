FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o follower-service

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/follower-service .
EXPOSE 8082
ENTRYPOINT [ "./follower-service" ]