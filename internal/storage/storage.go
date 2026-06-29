//Package storage which contains structure of storage and new function to intialize 
package storage


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
