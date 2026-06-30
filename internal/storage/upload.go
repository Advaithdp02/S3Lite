package storage

import (
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

func (s *Storage) Upload(path string) error {

	if err := s.InitializeNodes(); err != nil {
		return err
	}

	sourceFile, err := os.Open(path)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	info, err := sourceFile.Stat()
	if err != nil {
		return err
	}

	filename := filepath.Base(path)

	metadata := &Metadata{
		Name:      filename,
		Size:      info.Size(),
		ChunkSize: s.ChunkSize,
		Chunks:    []ChunkMetadata{},
	}

	buffer := make([]byte, s.ChunkSize)

	index := 0
	nodeIndex := 0

	for {

		bytesRead, err := sourceFile.Read(buffer)

		if err != nil && err != io.EOF {
			return err
		}

		if bytesRead == 0 {
			break
		}

		node := s.Nodes[nodeIndex]

		chunkID := uuid.NewString() + ".chunk"

		chunkPath := filepath.Join(
			node.Path,
			chunkID,
		)

		chunkFile, err := os.Create(chunkPath)
		if err != nil {
			return err
		}

		_, err = chunkFile.Write(buffer[:bytesRead])
		chunkFile.Close()

		if err != nil {
			return err
		}

		checksum := CalculateChecksum(buffer[:bytesRead])

		metadata.Chunks = append(metadata.Chunks, ChunkMetadata{
			ID:       chunkID,
			Index:    index,
			Size:     int64(bytesRead),
			Checksum: checksum,
			Node:     node.Name,
		})

		index++

		nodeIndex = (nodeIndex + 1) % len(s.Nodes)
	}

	return s.SaveMetadata(metadata)
}
