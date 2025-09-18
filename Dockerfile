FROM golang:alpine AS builder

ENV CGO_ENABLED=0

# Install git for fetching the dependencies
RUN apk update && apk add --no-cache build-base git curl bash 

# Setup folders
RUN mkdir /app
WORKDIR /app

COPY go.sum ./
COPY go.mod ./
RUN go mod download

COPY . .

RUN curl -sSL "https://github.com/bufbuild/buf/releases/download/v1.57.0/buf-$(uname -s)-$(uname -m)" -o "/usr/local/bin/buf" && chmod +x "/usr/local/bin/buf"

RUN go generate ./...

# Build the Go app
RUN --mount=type=cache,target=/root/.cache/go-build go build -o /build ./cmd/prochat

FROM alpine:latest AS goapp
WORKDIR /root/
COPY --from=builder /build .
CMD ["./build"]