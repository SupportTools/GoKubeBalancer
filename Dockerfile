# Use a full-featured base image for building with the correct Go version
FROM golang:1.22.0-alpine3.18 AS builder

# Install git and ca-certificates, necessary for go mod and secure connections respectively
RUN apk update && apk add --no-cache git ca-certificates

# Set the Current Working Directory inside the container
WORKDIR /src

# Copy the source code into the container
COPY . .

# Fetch dependencies using go mod if your project uses Go modules
RUN go mod download

# Version and Git Commit build arguments
ARG VERSION
ARG GIT_COMMIT
ARG BUILD_DATE

# Build the Go app with versioning information
RUN GOOS=linux GOARCH=amd64 go build -ldflags "-X github.com/supporttools/GoKubeBalancer/pkg/health.Version=$VERSION -X github.com/supporttools/GoKubeBalancer/pkg/health.GitCommit=$GIT_COMMIT -X github.com/supporttools/GoKubeBalancer/pkg/health.BuildTime=$BUILD_DATE" -o /bin/GoKubeBalancer

# Use Distroless as a runtime base
FROM gcr.io/distroless/static

# Set the working directory to /app
WORKDIR /app

# Copy the built binary and config file from the builder stage
COPY --from=builder /bin/GoKubeBalancer /app/

# Copy necessary CA certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Use an unprivileged user to run the application
USER nonroot:nonroot

# Set the binary as the entrypoint of the container
ENTRYPOINT ["/app/GoKubeBalancer"]
