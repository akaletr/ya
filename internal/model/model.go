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
