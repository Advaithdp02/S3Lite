package storage

import (
	"os"
	"path/filepath"
)

func (s *Storage) ReplicateChunk(chunkID string, data []byte, startNode int) ([]string, error) {

	replicas := make([]string, 0, s.ReplicationFactor)

	for i := 0; i < s.ReplicationFactor; i++ {

		node := s.Nodes[(startNode+i)%len(s.Nodes)]

		chunkPath := filepath.Join(
			node.Path,
			chunkID,
		)

		if err := os.WriteFile(chunkPath, data, 0644); err != nil {
			return nil, err
		}

		replicas = append(replicas, node.Name)
	}

	return replicas, nil
}
