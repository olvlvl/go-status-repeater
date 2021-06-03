package statusrepeater

import (
	"net/http"
	"sync"
	"time"
)

// DefaultDuration is a default duration.
const DefaultDuration = 15 * time.Minute

// DefaultFormatKey uses the request's method and URI to format a key.
func DefaultFormatKey(request *http.Request) (string, bool) {
	return request.Method + request.URL.RequestURI(), true
}

// Handler is a middleware that repeats a status code for a specified duration.
func Handler(
	next http.Handler,
	statusCode int,
	duration time.Duration,
	formatKey func(*http.Request) (string, bool),
) http.Handler {
	r := repeater{
		duration: duration,
	}

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		k, ok := formatKey(request)
		if !ok {
			next.ServeHTTP(writer, request)

			return
		}

		if r.has(k) {
			writer.WriteHeader(statusCode)

			return
		}

		ws := &writerSpy{next: writer}

		next.ServeHTTP(ws, request)

		if ws.statusCode == statusCode {
			r.add(k)
		}
	})
}

type repeater struct {
	// Duration of the repeating window.
	duration time.Duration
	// Where _key_ is a key for a request and _value_ is an expiration time.
	m sync.Map
}

func (r *repeater) has(key string) bool {
	t, ok := r.m.Load(key)
	if !ok {
		return false
	}

	return t.(int64) > time.Now().UnixNano()
}

func (r *repeater) add(key string) {
	r.m.Store(key, time.Now().Add(r.duration).UnixNano())
}

type writerSpy struct {
	next       http.ResponseWriter
	statusCode int
}

func (ws *writerSpy) Header() http.Header {
	return ws.next.Header()
}

func (ws *writerSpy) Write(b []byte) (int, error) {
	return ws.next.Write(b)
}

func (ws *writerSpy) WriteHeader(statusCode int) {
	ws.statusCode = statusCode

	ws.next.WriteHeader(statusCode)
}
