package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"github.com/Advaithdp02/s3lite/internal/storage"
	"time"
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

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	<-sig

	log.Println("Shutting down Metadata Server...")
}
