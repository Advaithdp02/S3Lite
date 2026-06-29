package main

import (
	"fmt"
	"os"

	"github.com/Advaithdp02/s3lite/internal/storage"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println(" upload <file>")
		fmt.Println(" download <file> <destination>")
		return
	}

	store := storage.New("storage", 1024)

	switch os.Args[1] {

	case "upload":

		if len(os.Args) != 3 {
			fmt.Println("Usage: upload <file>")
			return
		}

		if err := store.Upload(os.Args[2]); err != nil {
			fmt.Println(err)
		}

	case "download":

		if len(os.Args) != 4 {
			fmt.Println("Usage: download <file> <destination>")
			return
		}

		if err := store.Download(os.Args[2], os.Args[3]); err != nil {
			fmt.Println(err)
		}

	default:
		fmt.Println("Unknown command")
	}
}
