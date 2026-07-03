# s3lite

A minimal local object storage that chunks files, replicates chunks across storage nodes, and tracks metadata via JSON manifests. No daemon, no network — pure filesystem CLI tool.

## Quick Start

```bash
go build -o s3lite ./cmd/s3lite

./s3lite upload myfile.txt
./s3lite list
./s3lite stat myfile.txt
./s3lite download myfile.txt ./downloads
./s3lite delete myfile.txt
```

## Commands

| Command | Args | Description |
|---------|------|-------------|
| `upload <file>` | path to file | Chunks file, replicates chunks to nodes, saves manifest |
| `download <file> <dest>` | object name, output dir | Reconstructs file from chunks, verifies checksums |
| `list` | — | Lists all stored objects |
| `stat <file>` | object name | Shows object metadata and per-chunk details |
| `delete <file>` | object name | Removes all chunk replicas and metadata |

## Architecture

Nodes are monitored via a periodic heartbeat goroutine. If a node goes down, the recovery process re-replicates its chunks onto remaining healthy nodes to maintain the replication factor.

```
                    ┌──────────────┐
                    │  metadata/   │
                    │  *.json      │
                    └──────┬───────┘
                           │
┌──────────┐    upload/download/delete    ┌──────────────┐
│  s3lite  │ ◄──────────────────────────► │   Storage    │
│   CLI    │                              +──────────────┤
└──────────┘                              │ Root         │
                                          │ ChunkSize    │
                                          │ Replica      │
                                          │ Nodes[]      │
                                          │ Heartbeat ◄──┤── goroutine (every 2s)
                                          │ Recovery  ◄──┤── goroutine
                                          └──────┬───────┘
                                                  │
                    ┌─────────────────────────────┼──────────┐
                    │                             │          │
              ┌─────▼──────┐              ┌──────▼──────┐ ┌──▼───────┐
              │  node1/    │              │   node2/    │ │  node3/  │
              │  chunks/   │              │   chunks/   │ │  chunks/ │
              └────────────┘              └─────────────┘ └──────────┘
```

### Data flow

**Upload:** source → 1 MiB chunks → SHA-256 checksum → replicate to 2 of 3 nodes → save JSON manifest to `metadata/`.

**Download:** load manifest → try replicas in order → verify SHA-256 → first healthy replica wins → write reconstructed file.

**Delete:** load manifest → remove all replica chunks → remove manifest. Missing chunks are silently tolerated.

**Recovery:** background goroutine scans all manifests → for each chunk, checks if enough healthy replicas exist → re-replicates onto alive nodes that don't have it yet.

## Configuration

Hardcoded at the moment (see `cmd/s3lite/main.go`):

- **Root:** `storage/` (created at runtime)
- **Chunk size:** 1 MiB
- **Replication factor:** 2
- **Heartbeat interval:** 2 seconds
- **Nodes:** `node1`, `node2`, `node3` under root

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
