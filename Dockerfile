FROM golang:alpine AS builder

ENV CGO_ENABLED=0

# Install git for fetching the dependencies
RUN apk update && apk add --no-cache git && apk add --no-cach bash && apk add build-base

# Setup folders
RUN mkdir /app
WORKDIR /app

COPY go.sum ./
COPY go.mod ./
RUN go mod download

COPY . .

RUN go generate ./...

# Build the Go app
RUN --mount=type=cache,target=/root/.cache/go-build go build -o /build ./cmd/prochat

FROM alpine:latest AS goapp
WORKDIR /root/
COPY --from=builder /build .
CMD ["./build"]