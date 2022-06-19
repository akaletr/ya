package model

type ShortenerRequest struct {
	URL string `json:"url"`
}

type ShortenerResponse struct {
	Result string `json:"result"`
}
