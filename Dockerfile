# Stage 1: Build stage
FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN GOOS=linux go build -o myapp .

FROM ubuntu:24.04


RUN apt-get update && apt-get install -y \
    tzdata \
    scowl \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Set the working directory
WORKDIR /app

# Copy the binary from the build stage
COPY --from=builder /app/myapp .

# Set the entrypoint command
ENTRYPOINT ["/app/myapp"]