package main

import (
	"fmt"
	"os"

	"github.com/Advaithdp02/s3lite/internal/storage"
)

const (
	StorageRoot = "storage"
	ChunkSize   = 1024 * 1024 // 1 MB
)

func printUsage() {
	fmt.Println("S3Lite - Local Object Storage")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  upload <file>")
	fmt.Println("  download <file> <destination>")
	fmt.Println("  list")
	fmt.Println("  stat <file>")
	fmt.Println("  delete <file>")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	store := storage.New(StorageRoot, ChunkSize, 2)

	switch os.Args[1] {

	case "upload":
		if len(os.Args) != 3 {
			fmt.Println("Usage: upload <file>")
			return
		}

		if err := store.Upload(os.Args[2]); err != nil {
			fmt.Println("Upload failed:", err)
			return
		}

		fmt.Println("Upload completed successfully.")

	case "download":
		if len(os.Args) != 4 {
			fmt.Println("Usage: download <file> <destination>")
			return
		}

		if err := store.Download(os.Args[2], os.Args[3]); err != nil {
			fmt.Println("Download failed:", err)
			return
		}

		fmt.Println("Download completed successfully.")

	case "list":
		objects, err := store.List()
		if err != nil {
			fmt.Println("List failed:", err)
			return
		}

		fmt.Printf("%-20s %-12s %-10s\n", "NAME", "SIZE", "CHUNKS")

		for _, obj := range objects {
			fmt.Printf(
				"%-20s %-12d %-10d\n",
				obj.Name,
				obj.Size,
				obj.Chunks,
			)
		}
	case "stat":
		if len(os.Args) != 3 {
			fmt.Println("Usage: stat <file>")
			return
		}
		meta, err := store.Stat(os.Args[2])
		if err != nil {
			fmt.Println("Stat failed:", err)
			return
		}

		fmt.Println("========================================")
		fmt.Println("Object Information")
		fmt.Println("========================================")
		fmt.Println("Name       :", meta.Name)
		fmt.Println("Size       :", meta.Size, "bytes")
		fmt.Println("Chunk Size :", meta.ChunkSize, "bytes")
		fmt.Println("Chunks     :", len(meta.Chunks))
		fmt.Println()

		fmt.Println("========================================")
		fmt.Println("Chunk Details")
		fmt.Println("========================================")

		for _, chunk := range meta.Chunks {
			fmt.Printf("Chunk %d\n", chunk.Index)
			fmt.Println("----------------------------------------")
			fmt.Println("Chunk ID  :", chunk.ID)
			fmt.Println("Size      :", chunk.Size, "bytes")
			fmt.Println("Checksum  :", chunk.Checksum)
			fmt.Println("Replicas  :", chunk.Replicas)
			fmt.Println()
		}
	case "delete":
		if len(os.Args) != 3 {
			fmt.Println("Usage: delete <file>")
			return
		}

		if err := store.Delete(os.Args[2]); err != nil {
			fmt.Println("Delete failed:", err)
			return
		}

		fmt.Println("Object deleted successfully.")

	default:
		fmt.Printf("Unknown command: %s\n\n", os.Args[1])
		printUsage()
	}
}
