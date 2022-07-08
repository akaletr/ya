package model

type ShortenerRequest struct {
	URL string `json:"url,omitempty"`
}

type ShortenerResponse struct {
	Result string `json:"result,omitempty"`
}

type AllShortenerRequest []Item

type Item struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type BatchRequest []BatchRequestItem

type BatchRequestItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchResponse []BatchResponseItem

type BatchResponseItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type DataBatch []DataBatchItem

type DataBatchItem struct {
	ID          string
	Short       string
	Long        string
	Correlation string
}
