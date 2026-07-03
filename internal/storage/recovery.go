package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (s *Storage) StartRecovery(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		fmt.Println("Starting Recovery .... ")
		defer ticker.Stop()

		for {

			if err := s.RecoverCluster(); err != nil {
				fmt.Println("Recovery:", err)
			}

			<-ticker.C
		}
	}()
}

func (s *Storage) RecoverCluster() error {
	metadataDir := filepath.Join(s.Root, "metadata")

	files, err := os.ReadDir(metadataDir)
	if err != nil {
		return err
	}

	for _, file := range files {

		if file.IsDir() {
			continue
		}

		name := strings.TrimSuffix(file.Name(), ".json")

		metadata, err := s.LoadMetadata(name)
		if err != nil {
			continue
		}

		changed := false

		for i := range metadata.Chunks {

			ok, err := s.RecoverChunk(&metadata.Chunks[i])
			if err != nil {
				fmt.Println(err)
				continue
			}

			if ok {
				changed = true
			}
		}

		if changed {
			if err := s.SaveMetadata(metadata); err != nil {
				return err
			}
		}

	}

	return nil
}

func (s *Storage) RecoverChunk(chunk *ChunkMetadata) (bool, error) {
	healthy := []string{}

	var sourceData []byte

	for _, replica := range chunk.Replicas {

		node := s.GetNode(replica)

		if node == nil || !node.IsAlive() {
			continue
		}

		path := filepath.Join(
			node.Path,
			chunk.ID,
		)

		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		if CalculateChecksum(data) != chunk.Checksum {
			continue
		}

		healthy = append(healthy, replica)

		if sourceData == nil {
			sourceData = data
		}
	}

	if len(healthy) >= s.ReplicationFactor {
		return false, nil
	}

	for i := range s.Nodes {

		node := &s.Nodes[i]

		if !node.IsAlive() {
			continue
		}

		if contains(healthy, node.Name) {
			continue
		}

		chunkPath := filepath.Join(
			node.Path,
			chunk.ID,
		)

		if err := os.WriteFile(chunkPath, sourceData, 0o644); err != nil {
			return false, err
		}

		chunk.Replicas = append(
			chunk.Replicas,
			node.Name,
		)

		fmt.Printf(
			"Recovered chunk %s -> %s\n",
			chunk.ID,
			node.Name,
		)

		healthy = append(healthy, node.Name)

		if len(healthy) >= s.ReplicationFactor {
			break
		}
	}
	return true, nil
}
