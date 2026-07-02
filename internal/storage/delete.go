package storage

import (
	"os"
	"path/filepath"
)

func (s *Storage) Delete(filename string) error {
	metadata, err := s.LoadMetadata(filename)
	if err != nil {
		return err
	}

	for _, chunk := range metadata.Chunks {
		for _, replica := range chunk.Replicas {

			chunkPath := filepath.Join(
				s.Root,
				replica,
				"chunks",
				chunk.ID,
			)

			_ = os.Remove(chunkPath)
		}
	}
	metadataPath := filepath.Join(
		s.Root,
		"metadata",
		filename+".json",
	)

	if err := os.Remove(metadataPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}
