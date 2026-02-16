# Prochat
[![CI](https://github.com/varsotech/prochat-server/actions/workflows/ci.yaml/badge.svg)](https://github.com/varsotech/prochat-server/actions/workflows/ci.yaml)

**A federated chat and VoIP platform for communities.**

## Goals
* Low-latency text and voice chat
* Federated identity
* Accessible user experience for non-technical people.
* Easy and cheap to host for small communities.

## Non-goals
* Data replication in homeserver


## Contributing
### Prerequisites

1. Golang, with `~/.local/bin` added to PATH
2. `make`, `docker`, `docker-compose`
3. Run `make install-deps`
4. Copy .env.template to .enc

### Run server
```bash
docker compose up -d
go run ./cmd/prochat
```