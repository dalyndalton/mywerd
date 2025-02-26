FROM ubuntu:24.04

ARG BINARY

RUN apt-get update && apt-get install -y \
    tzdata \
    scowl \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Set the working directory
WORKDIR /app

# Copy the binary into the image
COPY ${BINARY} /app/myapp

# Set the entrypoint command
ENTRYPOINT ["/app/myapp"]
