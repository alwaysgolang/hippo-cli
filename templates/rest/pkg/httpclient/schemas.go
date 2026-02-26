package httpclient

import "net/http"

type JsonMap map[string]any

type Header struct {
	Key   string
	Value string
}
type Transport struct {
	BaseURL    string
	HTTPClient *http.Client
	Headers    []Header
}
