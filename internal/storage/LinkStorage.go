package storage

import "context"

func NewLinkStorage() *LinkStorage {

	return &LinkStorage{links: map[string]string{}, users: map[int]bool{}, userLinks: map[int][]string{}}
}

func (s *LinkStorage) GetFromOriginal(ctx context.Context, originalURL string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	return originalURL, nil
}

func (s *LinkStorage) Save(ctx context.Context, correlationID string, short string, original string, userId int) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	s.links[short] = original
	s.userLinks[userId] = append(s.userLinks[userId], short)
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

func (s *LinkStorage) SaveUser(ctx context.Context, userId int) error {
	select {
	case <-ctx.Done():
		return nil
	default:
	}
	s.users[userId] = true
	s.userLinks[userId] = []string{}
	return nil
}

func (s *LinkStorage) GetUserFromId(ctx context.Context, userId int) (bool, error) {
	select {
	case <-ctx.Done():
		return false, nil
	default:
	}
	if _, exist := s.users[userId]; !exist {
		return false, ErrUserNotFound
	}
	return true, nil
}

func (s *LinkStorage) GetNewUser(ctx context.Context) (int, error) {
	select {
	case <-ctx.Done():
		return 0, nil
	default:
	}
	NewIndexUser := len(s.users) + 1
	return NewIndexUser, nil
}

func (s *LinkStorage) GetLinksByUserId(ctx context.Context, userId int) (map[string]string, error) {
	select {
	case <-ctx.Done():
		return nil, nil
	default:
	}
	result := make(map[string]string)
	if _, exist := s.userLinks[userId]; !exist {
		return nil, ErrUserNotFound
	}
	for _, linkShort := range s.userLinks[userId] {
		if _, exist := s.links[linkShort]; !exist {
			continue
		}
		original := s.links[linkShort]
		result[linkShort] = original
	}
	return result, nil
}
