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

func (s *LinkStorage) Save(ctx context.Context, correlationID string, short string, original string, userID int) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	s.links[short] = original
	s.userLinks[userID] = append(s.userLinks[userID], short)
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

func (s *LinkStorage) SaveUser(ctx context.Context, userID int) error {
	select {
	case <-ctx.Done():
		return nil
	default:
	}
	s.users[userID] = true
	s.userLinks[userID] = []string{}
	return nil
}

func (s *LinkStorage) GetUserFromID(ctx context.Context, userID int) (bool, error) {
	select {
	case <-ctx.Done():
		return false, nil
	default:
	}
	if _, exist := s.users[userID]; !exist {
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

func (s *LinkStorage) GetLinksByUserID(ctx context.Context, userID int) (map[string]string, error) {
	select {
	case <-ctx.Done():
		return nil, nil
	default:
	}
	result := make(map[string]string)
	if _, exist := s.userLinks[userID]; !exist {
		return nil, ErrUserNotFound
	}
	for _, linkShort := range s.userLinks[userID] {
		if _, exist := s.links[linkShort]; !exist {
			continue
		}
		original := s.links[linkShort]
		result[linkShort] = original
	}
	return result, nil
}

func (s *LinkStorage) AddLinksBatch(ctx context.Context, links []InfoAboutURL, userID int) ([]string, error) {
	shortLinks := []string{}
	for _, link := range links {
		shortLink := link.ShortLink
		s.links[shortLink] = link.OriginalURL
		shortLinks = append(shortLinks, shortLink)
	}
	return shortLinks, nil
}

func (s *LinkStorage) DeleteURL(ctx context.Context, UUID string) error {
	s.deletedLinks[UUID] = true
	return nil
}

func (s *LinkStorage) GetUserFromUUID(ctx context.Context, UUID string) (int, error) {
	id, exist := s.uuidUser[UUID]
	if !exist {
		return 0, nil
	}
	return id, nil
}
