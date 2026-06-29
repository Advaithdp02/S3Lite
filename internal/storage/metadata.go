package storage

type Metadata struct{
	Name      string   `json:"name"`
	Size      int64    `json:"size"`
	ChunkSize int      `json:"chunkSize"`
	Chunks    []string `json:"chunks"`
}
