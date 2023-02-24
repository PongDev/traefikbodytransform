package traefikbodytransform

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestServeHTTP(t *testing.T) {
	cfg := CreateConfig()
	cfg.TransformerQueryParameterName = "transform"
	cfg.JSONTransformFieldName = "data"
	cfg.TokenTransformQueryParameterFieldName = "token"

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	handler, err := New(ctx, next, cfg, "transformer-plugin")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost?transform=body|json|bearer&token=TOKENDATA", strings.NewReader("RAWSTRING"))
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(recorder, req)

	assertJSONBody(t, req, map[string]string{"data": "RAWSTRING"})
	assertHeader(t, req, "Authorization", "Bearer TOKENDATA")
	assertHeader(t, req, "Content-Type", "application/json")
}

func assertHeader(t *testing.T, req *http.Request, key, expected string) {
	t.Helper()

	if req.Header.Get(key) != expected {
		t.Errorf("invalid header value: %s", req.Header.Get(key))
	}
}

func assertJSONBody(t *testing.T, req *http.Request, expected map[string]string) {
	t.Helper()

	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatal(err)
	}

	jsonBody := make(map[string]string)
	err = json.Unmarshal(reqBody, &jsonBody)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(jsonBody, expected) {
		t.Errorf("invalid json body: %s", jsonBody)
	}
}
