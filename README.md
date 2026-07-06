# s3lite

A minimal local object storage that chunks files, replicates chunks across storage nodes, tracks metadata via JSON manifests, and exposes operations via an HTTP API.

## Quick Start

### CLI

```bash
go build -o s3lite ./cmd/s3lite

./s3lite upload myfile.txt
./s3lite list
./s3lite stat myfile.txt
./s3lite download myfile.txt ./downloads
./s3lite delete myfile.txt
```

### Metadata Server

```bash
go build -o metadata-server ./cmd/metadata

./metadata-server
# Listening on :8080
```

## Commands / API

### CLI

| Command | Args | Description |
|---------|------|-------------|
| `upload <file>` | path to file | Chunks file, replicates chunks to nodes, saves manifest |
| `download <file> <dest>` | object name, output dir | Reconstructs file from chunks, verifies checksums |
| `list` | — | Lists all stored objects |
| `stat <file>` | object name | Shows object metadata and per-chunk details |
| `delete <file>` | object name | Removes all chunk replicas and metadata |

### HTTP Endpoints

The metadata server (`cmd/metadata`) exposes the same operations over HTTP on port `8080`:

| Endpoint | Description |
|----------|-------------|
| `GET /health` | Health check |
| `POST /upload` | Upload a file |
| `GET /download` | Download a file |
| `GET /list` | List stored objects |
| `GET /stat` | Show object metadata |
| `DELETE /delete` | Delete an object |

## Architecture

Nodes are monitored via a periodic heartbeat goroutine. If a node goes down, the recovery process re-replicates its chunks onto remaining healthy nodes to maintain the replication factor.

```
                    ┌──────────────┐
                    │  metadata/   │
                    │  *.json      │
                    └──────┬───────┘
                           │
┌──────────┐   CLI / HTTP    ┌──────────────┐
│  s3lite  │ ◄────────────► │   Storage    │
│  CLI /   │                +──────────────┤
│  Server  │                │ Root         │
└──────────┘                │ ChunkSize    │
                             │ Replica      │
                             │ Nodes[]      │
                             │ Heartbeat ◄──┤── goroutine (every 2s)
                             │ Recovery  ◄──┤── goroutine
                             └──────┬───────┘
                                     │
                     ┌───────────────┼──────────┐
                     │               │          │
               ┌─────▼──────┐ ┌─────▼──────┐ ┌──▼───────┐
               │  node1/    │ │  node2/    │ │  node3/  │
               │  chunks/   │ │  chunks/   │ │  chunks/ │
               └────────────┘ └────────────┘ └──────────┘
```

### Data flow

**Upload:** source → 1 MiB chunks → SHA-256 checksum → replicate to 2 of 3 nodes → save JSON manifest to `metadata/`.

**Download:** load manifest → try replicas in order → verify SHA-256 → first healthy replica wins → write reconstructed file.

**Delete:** load manifest → remove all replica chunks → remove manifest. Missing chunks are silently tolerated.

**Recovery:** background goroutine scans all manifests → for each chunk, checks if enough healthy replicas exist → re-replicates onto alive nodes that don't have it yet.

## Configuration

Hardcoded at the moment (see `cmd/s3lite/main.go` and `cmd/metadata/main.go`):

- **Root:** `storage/` (created at runtime)
- **Chunk size:** 1 MiB
- **Replication factor:** 2
- **Heartbeat interval:** 2 seconds
- **Recovery interval:** 5 seconds
- **Nodes:** `node1`, `node2`, `node3` under root
- **Metadata server port:** `:8080`

## Storage layout

```
storage/
├── metadata/
│   └── <filename>.json
├── node1/
│   └── chunks/
│       └── <uuid>.chunk
├── node2/
│   └── chunks/
│       └── <uuid>.chunk
└── node3/
    └── chunks/
        └── <uuid>.chunk
```

## Dependencies

Only [google/uuid](https://github.com/google/uuid) v1.6.0. Everything else is Go standard library.

## License

MIT
