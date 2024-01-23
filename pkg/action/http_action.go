package action

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

type HttpAction struct {
	url     string
	method  string
	headers map[string]string
	body    io.Reader
}

func NewHttpAction(url string, method string, headers map[string]string, body io.Reader) *HttpAction {
	return &HttpAction{url: url, method: method, headers: headers, body: body}
}

func (h *HttpAction) Do(ctx context.Context) error {
	request, err := http.NewRequestWithContext(ctx, h.method, h.url, h.body)
	for k, v := range h.headers {
		request.Header.Add(k, v)
	}
	if err != nil {
		return fmt.Errorf("HttpAction::Do error while creating request %s: %w", h.url, err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("HttpAction::Do error while calling url %s: %w", h.url, err)
	}
	defer func() {
		_ = response.Body.Close()
	}()
	if err != nil {
		return fmt.Errorf("HttpAction::Do error while calling url %s: %w", h.url, err)
	}
	if response.StatusCode/100 != 2 {
		r, err := io.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("HttpAction::Do error while calling url %s: status code is %d", h.url, response.StatusCode)
		}
		return fmt.Errorf("HttpAction::Do error while calling url %s: status code is %d, response body is: %s", h.url, response.StatusCode, string(r))
	}
	return nil
}
