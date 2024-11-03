package storage

type Storage interface {
	CreateShortURL(uuid, shortURL, originalURL string) error
	GetOriginalURL(shortURL string) (string, error)
	CreateBatchURLs(urls []BatchURLRequest) ([]BatchURLOutput, error)
	Ping() error
	Close() error
}
