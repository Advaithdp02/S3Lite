package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"github.com/Advaithdp02/s3lite/internal/storage"
	"github.com/Advaithdp02/s3lite/internal/metadata"
	"time"
	"net/http"
)
const (
	StorageRoot = "storage"
	ChunkSize   = 1024 * 1024
)
func main() {
	s := storage.New(StorageRoot, ChunkSize)

	s.StartHeartBeat(2 * time.Second)
	s.StartRecovery(5 * time.Second)

	log.Println("Metadata Server started")
	http.HandleFunc("/health", metadata.HealthHandler)
	http.HandleFunc("/upload", metadata.UploadHandler(s))
	http.HandleFunc("/download", metadata.DownloadHandler(s))
	http.HandleFunc("/list",     metadata.ListHandler(s))
	http.HandleFunc("/stat",     metadata.StatHandler(s))
	http.HandleFunc("/delete",   metadata.DeleteHandler(s))
	log.Fatal(http.ListenAndServe(":8080", nil))
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	<-sig

	log.Println("Shutting down Metadata Server...")
}
