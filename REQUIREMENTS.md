# Requirements: s3lite Architecture Fixes
**Date:** 2026-07-13
**Requested by:** Architecture Review
**Status:** DRAFT

## Summary
Formalization of 17 issues identified in the s3lite architecture review into testable, prioritized requirements. Each requirement includes acceptance criteria, dependencies, and verification strategy.

## Requirements by Priority

### P0 — Critical Bugs

#### R1: CLI stat Command Argument Validation
**File:** `cmd/s3lite/main.go:79`
**Issue:** `stat` command panics with `index out of range` when invoked with no arguments.

- **AC-1:** Given the CLI is invoked with `s3lite stat` (no filename), When execution completes, Then it prints `Usage: stat <file>` and exits with code 1 (or non-zero).
- **AC-2:** Given `s3lite stat filename`, When the file exists, Then stat info is printed and exit code is 0.
- **AC-3:** Given `s3lite stat extra args`, When more than 2 args are present, Then it prints usage and exits.
- **Test:** Run `s3lite stat` in isolation — must not panic, must print usage.
- **Depends on:** None.

#### R2: Metadata Server Signal Handling Dead Code
**File:** `cmd/metadata/main.go:31-37`
**Issue:** Signal handler code after `log.Fatal(http.ListenAndServe(...))` is unreachable.

- **AC-1:** Given the metadata server starts, When SIGINT/SIGTERM is received, Then the server performs graceful shutdown and logs "Shutting down Metadata Server...".
- **AC-2:** Given the server is running, When the listener fails, Then the error is logged and the process exits.
- **Test:** Start the server, send `kill -SIGTERM <pid>`, verify log output and clean exit.
- **Depends on:** None.

#### R3: Double Heartbeat Start
**File:** `internal/storage/storage.go:21` + `cmd/metadata/main.go:20`
**Issue:** `storage.New()` starts a heartbeat, then `cmd/metadata/main.go` starts another via `s.StartHeartBeat()`.

- **AC-1:** Given `storage.New()` is called exactly once, When the object is returned, Then exactly one heartbeat goroutine is running (verified by ticker count or flag).
- **AC-2:** Given `cmd/metadata/main.go` calls `storage.New()`, When it does NOT also call `StartHeartBeat()`, Then heartbeat still runs exactly once.
- **Test:** Instrument heartbeat start (counter or log line); create `New()`, assert counter == 1.
- **Depends on:** None.

#### R4: StatHandler JSON Response Corruption
**File:** `internal/metadata/handlers.go:111`
**Issue:** `w.Write([]byte("Stat endpoint"))` appends literal text after JSON, producing invalid response body.

- **AC-1:** Given a valid `GET /stat?file=foo`, When the response is received, Then the body is parseable JSON with no trailing non-JSON text.
- **AC-2:** Given the same request, When inspecting the response, Then `Content-Type` is `application/json` and body contains exactly the JSON-encoded metadata object.
- **Test:** `curl http://localhost:8080/stat?file=testfile` — pipe through `python3 -m json.tool` (must succeed). Assert no trailing bytes.
- **Depends on:** None.

#### R5: Recovery Writes Nil Data When All Replicas Dead
**File:** `internal/storage/recovery.go:130`
**Issue:** If all replicas are dead/unhealthy, `sourceData` remains nil, and `os.WriteFile` writes nil/empty to disk.

- **AC-1:** Given a chunk with all replicas unreachable, When `RecoverChunk` is called, Then it returns an error indicating no source data available and writes nothing to disk.
- **AC-2:** Given a chunk with at least one healthy replica, When recovery runs, Then valid data is written to target nodes and replicas list is updated.
- **Test:** Mock all nodes as dead for a specific chunk; assert `RecoverChunk` returns error, assert no file created on disk at target node path.
- **Depends on:** None.

---

### P1 — Security

#### R6: Path Traversal in Upload Filename
**File:** `internal/metadata/handlers.go:38`
**Issue:** `header.Filename` is used directly in `filepath.Join`, allowing `../../etc/passwd` style traversal.

- **AC-1:** Given an upload with filename `../../etc/cron.d/evil`, When the handler processes it, Then the filename is sanitized (e.g., `filepath.Base()` applied) and the file is written only within the temp directory, not to an absolute or parent-relative path.
- **AC-2:** Given an upload with a normal filename `photo.jpg`, When processed, Then it works identically to before (no regression).
- **Test:** Craft a multipart upload with `filename: ../../tmp/pwned`, assert the resulting file path does not escape the temp directory. Assert the final stored file name is just `pwned` (basename only).
- **Depends on:** None.

#### R7: No HTTP Method Validation on Handlers
**File:** `internal/metadata/handlers.go`
**Issue:** All handlers accept any HTTP method (GET, POST, DELETE, etc.) without restriction.

