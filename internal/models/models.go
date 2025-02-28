package models

// easyjson:json
type ShortenRequest struct {
	URL string `json:"url"`
}

// easyjson:json
type BatchShortenRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// easyjson:json
type ShortenResponse struct {
	Result string `json:"result"`
}

// easyjson:json
type BatchShortenResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
