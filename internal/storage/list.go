package storage

import (
	"os"
	"path/filepath"
	"strings"
)

func (s *Storage) List() ([]ObjectInfo, error) {

	metadataDir := filepath.Join(s.Root, "metadata")

	files, err := os.ReadDir(metadataDir)
	if err != nil {
		return nil, err
	}

	var objects []ObjectInfo

	for _, file := range files {

		if file.IsDir() {
			continue
		}

		name := strings.TrimSuffix(file.Name(), ".json")

		meta, err := s.LoadMetadata(name)
		if err != nil {
			continue
		}

		objects = append(objects, ObjectInfo{
			Name:   meta.Name,
			Size:   meta.Size,
			Chunks: len(meta.Chunks),
		})
	}

	return objects, nil
}