- **AC-1:** Given `POST /stat?file=foo` (wrong method for a read endpoint), When the handler receives it, Then it responds with `405 Method Not Allowed`.
- **AC-2:** Given `GET /upload` (wrong method for an upload endpoint), When received, Then it responds with `405 Method Not Allowed`.
- **AC-3:** Given correct methods (`GET /stat`, `POST /upload`, `GET /download`, `GET /list`, `DELETE /delete` or appropriate), When received, Then handlers function normally.
- **Test:** For each endpoint, send wrong method, assert 405. Send correct method, assert normal operation.
- **Depends on:** None.

---

### P2 — Functional

#### R8: InitializeNodes Runs Every Upload
**File:** `internal/storage/upload.go:12`
**Issue:** `s.InitializeNodes()` (which calls `os.MkdirAll` for every node) runs on every upload call.

- **AC-1:** Given a `Storage` instance, When `Upload()` is called, Then `InitializeNodes()` is NOT called (or is called only once via a sync.Once or startup path).
- **AC-2:** Given nodes are already initialized, When `Upload()` runs, Then no `os.MkdirAll` calls are made (verify via strace/dtrace or mock).
- **Test:** Mock `os.MkdirAll`; upload twice; assert call count ≤ 1 (ideally 0 if initialized at construction).
- **Depends on:** R3 (if `New()` handles initialization, this becomes moot).

#### R9: Missing Error Checks in UploadHandler
**File:** `internal/metadata/handlers.go:35,39`
**Issue:** `io.Copy` and `os.Rename` return values are ignored in `UploadHandler`.

- **AC-1:** Given a disk-full or permission-denied scenario for `io.Copy`, When upload is attempted, Then a `500 Internal Server Error` response is returned with a meaningful error message.
- **AC-2:** Given a failing `os.Rename` (e.g., destination exists), When upload is attempted, Then a `500` is returned, temp file is cleaned up.
- **Test:** Mock `io.Copy` to return error; assert HTTP 500. Mock `os.Rename` to return error; assert HTTP 500 and no stale temp files.
- **Depends on:** None.

#### R10: Download Reads Entire Chunk Into Memory
**File:** `internal/storage/download.go:38`
**Issue:** `os.ReadFile` loads each chunk fully into memory; problematic for large chunks/files.

- **AC-1:** Given a download of a file with chunks, When chunks are read, Then they are streamed to the output file using `io.Copy` (or equivalent streaming read) rather than `os.ReadFile` loading entire chunk into a `[]byte`.
- **AC-2:** Given the same download, When complete, Then the output file is byte-identical to the original upload (checksum verified).
- **Test:** Upload a 10MB file, download it, compare checksums. Optionally, monitor memory usage during download (should not spike to full file size).
- **Depends on:** None.

#### R11: Download Handler Cleanup Race with ServeFile
**File:** `internal/metadata/handlers.go:59-74`
**Issue:** `defer os.RemoveAll(tmpDir)` may execute before `http.ServeFile` finishes writing the response.

- **AC-1:** Given a download request, When `ServeFile` is writing the response, Then the temp directory is NOT deleted until after the response is fully sent.
- **AC-2:** Given the same request, When the handler returns, Then the temp directory is cleaned up (no leak).
- **Test:** Download a large file (slow enough to observe timing); verify temp directory exists during transfer and is gone after completion. Alternatively, verify via code analysis that `os.RemoveAll` is called in a callback or after `ServeFile` returns.
- **Depends on:** None.

---

### P3 — Code Quality

#### R12: Hardcoded Replication Factor
**File:** `internal/storage/storage.go:18`
**Issue:** `ReplicationFactor` is hardcoded to `2` with no way to configure it.

- **AC-1:** Given `storage.New(root, chunkSize)`, When called without specifying replication factor, Then it defaults to 2 (backward compatible).
- **AC-2:** Given a `Storage` instance, When `ReplicationFactor` is set via constructor option, function parameter, or config, Then that value is used for replication decisions.
- **Test:** Create storage with default (assert RF=2). Create with RF=3, upload a file, assert 3 replicas exist per chunk.
- **Depends on:** None.

#### R13: First Heartbeat Delayed by Full Interval
**File:** `internal/storage/heartbeat.go:26`
**Issue:** `<-ticker.C` is at the end of the loop body, so the first heartbeat fires after one full interval instead of immediately.

- **AC-1:** Given `StartHeartBeat(interval)`, When called, Then the first heartbeat check runs immediately (or within a negligible time, e.g., < 100ms).
- **AC-2:** Given subsequent heartbeats, When they run, Then they occur at the configured interval.
- **Test:** Start heartbeat with 1s interval; assert node liveness is set within 100ms of start (not after 1s).
- **Depends on:** None.

#### R14: Directory Permissions 0777
**File:** `internal/storage/node.go:38`
**Issue:** `os.ModePerm` (0777) is used for `MkdirAll`; should be 0755.

