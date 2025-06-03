package models

type URLRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}

type ShortURLBatchRequest []URLBatchRequest

type URLBatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type ShortURLBatchResponse []URLBatchResponse

type URLBatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// [
//     {
//         "correlation_id": "<строковый идентификатор из объекта запроса>",
//         "short_url": "<результирующий сокращённый URL>"
//     },
//     ...
// ]
// [
//     {
//         "correlation_id": "<строковый идентификатор>",
//         "original_url": "<URL для сокращения>"
//     },
//     ...
// ]
