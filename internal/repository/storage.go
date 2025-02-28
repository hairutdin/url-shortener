package repository

type Storage interface {
	CreateShortURL(uuid, shortURL, originalURL string) (string, error)
	GetOriginalURL(shortURL string) (string, error)
	CreateBatchURLs(urls []BatchURLRequest) ([]BatchURLOutput, error)
	Ping() error
	Close() error
}
