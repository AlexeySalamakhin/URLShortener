package models

// URLRecord представляет запись URL в хранилище.
type URLRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id"`
	DeletedFlag bool   `json:"is_deleted"`
}

// ShortenRequest — запрос на сокращение URL.
type ShortenRequest struct {
	URL string `json:"url"`
}

// ShortenResponse — ответ с результатом сокращения URL.
type ShortenResponse struct {
	Result string `json:"result"`
}

// ShortURLBatchRequest — пакет запросов на сокращение URL.
type ShortURLBatchRequest []URLBatchRequest

// URLBatchRequest — элемент пакетного запроса.
type URLBatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// URLBatchResponse — элемент пакетного ответа.
type URLBatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// UserURLsResponse — DTO для вывода ссылок пользователя.
type UserURLsResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	DeletedFlag bool   `json:"is_deleted"`
}
