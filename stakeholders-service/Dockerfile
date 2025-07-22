FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod go.sum .
RUN go mod download
COPY . .
RUN go build -o stakeholders-service

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/stakeholders-service .
EXPOSE 8080
ENTRYPOINT ["./stakeholders-service"]