// Package storage which contains structure of storage and new function to intialize
package storage

import "time"


type Storage struct{
	Root string
	ChunkSize int
	ReplicationFactor int
	Nodes []Node
}
//New is a constructer
func New(root string,chunksize int) *Storage{
	s:=&Storage{
		Root: root,
		ChunkSize: chunksize,
		ReplicationFactor: 2,
		Nodes: DefaultNodes(root),
	}
	s.StartHeartBeat(2*time.Second)
	return s

}
