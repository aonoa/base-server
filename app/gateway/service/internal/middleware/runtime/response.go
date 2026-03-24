package runtime

import (
	"bytes"
	"io"
	"net/http"
)

func EmptyResponse(statusCode int) *http.Response {
	return &http.Response{
		Status:     http.StatusText(statusCode),
		StatusCode: statusCode,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(nil)),
	}
}
