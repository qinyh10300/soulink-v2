package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// DecodeJSON 按统一规则解析请求体 JSON。
func DecodeJSON(r *http.Request, dst any) error {
	if r.Body == nil {
		return NewError(http.StatusBadRequest, CodeInvalidParams, "鍙傛暟閿欒")
	}

	defer r.Body.Close()

	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return decodeErr(err)
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return NewError(http.StatusBadRequest, CodeInvalidParams, "鍙傛暟閿欒")
	}

	return nil
}

// decodeErr 将 JSON 解析错误映射成统一业务错误。
func decodeErr(err error) error {
	switch {
	case errors.Is(err, io.EOF):
		return NewError(http.StatusBadRequest, CodeInvalidParams, "鍙傛暟閿欒")
	case errors.As(err, new(*json.SyntaxError)):
		return NewError(http.StatusBadRequest, CodeInvalidParams, "鍙傛暟閿欒")
	case errors.As(err, new(*json.UnmarshalTypeError)):
		return NewError(http.StatusBadRequest, CodeInvalidParams, "鍙傛暟閿欒")
	default:
		return Wrap(fmt.Errorf("decode json: %w", err), http.StatusBadRequest, CodeInvalidParams, "鍙傛暟閿欒")
	}
}
