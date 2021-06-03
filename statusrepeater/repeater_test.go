package statusrepeater_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/olvlvl/go-status-repeater/statusrepeater"
)

type statusCodeRecorder struct {
	statusCodes []int
}

func (rw *statusCodeRecorder) Header() http.Header {
	panic("should not be called")
}

func (rw *statusCodeRecorder) Write([]byte) (int, error) {
	panic("should not be called")
}

func (rw *statusCodeRecorder) WriteHeader(statusCode int) {
	rw.statusCodes = append(rw.statusCodes, statusCode)
}

func TestHandler(t *testing.T) {
	r1, _ := http.NewRequest(http.MethodGet, "/articles?id=123", nil)
	r2, _ := http.NewRequest(http.MethodGet, "/articles?id=456", nil)

	cases := map[string]struct {
		formatKey          func(request *http.Request) (string, bool)
		handlerStatusCodes []int
		requests           []*http.Request
		want               []int
	}{
		"unable to format key: the repeater shouldn't activate": {
			formatKey: func(request *http.Request) (string, bool) {
				return "", false
			},
			handlerStatusCodes: []int{http.StatusNotFound, http.StatusNotFound, http.StatusNotFound},
			requests:           []*http.Request{r1, r1, r1},
			want:               []int{http.StatusNotFound, http.StatusNotFound, http.StatusNotFound},
		},

		"ok then not ok": {
			formatKey:          statusrepeater.DefaultFormatKey,
			handlerStatusCodes: []int{http.StatusOK, http.StatusCreated, http.StatusAccepted, http.StatusNotFound},
			requests:           []*http.Request{r1, r1, r1, r1, r1},
			want:               []int{http.StatusOK, http.StatusCreated, http.StatusAccepted, http.StatusNotFound, http.StatusNotFound},
		},

		"multiple requests: the repeater should activate for the bad one only": {
			formatKey:          statusrepeater.DefaultFormatKey,
			handlerStatusCodes: []int{http.StatusNotFound, http.StatusOK},
			requests:           []*http.Request{r1, r1, r2, r1, r1},
			want:               []int{http.StatusNotFound, http.StatusNotFound, http.StatusOK, http.StatusNotFound, http.StatusNotFound},
		},
	}

	for name, tc := range cases {
		tc := tc

		t.Run(name, func(t *testing.T) {
			i := 0
			h := statusrepeater.Handler(
				http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					writer.WriteHeader(tc.handlerStatusCodes[i])
					i++
				}),
				http.StatusNotFound,
				statusrepeater.DefaultDuration,
				tc.formatKey,
			)

			writer := &statusCodeRecorder{}

			for _, request := range tc.requests {
				h.ServeHTTP(writer, request)
			}

			assert.Equal(t, tc.want, writer.statusCodes)
			assert.Equal(t, len(tc.handlerStatusCodes), i, "not all status codes have been pushed")
		})
	}
}

func TestRequestWriter(t *testing.T) {
	recorder := httptest.NewRecorder()
	body := []byte("Hello!")
	statusCode := http.StatusCreated

	h := statusrepeater.Handler(
		http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("X-Testing", "Testing")
			writer.WriteHeader(statusCode)
			_, _ = writer.Write(body)
		}),
		http.StatusNotFound,
		statusrepeater.DefaultDuration,
		statusrepeater.DefaultFormatKey,
	)

	request, _ := http.NewRequest(http.MethodGet, "", nil)

	h.ServeHTTP(recorder, request)

	assert.Equal(t, statusCode, recorder.Code)
	assert.Equal(t, "Testing", recorder.Header().Get("X-Testing"))
	assert.Equal(t, body, recorder.Body.Bytes())
}
