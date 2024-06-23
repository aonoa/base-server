package server

import (
	"github.com/go-kratos/kratos/v2/transport/http"
	http1 "net/http"
	"strings"
)

const (
	baseContentType = "application"
)

// ContentType returns the content-type with base prefix.
func ContentType(subtype string) string {
	return strings.Join([]string{baseContentType, subtype}, "/")
}

// ResponseEncoder encodes the object to the HTTP response.
func ResponseEncoder(w http.ResponseWriter, r *http.Request, v interface{}) error {
	if v == nil {
		return nil
	}
	if rd, ok := v.(http.Redirector); ok {
		url, code := rd.Redirect()
		http1.Redirect(w, r, url, code)
		return nil
	}
	codec, _ := http.CodecForRequest(r, "Accept")

	type Aa struct {
		Code    int          `json:"code"`
		Result  *interface{} `json:"result,omitempty"`
		Message string       `json:"Message"`
		Type    string       `json:"type"`
	}

	item := &Aa{
		Result:  &v,
		Message: "ok",
		Type:    "success",
	}
	data, err := codec.Marshal(item)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", ContentType(codec.Name()))
	_, err = w.Write(data)
	if err != nil {
		return err
	}
	return nil
}
