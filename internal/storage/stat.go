package storage

import "fmt"

func (s *Storage) Stat(filename string) error {

	meta, err := s.LoadMetadata(filename)
	if err != nil {
		return err
	}

	fmt.Println("========================================")
	fmt.Println("Object Information")
	fmt.Println("========================================")
	fmt.Println("Name       :", meta.Name)
	fmt.Println("Size       :", meta.Size, "bytes")
	fmt.Println("Chunk Size :", meta.ChunkSize, "bytes")
	fmt.Println("Chunks     :", len(meta.Chunks))
	fmt.Println()

	fmt.Println("========================================")
	fmt.Println("Chunk Details")
	fmt.Println("========================================")

	for _, chunk := range meta.Chunks {

		fmt.Printf("Chunk %d\n", chunk.Index)
		fmt.Println("----------------------------------------")
		fmt.Println("Chunk ID  :", chunk.ID)
		fmt.Println("Size      :", chunk.Size, "bytes")
		fmt.Println("Checksum  :", chunk.Checksum)
		fmt.Println("Replicas  :",chunk.Replicas)
		fmt.Println()
	}

	return nil
}
