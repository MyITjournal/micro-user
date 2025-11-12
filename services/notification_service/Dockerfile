# Builder stage
FROM golang:1.24-bullseye AS builder

WORKDIR /app

COPY go.mod go.sum ./

ENV GOPROXY=https://proxy.golang.org,direct

# Install librdkafka for confluent-kafka-go
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        ca-certificates git \
        librdkafka-dev && \
    rm -rf /var/lib/apt/lists/*

RUN go mod download

COPY . .

# Remove CGO_DISABLED=0 (needs CGO)
RUN go build -o notification-email-service main.go




# FROM golang:1.24-bullseye AS builder

# WORKDIR /app

# COPY go.mod go.sum ./
# RUN go mod download

# COPY . .

# RUN CGO_ENABLED=1 go build -o notification-email-service main.go

# FROM debian:bullseye-slim
# WORKDIR /app
# COPY --from=builder /app/notification-email-service .
# COPY --from=builder /app/.env .

# CMD ["./notification-email-service"]
