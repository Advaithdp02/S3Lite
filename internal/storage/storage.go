//Package storage which contains structure of storage and new function to intialize 
package storage

import (
	"os"
	"fmt"
	"encoding/json"
	"io"

	"github.com/google/uuid"
	"path/filepath"
)

type Storage struct{
	Root string
	ChunkSize int
}
//New is a constructer
func New(root string,chunksize int) *Storage{
	return &Storage{
		Root: root,
		ChunkSize: chunksize,
	}
}
//Upload is used to upload a single file to destination folder

func (s *Storage) Upload(path string) error{
	chunksDir := filepath.Join(s.Root, "chunks")
	metadataDir := filepath.Join(s.Root, "metadata")
	
	if err := os.MkdirAll(chunksDir, os.ModePerm); err != nil {
		return err
	}

	if err := os.MkdirAll(metadataDir, os.ModePerm); err != nil {
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

	metadata := Metadata{
		Name:      filename,
		Size:      info.Size(),
		ChunkSize: s.ChunkSize,
		Chunks:    []string{},
	}

		buffer := make([]byte, s.ChunkSize)

	for {

		bytesRead, err := sourceFile.Read(buffer)

		if err != nil && err != io.EOF {
			return err
		}

		if bytesRead == 0 {
			break
		}

		chunkID := uuid.New().String() + ".chunk"

		chunkPath := filepath.Join(chunksDir, chunkID)

		chunkFile, err := os.Create(chunkPath)
		if err != nil {
			return err
		}

		_, err = chunkFile.Write(buffer[:bytesRead])
		chunkFile.Close()

		if err != nil {
			return err
		}

		metadata.Chunks = append(metadata.Chunks, chunkID)
	}

	metadataPath := filepath.Join(metadataDir, filename+".json")

	metaFile, err := os.Create(metadataPath)
	if err != nil {
		return err
	}
	defer metaFile.Close()

	encoder := json.NewEncoder(metaFile)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(metadata); err != nil {
		return err
	}

	fmt.Println("Uploaded:", filename)
	fmt.Printf("Chunks created: %d\n", len(metadata.Chunks))

	return nil
} 
func (s *Storage) LoadMetadata(filename string) (*Metadata, error) {

	path := filepath.Join(s.Root, "metadata", filename+".json")

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

func (s *Storage) Download(filename string, destination string) error {

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

	for _, chunkID := range metadata.Chunks {

		chunkPath := filepath.Join(s.Root, "chunks", chunkID)

		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			return err
		}

		_, err = io.Copy(outputFile, chunkFile)
		chunkFile.Close()

		if err != nil {
			return err
		}
	}

	fmt.Println("Downloaded:", filename)

	return nil
}
