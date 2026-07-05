# editstream
An event-sourced Wikipedia edit pipeline.

A single Kafka log is the source of truth, a consumer folds it into a live per-wiki leaderboard, and can rebuild that leaderboard deterministically by replaying the log from offset 0.
## Architecture

```
editstream/
├── docker-compose.yml          # KRaft broker (cp-kafka) + topic init
|
├── internal/
│   ├── edit/
│   │   ├── edit.go             # Edit type (Wiki, User, Type, Bot, ByteDelta, ...)
│   │   ├── parse.go            # ParseEdit: recentchange json bytes -> Edit
│   │   ├── parse_test.go       # table-driven tests for ParseEdit
│   │   ├── tally.go            # Tally: the fold (Apply) + TopWikis leaderboard
│   │   └── tally_test.go       # fold + replay test
│   │
│   └── stream/                 # SSE ingestion
│       ├── sse.go              # Client.Run
│       └── sse_test.go         # table-driven test for dataPayload
│
└── cmd/
    ├── producer/
    │   └── main.go             # SSE -> ParseEdit (validate) -> append raw, keyed by wiki
    └── consumer/
        └── main.go             # edits -> Tally (fold) -> leaderboard, --replay flag
```

## Prerequisites

- Docker
- Go 1.23+
- A C toolchain, confluent-kafka-go uses cgo

## Quick Start

### 1. Clone and Setup

```bash
git clone https://github.com/rustpill/editstream
cd editstream
```

### 2. Start docker

```bash
docker compose up -d
docker compose logs kafka-init      # sanity check
```

### 3. Start producer and consumer

```bash
go run ./cmd/producer               # Terminal 1

go run ./cmd/consumer               # Terminal 2
```

## Replay demo
Let the producer run a while before starting consumer with replay flag, it will catch up to current offset.
```bash
go run ./cmd/consumer --replay
```

## Tests
Only edit and stream packages have written tests.
```bash
go test ./internal/...
```