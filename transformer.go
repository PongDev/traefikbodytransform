// Package traefikbodytransform plugin.
package traefikbodytransform

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
)

// Config the plugin configuration.
type Config struct {
	TransformerQueryParameterName         string `json:"transformerQueryParameterName,omitempty"`
	JSONTransformFieldName                string `json:"jsonTransformFieldName,omitempty"`
	TokenTransformQueryParameterFieldName string `json:"tokenTransformQueryParameterFieldName,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		TransformerQueryParameterName:         "transformer",
		JSONTransformFieldName:                "data",
		TokenTransformQueryParameterFieldName: "token",
	}
}

// transformer plugin.
type transformer struct {
	next                                  http.Handler
	name                                  string
	transformerQueryParameterName         string
	jsonTransformFieldName                string
	tokenTransformQueryParameterFieldName string
}

// New created a new transformer plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	return &transformer{
		next:                                  next,
		name:                                  name,
		transformerQueryParameterName:         config.TransformerQueryParameterName,
		jsonTransformFieldName:                config.JSONTransformFieldName,
		tokenTransformQueryParameterFieldName: config.TokenTransformQueryParameterFieldName,
	}, nil
}

func (a *transformer) log(format string) {
	_, writeLogError := os.Stderr.WriteString(a.name + ": " + format)
	if writeLogError != nil {
		panic(writeLogError.Error())
	}
}

func (a *transformer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	transformerOption := make(map[string]bool)

	if param := req.URL.Query().Get(a.transformerQueryParameterName); len(param) > 0 {
		for _, opt := range strings.Split(strings.ToLower(param), "|") {
			transformerOption[opt] = true
		}
	}

	if transformerOption["body"] {
		reqBody, err := io.ReadAll(req.Body)
		if err != nil {
			a.log(err.Error())

			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		jsonBody, err := json.Marshal(map[string]string{
			a.jsonTransformFieldName: string(reqBody),
		})
		if err != nil {
			a.log(err.Error())

			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		req.Body = io.NopCloser(strings.NewReader(string(jsonBody)))
		req.ContentLength = int64(len(jsonBody))
	}
	if transformerOption["json"] {
		req.Header.Set("Content-Type", "application/json")
	}
	if transformerOption["bearer"] {
		token := req.URL.Query().Get(a.tokenTransformQueryParameterFieldName)
		req.Header.Set("Authorization", "Bearer "+token)
	}

	a.next.ServeHTTP(rw, req)
}
