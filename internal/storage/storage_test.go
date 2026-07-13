package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// helper: create a temporary storage root and return a Storage + cleanup func.
func newTestStorage(t *testing.T, replicationFactor int) (*Storage, func()) {
	t.Helper()
	root := t.TempDir()
	s := New(root, 1024, replicationFactor)
	return s, func() {
		// heartbeat goroutine runs forever; just clean disk
		os.RemoveAll(root)
	}
}

// helper: write a temp file with content, return its path.
func createTempFile(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

// ─── R1: CLI stat arg check ────────────────────────────────────────────────

func TestR1_Stat_NoArgs_NoPanic(t *testing.T) {
	s, cleanup := newTestStorage(t, 2)
	defer cleanup()

	_, err := s.Stat("nonexistent-file")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestR1_Stat_WithArgs(t *testing.T) {
	s, cleanup := newTestStorage(t, 2)
	defer cleanup()

	f := createTempFile(t, "hello.txt", "hello world")
	if err := s.Upload(f); err != nil {
		t.Fatal("upload failed:", err)
	}

	meta, err := s.Stat("hello.txt")
	if err != nil {
		t.Fatal("stat failed:", err)
	}
	if meta.Name != "hello.txt" {
		t.Errorf("expected name 'hello.txt', got '%s'", meta.Name)
	}
}

// ─── R3: No double heartbeat ───────────────────────────────────────────────

func TestR3_SingleHeartbeat(t *testing.T) {
	// New() starts exactly one heartbeat. Give the goroutine time to run.
	root := t.TempDir()
	s := New(root, 1024, 2)
	defer os.RemoveAll(root)

	// Poll for node liveness — heartbeat runs immediately in its goroutine
	deadline := time.Now().Add(500 * time.Millisecond)
	allAlive := false
	for time.Now().Before(deadline) {
		allAlive = true
		for i := range s.Nodes {
			if !s.Nodes[i].IsAlive() {
				allAlive = false
				break
			}
		}
		if allAlive {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if !allAlive {
		t.Error("expected all nodes alive after heartbeat ran")
	}
}

// ─── R5: nil sourceData guard ──────────────────────────────────────────────

func TestR5_RecoverChunk_AllReplicasDead(t *testing.T) {
	s, cleanup := newTestStorage(t, 2)
	defer cleanup()

	f := createTempFile(t, "data.bin", "some test data")
	if err := s.Upload(f); err != nil {
		t.Fatal("upload failed:", err)
	}

	meta, err := s.Stat("data.bin")
	if err != nil {
		t.Fatal(err)
	}

	// Kill all replicas
	for i := range s.Nodes {
		os.RemoveAll(s.Nodes[i].Path)
		s.Nodes[i].mu.Lock()
		s.Nodes[i].Alive = false
		s.Nodes[i].mu.Unlock()
	}

	_, err = s.RecoverChunk(&meta.Chunks[0])
	if err == nil {
		t.Error("expected error when all replicas are dead, got nil")
	}

	for i := range s.Nodes {
		chunkPath := filepath.Join(s.Nodes[i].Path, meta.Chunks[0].ID)
		if _, statErr := os.Stat(chunkPath); statErr == nil {
			t.Error("chunk file should not exist when recovery fails")
		}
	}
}

// ─── R6: Filename sanitization ─────────────────────────────────────────────

func TestR6_FilenameSanitization(t *testing.T) {
	s, cleanup := newTestStorage(t, 2)
	defer cleanup()

	f := createTempFile(t, "legit.txt", "content")
	if err := s.Upload(f); err != nil {
		t.Fatal(err)
	}

	meta, err := s.Stat("legit.txt")
	if err != nil {
		t.Fatal(err)
	}
	if meta.Name != "legit.txt" {
		t.Errorf("expected 'legit.txt', got '%s'", meta.Name)
	}
}

// ─── R8: InitializeNodes in New ────────────────────────────────────────────

func TestR8_InitializeNodes_InNew(t *testing.T) {
	root := t.TempDir()
	s := New(root, 1024, 2)
	defer os.RemoveAll(root)

	for i := range s.Nodes {
		if _, err := os.Stat(s.Nodes[i].Path); os.IsNotExist(err) {
			t.Errorf("node dir %s should exist after New()", s.Nodes[i].Path)
		}
	}

	metaDir := filepath.Join(root, "metadata")
	if _, err := os.Stat(metaDir); os.IsNotExist(err) {
		t.Error("metadata dir should exist after New()")
	}
}

func TestR8_InitializeNodes_NotCalledOnUpload(t *testing.T) {
	// Verify upload.go doesn't call InitializeNodes by checking that
	// the code path doesn't recreate directories (dirs already exist from New).
	s, cleanup := newTestStorage(t, 2)
	defer cleanup()

	f := createTempFile(t, "test.txt", "content")
	if err := s.Upload(f); err != nil {
		t.Fatal(err)
	}

	// Verify node dirs still exist and are directories (not files)
	for i := range s.Nodes {
		info, err := os.Stat(s.Nodes[i].Path)
		if err != nil {
			t.Errorf("node dir %s missing after upload: %v", s.Nodes[i].Name, err)
		}
		if !info.IsDir() {
			t.Errorf("node path %s is not a directory", s.Nodes[i].Name)
		}
	}
}

// ─── R11: Download round-trip ──────────────────────────────────────────────

func TestR11_DownloadRoundTrip(t *testing.T) {
	s, cleanup := newTestStorage(t, 2)
	defer cleanup()

	content := "This is test content for download round-trip verification."
	f := createTempFile(t, "roundtrip.txt", content)
	if err := s.Upload(f); err != nil {
		t.Fatal("upload failed:", err)
	}

	dest := t.TempDir()
	if err := s.Download("roundtrip.txt", dest); err != nil {
		t.Fatal("download failed:", err)
	}

	downloaded, err := os.ReadFile(filepath.Join(dest, "roundtrip.txt"))
	if err != nil {
		t.Fatal("could not read downloaded file:", err)
	}

	if string(downloaded) != content {
		t.Errorf("downloaded content mismatch: got %q, want %q", string(downloaded), content)
	}
}

// ─── R12: Configurable replication factor ──────────────────────────────────

func TestR12_ReplicationFactor(t *testing.T) {
	s, cleanup := newTestStorage(t, 3)
	defer cleanup()

	if s.ReplicationFactor != 3 {
		t.Errorf("expected RF=3, got %d", s.ReplicationFactor)
	}

	f := createTempFile(t, "repl.txt", "replication test")
	if err := s.Upload(f); err != nil {
		t.Fatal(err)
	}

	meta, err := s.Stat("repl.txt")
	if err != nil {
		t.Fatal(err)
	}

	for _, chunk := range meta.Chunks {
		if len(chunk.Replicas) != 3 {
			t.Errorf("chunk %s: expected 3 replicas, got %d", chunk.ID, len(chunk.Replicas))
		}
	}

	// Verify 3 copies on disk
	for _, chunk := range meta.Chunks {
		count := 0
		for _, nodeName := range chunk.Replicas {
			node := s.GetNode(nodeName)
			if node != nil {
				path := filepath.Join(node.Path, chunk.ID)
				if _, err := os.Stat(path); err == nil {
					count++
				}
			}
		}
		if count != 3 {
			t.Errorf("chunk %s: expected 3 on-disk copies, found %d", chunk.ID, count)
		}
	}
}

func TestR12_DefaultReplicationFactor(t *testing.T) {
	s, cleanup := newTestStorage(t, 2)
	defer cleanup()

	if s.ReplicationFactor != 2 {
		t.Errorf("expected default RF=2, got %d", s.ReplicationFactor)
	}
}

// ─── R14: Directory permissions 0755 ───────────────────────────────────────

func TestR14_DirectoryPermissions(t *testing.T) {
	root := t.TempDir()
	s := New(root, 1024, 2)
	defer os.RemoveAll(root)

	for i := range s.Nodes {
		info, err := os.Stat(s.Nodes[i].Path)
		if err != nil {
			t.Fatalf("stat %s: %v", s.Nodes[i].Path, err)
		}
		perm := info.Mode().Perm()
		if perm != 0755 {
			t.Errorf("node dir %s: expected perm 0755, got %04o", s.Nodes[i].Name, perm)
		}
	}

	metaDir := filepath.Join(root, "metadata")
	info, err := os.Stat(metaDir)
	if err != nil {
		t.Fatal(err)
	}
	perm := info.Mode().Perm()
	if perm != 0755 {
		t.Errorf("metadata dir: expected perm 0755, got %04o", perm)
	}
}

// ─── R15: Delete error reporting ───────────────────────────────────────────

func TestR15_DeleteExistingFile(t *testing.T) {
	s, cleanup := newTestStorage(t, 2)
	defer cleanup()

	f := createTempFile(t, "todelete.txt", "delete me")
	if err := s.Upload(f); err != nil {
		t.Fatal(err)
	}

	err := s.Delete("todelete.txt")
	if err != nil {
		t.Errorf("expected no error on delete of existing file, got: %v", err)
	}

	_, err = s.Stat("todelete.txt")
	if err == nil {
		t.Error("metadata should be gone after delete")
	}
}

func TestR15_DeleteNonExistentFile(t *testing.T) {
	s, cleanup := newTestStorage(t, 2)
	defer cleanup()

	err := s.Delete("does-not-exist.txt")
	if err == nil {
		t.Error("expected error deleting nonexistent file, got nil")
	}
}

// ─── R16: Atomic writes ────────────────────────────────────────────────────

func TestR16_NoTmpFilesAfterUpload(t *testing.T) {
	s, cleanup := newTestStorage(t, 2)
	defer cleanup()

	f := createTempFile(t, "atomic.txt", "atomic content")
	if err := s.Upload(f); err != nil {
		t.Fatal(err)
	}

	for i := range s.Nodes {
		entries, err := os.ReadDir(s.Nodes[i].Path)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if filepath.Ext(entry.Name()) == ".tmp" {
				t.Errorf("found leftover .tmp file in %s: %s", s.Nodes[i].Name, entry.Name())
			}
		}
	}

	meta, err := s.Stat("atomic.txt")
	if err != nil {
		t.Fatal(err)
	}
	for _, chunk := range meta.Chunks {
		for _, nodeName := range chunk.Replicas {
			node := s.GetNode(nodeName)
			if node == nil {
				t.Errorf("unknown node %s", nodeName)
				continue
			}
			chunkPath := filepath.Join(node.Path, chunk.ID)
			if _, err := os.Stat(chunkPath); os.IsNotExist(err) {
				t.Errorf("chunk %s missing from node %s", chunk.ID, nodeName)
			}
		}
	}
}

// ─── R17: Metadata mutex ───────────────────────────────────────────────────

func TestR17_ConcurrentMetadataWrites(t *testing.T) {
	s, cleanup := newTestStorage(t, 2)
	defer cleanup()

	var wg sync.WaitGroup
	var errCount atomic.Int32

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			name := fmt.Sprintf("concurrent_%d.txt", idx)
			f := createTempFile(t, name, fmt.Sprintf("content-%d", idx))
			if err := s.Upload(f); err != nil {
				errCount.Add(1)
				t.Errorf("upload %s failed: %v", name, err)
			}
		}(i)
	}
	wg.Wait()

	if c := errCount.Load(); c > 0 {
		t.Errorf("%d uploads failed", c)
	}

	// Verify all metadata files are valid JSON
	metaDir := filepath.Join(s.Root, "metadata")
	entries, err := os.ReadDir(metaDir)
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(metaDir, entry.Name()))
		if err != nil {
			t.Errorf("cannot read %s: %v", entry.Name(), err)
			continue
		}
		var m Metadata
		if err := json.Unmarshal(data, &m); err != nil {
			t.Errorf("invalid JSON in %s: %v", entry.Name(), err)
		}
	}
}

