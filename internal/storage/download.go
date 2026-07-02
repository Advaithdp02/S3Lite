package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

func (s *Storage) Download(filename, destination string) error {

	metadata, err := s.LoadMetadata(filename)
	if err != nil {
		return err
	}

	outputPath := filepath.Join(destination, filename)

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	for _, chunk := range metadata.Chunks {

		var data []byte
		var found bool

		for _, replica := range chunk.Replicas {

			chunkPath := filepath.Join(
				s.Root,
				replica,
				"chunks",
				chunk.ID,
			)

			bytes, err := os.ReadFile(chunkPath)
			if err != nil {
				continue
			}

			if CalculateChecksum(bytes) != chunk.Checksum {
				continue
			}

			data = bytes
			found = true
			break
		}

		if !found {
			return fmt.Errorf(
				"no healthy replica found for chunk %s",
				chunk.ID,
			)
		}

		_, err := outputFile.Write(data)
		if err != nil {
			return err
		}
	}

	return nil
}
