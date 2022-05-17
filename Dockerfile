############################
# Use the offical Golang image to create a build artifact.
# This is based on Debian and sets the GOPATH to /go.
# https://hub.docker.com/_/golang
############################
FROM golang:1.18-alpine as builder

# Arguments
ARG GID=2000

# Create an user group
RUN addgroup --gid ${GID} appgroup

# Create non root user
RUN adduser -D -g ${GID} appuser

# Create app directory.
WORKDIR /usr

# Copy go.mod & go.sum files
COPY go.mod go.sum ./

# Install app dependencies.
RUN go mod download

# Copy local code to the container image.
COPY . .

# Build the outyet command inside the container.
RUN CGO_ENABLED=0 go build -o app

############################
# Use an Alpine image to obtain CA certificates and some binaries
############################
FROM alpine:latest as utils

RUN apk --no-cache add ca-certificates

# Health check binary
RUN GRPC_HEALTH_PROBE_VERSION=v0.4.11 && \
    wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
    chmod +x /bin/grpc_health_probe

############################
# Use a Docker multi-stage build to create a lean production image.
############################
FROM scratch

# Enviroment variables.
ENV PORT=8080

# Copy the ca certificates certs stage.
COPY --from=utils /bin/grpc_health_probe /bin/grpc_health_probe

# Copy the ca certificates certs stage.
COPY --from=utils /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# Copy the user and group files from the builder stage.
COPY --from=builder /etc/passwd /etc/passwd

# Copy the binary to the production image from the builder stage.
COPY --from=builder /usr/app /go/bin/app

# Expose port.
EXPOSE $PORT

# Use an unprivileged user.
USER appuser

# Run command.
ENTRYPOINT ["/go/bin/app"]
