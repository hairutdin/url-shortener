package storage

type BatchURLRequest struct {
	UUID        string
	ShortURL    string
	OriginalURL string
}

type BatchURLOutput struct {
	CorrelationID string
	ShortURL      string
}
