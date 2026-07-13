package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

func (s *Storage) Delete(filename string) error {
	metadata, err := s.LoadMetadata(filename)
	if err != nil {
		return err
	}

	var errs []error
	for _, chunk := range metadata.Chunks {
		for _, replica := range chunk.Replicas {
			chunkPath := filepath.Join(s.Root, replica, "chunks", chunk.ID)
			if err := os.Remove(chunkPath); err != nil && !os.IsNotExist(err) {
				errs = append(errs, fmt.Errorf("remove %s: %w", chunkPath, err))
			}
		}
	}
	metadataPath := filepath.Join(s.Root, "metadata", filename+".json")
	if err := os.Remove(metadataPath); err != nil && !os.IsNotExist(err) {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return fmt.Errorf("partial deletion: %v", errs)
	}
	return nil
}