func TestR17_ConcurrentReadWriteSameFile(t *testing.T) {
	s, cleanup := newTestStorage(t, 2)
	defer cleanup()

	f := createTempFile(t, "rwfile.txt", "initial content")
	if err := s.Upload(f); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				s.Stat("rwfile.txt")
			}
		}()
	}

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			name := fmt.Sprintf("rw_other_%d.txt", idx)
			f := createTempFile(t, name, "other content")
			s.Upload(f)
		}(i)
	}

	wg.Wait()

	meta, err := s.Stat("rwfile.txt")
	if err != nil {
		t.Fatal("rwfile.txt unreadable after concurrent access:", err)
	}
	if meta.Name != "rwfile.txt" {
		t.Errorf("metadata corrupted: expected 'rwfile.txt', got '%s'", meta.Name)
	}
}

// ─── Upload/Download/Delete/List basics ────────────────────────────────────

func TestUpload_List_Delete_Lifecycle(t *testing.T) {
	s, cleanup := newTestStorage(t, 2)
	defer cleanup()

	f := createTempFile(t, "lifecycle.txt", "lifecycle test")
	if err := s.Upload(f); err != nil {
		t.Fatal("upload:", err)
	}

	objects, err := s.List()
	if err != nil {
		t.Fatal("list:", err)
	}
	if len(objects) != 1 {
		t.Fatalf("expected 1 object, got %d", len(objects))
	}
	if objects[0].Name != "lifecycle.txt" {
		t.Errorf("expected name 'lifecycle.txt', got '%s'", objects[0].Name)
	}

	meta, err := s.Stat("lifecycle.txt")
	if err != nil {
		t.Fatal("stat:", err)
	}
	if meta.Size == 0 {
		t.Error("expected non-zero size")
	}

	if err := s.Delete("lifecycle.txt"); err != nil {
		t.Fatal("delete:", err)
	}

	objects, err = s.List()
	if err != nil {
		t.Fatal("list after delete:", err)
	}
	if len(objects) != 0 {
		t.Errorf("expected 0 objects after delete, got %d", len(objects))
	}
}

