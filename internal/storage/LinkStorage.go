package storage

func (s *LinkStorage) GetFromOriginal(originalURL string) (string, error) {
	return originalURL, nil
}

func NewLinkStorage() *LinkStorage {

	return &LinkStorage{links: map[string]string{}}
}

func (s *LinkStorage) Save(correlationID string, short string, original string) (string, error) {
	s.links[short] = original
	return short, nil
}

func (s *LinkStorage) Get(short string) (string, bool, error) {
	original, exists := s.links[short]
	return original, exists, nil
}

func (s *LinkStorage) Len() int {
	return len(s.links)
}

func (s *LinkStorage) Ping() error {
	return nil
}
