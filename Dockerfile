# Use the official Golang image as the base image - https://hub.docker.com/_/golang/tags
FROM golang:1.20-alpine as builder

# Set the working directory
WORKDIR /app

# Copy source to this builder container
COPY . /app

# Download build dependencies
RUN go get github.com/codingo/dorky

# Build the application
RUN go build -o main .

# Bump the application with --help to know it was built
RUN /app/main --help


# Use an Alpine container to deploy/run the application from - https://hub.docker.com/_/alpine/tags
FROM alpine:3.18

WORKDIR /app
COPY --from=builder /app/main /app/main

ENTRYPOINT ["/app/main"]
