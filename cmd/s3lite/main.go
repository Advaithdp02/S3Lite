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

	store := storage.New(StorageRoot, ChunkSize)

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
		if err := store.List(); err != nil {
			fmt.Println("List failed:", err)
		}

	case "stat":
		if len(os.Args) != 3 {
			fmt.Println("Usage: stat <file>")
			return
		}

		if err := store.Stat(os.Args[2]); err != nil {
			fmt.Println("Stat failed:", err)
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
