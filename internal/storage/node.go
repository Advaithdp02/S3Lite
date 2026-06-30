package storage

import (
	"os"
	"path/filepath")

type Node struct {
	Name string
	Path string
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

	for _, node := range s.Nodes {

		if err := os.MkdirAll(node.Path, os.ModePerm); err != nil {
			return err
		}
	}

	return os.MkdirAll(filepath.Join(s.Root, "metadata"), os.ModePerm)
}
