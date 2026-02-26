package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	customErrors "gotemplate/pkg/errors"
	"gotemplate/pkg/logs"
)

func NewTransport(baseURL string, headers []Header) (*Transport, func()) {
	logs.Info("Creating new HTTP transport", "baseURL", baseURL, "headersCount", len(headers))

	transport := &Transport{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 60 * time.Second},
		Headers:    headers,
	}

	cleanup := func() {
		logs.Info("Cleaning up HTTP transport", "baseURL", baseURL, "headersCount", len(headers))
		transport.HTTPClient.CloseIdleConnections()
	}

	return transport, cleanup
}

func (t *Transport) doRequest(ctx context.Context, uri, method string, body []byte, queryString *JsonMap, extraHeaders *[]Header) ([]byte, int, error) {
	if len(uri) > 0 && uri[0] == '/' {
		uri = uri[1:]
	}

	fullURL := t.BaseURL + uri

	logs.InfoCtx(ctx, "Preparing HTTP request",
		"method", method,
		"url", fullURL,
		"bodySize", len(body),
		"hasQueryString", queryString != nil && len(*queryString) > 0,
	)

	logs.DebugCtx(ctx, "Request details",
		"method", method,
		"url", fullURL,
		"body", string(body),
		"queryString", queryString,
		"extraHeaders", extraHeaders,
	)

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bytes.NewBuffer(body))
	if err != nil {
		logs.PanicCtx(ctx, "Failed to create HTTP request", "error", err, "url", fullURL, "method", method)
		return nil, 0, customErrors.WrapSystemError(err)
	}

	for _, header := range *extraHeaders {
		req.Header.Set(header.Key, header.Value)
		logs.DebugCtx(ctx, "Setting extra header", "key", header.Key, "value", header.Value)
	}
	for _, header := range t.Headers {
		req.Header.Set(header.Key, header.Value)
		logs.DebugCtx(ctx, "Setting transport header", "key", header.Key, "value", header.Value)
	}

	if queryString != nil && len(*queryString) > 0 {
		query := req.URL.Query()
		for key, value := range *queryString {
			query.Add(key, fmt.Sprintf("%v", value))
			logs.DebugCtx(ctx, "Adding query parameter", "key", key, "value", value)
		}
		req.URL.RawQuery = query.Encode()
		logs.DebugCtx(ctx, "Final query string", "rawQuery", req.URL.RawQuery)
	}

	startTime := time.Now()
	logs.InfoCtx(ctx, "Sending HTTP request", "method", method, "url", fullURL)

	resp, err := t.HTTPClient.Do(req)
	duration := time.Since(startTime)

	if err != nil {
		logs.ErrorCtx(ctx, "HTTP request failed",
			"error", err,
			"url", fullURL,
			"method", method,
			"duration", duration,
		)
		return nil, 0, customErrors.WrapSystemError(err)
	}
	defer func() { _ = resp.Body.Close() }()

	logs.InfoCtx(ctx, "Received HTTP response",
		"statusCode", resp.StatusCode,
		"status", resp.Status,
		"duration", duration,
		"url", fullURL,
	)

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		logs.ErrorCtx(ctx, "Failed to read response body",
			"error", err,
			"statusCode", resp.StatusCode,
			"url", fullURL,
		)
		return nil, resp.StatusCode, customErrors.WrapSystemError(err)
	}

	logs.DebugCtx(ctx, "Response body received",
		"statusCode", resp.StatusCode,
		"contentSize", len(content),
		"content", string(content),
	)
	switch {
	case resp.StatusCode >= http.StatusInternalServerError:
		logs.ErrorCtx(ctx, "HTTP request returned server error status",
			"statusCode", resp.StatusCode,
			"status", resp.Status,
			"url", fullURL,
			"responseBody", string(content),
		)
		err = customErrors.WrapSystemError(errors.New(string(content)))
	case resp.StatusCode >= http.StatusBadRequest:
		logs.WarnCtx(ctx, "HTTP request returned client error status",
			"statusCode", resp.StatusCode,
			"status", resp.Status,
			"url", fullURL,
			"responseBody", string(content),
		)
		err = customErrors.WrapExternalServiceError(errors.New(string(content)))
	default:
		logs.InfoCtx(ctx, "HTTP request completed successfully",
			"statusCode", resp.StatusCode,
			"duration", duration,
			"responseSize", len(content),
			"url", fullURL,
		)
	}

	return content, resp.StatusCode, err
}

func (t *Transport) Get(ctx context.Context, url string, queryString *JsonMap) ([]byte, int, error) {
	logs.InfoCtx(ctx, "Executing GET request", "url", url, "hasQueryString", queryString != nil && len(*queryString) > 0)
	return t.doRequest(ctx, url, http.MethodGet, nil, queryString, &[]Header{})
}

func (t *Transport) Post(ctx context.Context, url string, body any) ([]byte, int, error) {
	logs.InfoCtx(ctx, "Executing POST request", "url", url, "bodyType", fmt.Sprintf("%T", body))

	var requestBody []byte
	var err error
	headers := make([]Header, 0)

	switch v := body.(type) {
	case *JsonMap:
		requestBody, err = json.Marshal(*v)
		headers = append(headers, Header{"Content-Type", "application/json"})
	case JsonMap:
		requestBody, err = json.Marshal(v)
		headers = append(headers, Header{"Content-Type", "application/json"})
	case map[string]any:
		requestBody, err = json.Marshal(v)
		headers = append(headers, Header{"Content-Type", "application/json"})
	case []byte:
		requestBody = v
	default:
		err = fmt.Errorf("unsupported body type %T", body)
	}

	if err != nil {
		logs.PanicCtx(ctx, "Failed to prepare POST request body", "error", err, "url", url)
		return nil, 0, customErrors.WrapSystemError(err)
	}
	return t.doRequest(ctx, url, http.MethodPost, requestBody, nil, &headers)
}