- **AC-1:** Given `InitializeNodes()`, When directories are created, Then permissions are `0755` (rwxr-xr-x), not `0777`.
- **AC-2:** Given the same function, When the metadata directory is created, Then its permissions are also `0755`.
- **Test:** Run `InitializeNodes()`, then `stat -c '%a' <dir>` on each created directory; assert output is `755`.
- **Depends on:** None.

#### R15: Silent Chunk Deletion Errors
**File:** `internal/storage/delete.go:24`
**Issue:** `_ = os.Remove(chunkPath)` silently discards errors during chunk deletion.

- **AC-1:** Given a `Delete()` call, When a chunk file cannot be removed (e.g., permission denied), Then the error is collected and returned (or at minimum logged) — not silently swallowed.
- **AC-2:** Given a successful delete, When all chunk files and metadata are removed, Then no error is returned.
- **Test:** Make one chunk file read-only; call `Delete()`; assert error is returned indicating which chunk failed. Assert metadata file is NOT deleted if chunk deletion failed (or document the partial-delete behavior).
- **Depends on:** None.

#### R16: No Atomic Writes for Replication
**File:** `internal/storage/replication.go`
**Issue:** `os.WriteFile` writes directly to final path; a crash mid-write leaves a corrupt partial chunk.

- **AC-1:** Given a chunk replication, When writing to a node, Then the write targets a temporary file first, then is atomically renamed to the final path.
- **AC-2:** Given a crash during write, When the node recovers, Then no corrupt partial chunk file exists at the final path (only temp file or complete file).
- **Test:** Mock `os.Rename` to fail after temp file creation; assert no file at final path. Normal path: upload, verify final file exists and is correct.
- **Depends on:** None.

#### R17: No File Locking for Concurrent Metadata Access
**File:** `internal/storage/metadata.go`
**Issue:** `SaveMetadata`/`LoadMetadata` have no locking; concurrent uploads of different files to the same metadata directory, or concurrent reads/writes of the same metadata file, could race.

- **AC-1:** Given two concurrent `SaveMetadata` calls for different files, When both complete, Then both metadata files are valid JSON and complete (no truncation).
- **AC-2:** Given a concurrent `SaveMetadata` and `LoadMetadata` for the same file, When both complete, Then `LoadMetadata` returns either the old or new complete metadata (no partial/corrupt JSON).
- **Test:** Run N concurrent uploads (different files) in goroutines; assert all metadata JSON files are valid. Run concurrent read+write for same file in goroutines; assert no panic, no corrupt JSON.
- **Depends on:** None.

---

## Dependencies (DAG)

```
R3 (Double Heartbeat) ──┐
                        ▼
R8 (InitNodes Every Upload)
  If New() handles init, both R3 and R8 become "move side effects to New()" — resolved together.

All other requirements are independent; can be implemented in any order.
```

## Testing Strategy

| Priority | Verification Method |
|----------|-------------------|
| **P0 (R1-R5)** | Unit tests + integration tests. Each fix must have a regression test that fails before the fix and passes after. Run with `go test ./...` and manual CLI/server verification. |
| **P1 (R6-R7)** | Security-focused unit tests with crafted inputs. Fuzz testing for path traversal. HTTP method matrix testing for all endpoints. |
| **P2 (R8-R11)** | Functional tests with real files. Upload/download round-trip with checksum verification. Race detector (`go test -race`) for R11. |
| **P3 (R12-R17)** | Unit tests. `go test -race` for R17. Permission checks via `os.Stat` for R14. Strace/dtrace for R8 (optional). |

## Implementation Order (Recommended)

```
Phase 1 — P0 Criticals (must fix first):
  R1 → R3 → R4 → R5 → R2

Phase 2 — Security (before any public exposure):
  R6 → R7

Phase 3 — Functional (improve reliability):
  R9 → R11 → R10 → R8

Phase 4 — Quality (hardening):
  R14 → R15 → R16 → R13 → R12 → R17
```

## Open Questions

- **OQ-1 (R7):** What are the intended HTTP methods for each endpoint? The document assumes: `GET` for read operations (stat, list, download), `POST` for upload, `DELETE` for delete. Is this correct?
- **OQ-2 (R15):** Should chunk deletion errors be returned to the caller, logged and ignored, or should partial deletion be prevented (all-or-nothing)?
- **OQ-3 (R17):** Should file locking be per-metadata-file (fine-grained) or per-metadata-directory (coarse-grained)? The current implementation suggests per-directory since metadata files are in a shared directory.

## Out of Scope
- Performance optimization beyond the specific issues listed
- New features not related to the 17 identified issues
- Configuration file / CLI flags (R12 adds constructor flexibility, not full config system)
- Distributed consensus / leader election
- Authentication / authorization (not in current architecture)
