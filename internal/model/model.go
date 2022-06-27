package model

type ShortenerRequest struct {
	URL string `json:"url,omitempty"`
}

type ShortenerResponse struct {
	Result string `json:"result,omitempty"`
}
