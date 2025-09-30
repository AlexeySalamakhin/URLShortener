package models

// URLRecord представляет запись URL в хранилище.
// Используется только во внутренней логике, не экспортируется в gRPC.
type URLRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id"`
	DeletedFlag bool   `json:"is_deleted"`
}

// ShortenRequest — запрос на сокращение URL (gRPC: CreateShortURLRequest)
type ShortenRequest struct {
	URL string `json:"url"`
}

// ShortenResponse — ответ с результатом сокращения URL (gRPC: CreateShortURLResponse)
type ShortenResponse struct {
	Result string `json:"result"`
}

// ShortURLBatchRequest — пакет запросов на сокращение URL (gRPC: BatchShortenRequest)
type ShortURLBatchRequest []URLBatchRequest

// URLBatchRequest — элемент пакетного запроса (gRPC: BatchShortenItem)
type URLBatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// URLBatchResponse — элемент пакетного ответа (gRPC: BatchShortenResponseItem)
type URLBatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// UserURLsResponse — DTO для вывода ссылок пользователя (gRPC: UserURL)
type UserURLsResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	DeletedFlag bool   `json:"is_deleted"`
}
