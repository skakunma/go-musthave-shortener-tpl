package storage

import "context"

func NewLinkStorage() *LinkStorage {

	return &LinkStorage{links: map[string]string{}}
}

func (s *LinkStorage) GetFromOriginal(ctx context.Context, originalURL string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	return originalURL, nil
}

func (s *LinkStorage) Save(ctx context.Context, correlationID string, short string, original string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	s.links[short] = original
	return short, nil
}

func (s *LinkStorage) Get(ctx context.Context, short string) (string, bool, error) {
	select {
	case <-ctx.Done():
		return "", false, ctx.Err()
	default:
	}
	original, exists := s.links[short]
	return original, exists, nil
}

func (s *LinkStorage) Len(ctx context.Context) int {
	select {
	case <-ctx.Done():
		return 0
	default:
	}
	return len(s.links)
}

func (s *LinkStorage) Ping(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return nil
	default:
	}
	return nil
}
