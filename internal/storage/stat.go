package storage

func (s *Storage) Stat(filename string) (*Metadata, error) {
	return s.LoadMetadata(filename)
}
