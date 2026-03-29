package client

type PutRecordRequest struct {
	OriginalURL string `json:"original_url"`
	ShortCode   string `json:"short_code"`
	ExpTime     int    `json:"exp_time"`
}

type URLResponse struct {
	OriginalURL string `json:"original_url"`
	ShortCode   string `json:"short_code"`
	ExpTime     int    `json:"exp_time"`
}
