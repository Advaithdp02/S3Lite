package storage
import (
	"bytes"
	"fmt"
	"io"
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

	chunkPath := filepath.Join(
		s.Root,
		"chunks",
		chunk.ID,
	)

	data, err := os.ReadFile(chunkPath)
	if err != nil {
		return err
	}

	checksum := CalculateChecksum(data)

	if checksum != chunk.Checksum {
		return fmt.Errorf(
			"checksum mismatch for chunk %s",
			chunk.ID,
		)
	}

	_, err = io.Copy(outputFile, bytes.NewReader(data))
	if err != nil {
		return err
	}
}

	return nil
}
