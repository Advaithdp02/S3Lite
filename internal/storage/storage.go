// Package storage which contains structure of storage and new function to intialize
package storage

import (
	"sync"
	"time"
)


type Storage struct{
	Root string
	ChunkSize int
	ReplicationFactor int
	Nodes []Node
	mu sync.Mutex
}
//New is a constructer
func New(root string, chunksize int, replicationFactor int) *Storage{
	s:=&Storage{
		Root: root,
		ChunkSize: chunksize,
		ReplicationFactor: replicationFactor,
		Nodes: DefaultNodes(root),
	}
	if err := s.InitializeNodes(); err != nil {
		panic(err)
	}
	s.StartHeartBeat(2*time.Second)
	return s
}
