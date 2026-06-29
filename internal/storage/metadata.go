package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func (s *Storage) SaveMetadata(metadata *Metadata) error {

	metadataDir := filepath.Join(s.Root, "metadata")

	if err := os.MkdirAll(metadataDir, os.ModePerm); err != nil {
		return err
	}

	path := filepath.Join(metadataDir, metadata.Name+".json")

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	return encoder.Encode(metadata)
}

func (s *Storage) LoadMetadata(filename string) (*Metadata, error) {

	path := filepath.Join(
		s.Root,
		"metadata",
		filename+".json",
	)

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var metadata Metadata

	if err := json.NewDecoder(file).Decode(&metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}
