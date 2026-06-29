package storage


type ChunkMetadata struct {
	ID    string `json:"id"`
	Index int    `json:"index"`
	Size  int64  `json:"size"`
	Checksum string `json:"checksum"`
}

type Metadata struct {
	Name      string          `json:"name"`
	Size      int64           `json:"size"`
	ChunkSize int             `json:"chunk_size"`
	Chunks    []ChunkMetadata `json:"chunks"`
}
