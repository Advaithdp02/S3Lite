package storage

import (
	"os"
	"path/filepath"
)

func (s *Storage) Delete(filename string) error {

	meta, err := s.LoadMetadata(filename)
	if err != nil {
		return err
	}

	for _, chunk := range meta.Chunks {

		chunkPath := filepath.Join(
			s.Root,
			"chunks",
			chunk.ID,
		)

		if err := os.Remove(chunkPath); err != nil {
			return err
		}
	}

	metadataPath := filepath.Join(
		s.Root,
		"metadata",
		filename+".json",
	)

	return os.Remove(metadataPath)
}
