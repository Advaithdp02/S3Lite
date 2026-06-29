package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (s *Storage) List() error {

	metadataDir := filepath.Join(s.Root, "metadata")

	files, err := os.ReadDir(metadataDir)
	if err != nil {
		return err
	}

	fmt.Printf("%-20s %-12s %-10s\n", "NAME", "SIZE", "CHUNKS")

	for _, file := range files {

		if file.IsDir() {
			continue
		}

		name := strings.TrimSuffix(file.Name(), ".json")

		meta, err := s.LoadMetadata(name)
		if err != nil {
			continue
		}

		fmt.Printf(
			"%-20s %-12d %-10d\n",
			meta.Name,
			meta.Size,
			len(meta.Chunks),
		)
	}

	return nil
}
