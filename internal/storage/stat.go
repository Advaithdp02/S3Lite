package storage

import "fmt"

func (s *Storage) Stat(filename string) error {

	meta, err := s.LoadMetadata(filename)
	if err != nil {
		return err
	}

	fmt.Println("Name      :", meta.Name)
	fmt.Println("Size      :", meta.Size)
	fmt.Println("Chunk Size:", meta.ChunkSize)
	fmt.Println("Chunks    :", len(meta.Chunks))

	fmt.Println()

	for _, chunk := range meta.Chunks {

		fmt.Printf("Chunk %d\n", chunk.Index)
		fmt.Println("  ID       :", chunk.ID)
		fmt.Println("  Size     :", chunk.Size)
		fmt.Println("  Checksum :", chunk.Checksum)
		fmt.Println()
	}
	return nil
}
