package http

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/opcotech/elemo/internal/pkg/convert"
)

func TestWriteJSONResponse(t *testing.T) {
	type fields struct {
		response string
		status   int
	}
	type args struct {
		writer   http.ResponseWriter
		response any
		status   int
	}
	tests := []struct {
		name   string
		args   args
		fields fields
	}{
		{
			name: "write JSON response",
			args: args{
				writer: httptest.NewRecorder(),
				response: map[string]string{
					"message": "hello",
				},
				status: http.StatusOK,
			},
			fields: fields{
				response: "{\"message\":\"hello\"}",
				status:   http.StatusOK,
			},
		},
		{
			name: "write JSON response with error",
			args: args{
				writer:   httptest.NewRecorder(),
				response: func() {},
				status:   http.StatusOK,
			},
			fields: fields{
				response: errors.Join(convert.ErrMarshal, fmt.Errorf("json: unsupported type: func()")).Error(),
				status:   http.StatusInternalServerError,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			WriteJSONResponse(tt.args.writer, tt.args.response, tt.args.status)
			rr := tt.args.writer.(*httptest.ResponseRecorder)

			assert.Equal(t, tt.fields.status, rr.Code, "status code should be equal")
			assert.Equal(t, "application/json", rr.Header().Get("Content-Type"), "content type should be equal")
			assert.Equal(t, tt.fields.response, rr.Body.String(), "body should be equal")
		})
	}
}
