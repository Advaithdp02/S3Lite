package storage

func contains(slice []string, value string) bool {

	for _, v := range slice {

		if v == value {
			return true
		}
	}

	return false
}

func (s *Storage) GetNode(name string) *Node {

	for i := range s.Nodes {

		if s.Nodes[i].Name == name {
			return &s.Nodes[i]
		}
	}

	return nil
}
