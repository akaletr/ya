package model

type ShortenerRequest struct {
	URL string `json:"url,omitempty"`
}

type ShortenerResponse struct {
	Result string `json:"result,omitempty"`
}

type Item struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type BatchRequestItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchResponseItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type DataBatchItem struct {
	ID          string
	Short       string
	Long        string
	Correlation string
}
