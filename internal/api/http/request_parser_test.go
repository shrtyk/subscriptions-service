package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadJson(t *testing.T) {
	t.Parallel()

	type request struct {
		Foo string `json:"foo"`
	}

	tests := []struct {
		name     string
		body     string
		dst      any
		expErr   string
		checkErr func(t *testing.T, err error, expErr string)
	}{
		{
			name:     "success",
			body:     `{"foo":"bar"}`,
			dst:      &request{},
			checkErr: func(t *testing.T, err error, expErr string) { require.NoError(t, err) },
		},
		{
			name:     "invalid json",
			body:     `{"foo":"bar"`,
			dst:      &request{},
			checkErr: func(t *testing.T, err error, expErr string) { assert.Error(t, err) },
		},
		{
			name:     "empty body",
			body:     "",
			dst:      &request{},
			expErr:   "body must not be empty",
			checkErr: func(t *testing.T, err error, expErr string) { assert.EqualError(t, err, expErr) },
		},
		{
			name:     "unknown field",
			body:     `{"bar":"baz"}`,
			dst:      &request{},
			expErr:   "body contains unknown key",
			checkErr: func(t *testing.T, err error, expErr string) { assert.Contains(t, err.Error(), expErr) },
		},
		{
			name:     "multiple json values",
			body:     `{"foo":"bar"}{"bar":"baz"}`,
			dst:      &request{},
			expErr:   "body must contain a single JSON value",
			checkErr: func(t *testing.T, err error, expErr string) { assert.EqualError(t, err, expErr) },
		},
		{
			name:     "syntax error",
			body:     `{"foo":"bar"`,
			dst:      &request{},
			expErr:   "body contains badly-formed JSON",
			checkErr: func(t *testing.T, err error, expErr string) { assert.Contains(t, err.Error(), expErr) },
		},
		{
			name:     "unmarshal type error",
			body:     `{"foo":123}`,
			dst:      &request{},
			expErr:   "body contains incorrect JSON type for field",
			checkErr: func(t *testing.T, err error, expErr string) { assert.Contains(t, err.Error(), expErr) },
		},
		{
			name:     "max bytes error",
			body:     `{"foo":"` + strings.Repeat("a", maxRequestBodyBytes) + `"}`,
			dst:      &request{},
			expErr:   "body must not be larger than",
			checkErr: func(t *testing.T, err error, expErr string) { assert.Contains(t, err.Error(), expErr) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))

			err := ReadJSON(w, r, tt.dst)

			tt.checkErr(t, err, tt.expErr)
		})
	}
}
