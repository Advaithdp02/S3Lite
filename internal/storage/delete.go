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

		chunkPath := filepath.Join(
			s.Root,
			chunk.Node,
			"chunks",
			chunk.ID,
		)

		if err := os.Remove(chunkPath); err != nil && !os.IsNotExist(err) {
			return err
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
