package storage

import (
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Node struct {
	Name string
	Path string
	Alive bool
	LastHeartbeat time.Time
	mu sync.RWMutex
}

func DefaultNodes(root string) []Node {
	return []Node{
		{
			Name: "node1",
			Path: filepath.Join(root, "node1", "chunks"),
		},
		{
			Name: "node2",
			Path: filepath.Join(root, "node2", "chunks"),
		},
		{
			Name: "node3",
			Path: filepath.Join(root, "node3", "chunks"),
		},
	}
}
func (s *Storage) InitializeNodes() error {
	for i := range s.Nodes {
		node := &s.Nodes[i]

		if err := os.MkdirAll(node.Path, os.ModePerm); err != nil {
			return err
		}
	}

	return os.MkdirAll(filepath.Join(s.Root, "metadata"), os.ModePerm)
}