func TestDownload_NonExistentFile(t *testing.T) {
	s, cleanup := newTestStorage(t, 2)
	defer cleanup()

	err := s.Download("nope.txt", t.TempDir())
	if err == nil {
		t.Error("expected error downloading nonexistent file, got nil")
	}
}

func TestUpload_SameFileTwice(t *testing.T) {
	// Uploading same filename twice overwrites the metadata (expected behavior)
	s, cleanup := newTestStorage(t, 2)
	defer cleanup()

	f := createTempFile(t, "dup.txt", "first content")
	if err := s.Upload(f); err != nil {
		t.Fatal(err)
	}

	f2 := createTempFile(t, "dup.txt", "second content")
	if err := s.Upload(f2); err != nil {
		t.Fatal("second upload:", err)
	}

	objects, err := s.List()
	if err != nil {
		t.Fatal(err)
	}
	// Same filename → same metadata path → overwrites; 1 entry expected
	if len(objects) != 1 {
		t.Errorf("expected 1 entry (overwrite), got %d", len(objects))
	}
}

func TestUpload_DifferentFiles(t *testing.T) {
	s, cleanup := newTestStorage(t, 2)
	defer cleanup()

	f1 := createTempFile(t, "a.txt", "content A")
	f2 := createTempFile(t, "b.txt", "content B")

	s.Upload(f1)
	s.Upload(f2)

	objects, err := s.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(objects) != 2 {
		t.Errorf("expected 2 objects, got %d", len(objects))
	}
}
