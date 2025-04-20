package sseserver

import (
	"fmt"
	"net/http"

	"github.com/awryme/sse-go"
)

const headerKeyContentType = "Content-Type"

// Server is an Server Side Events server implementation.
type Server struct {
	resp http.ResponseWriter
	req  *http.Request
	rc   *http.ResponseController
}

// New create an Server instance.
// It immediately writes necessary response headers and flushes the response.
// No compression mechanism is provided as of yet.
func New(w http.ResponseWriter, r *http.Request) (*Server, error) {
	rc := http.NewResponseController(w)

	w.Header().Set(headerKeyContentType, sse.ContentType)
	w.Header().Set("Cache-Control", "no-cache")
	if r.ProtoMajor == 1 {
		// keepalive is forbidden for http2/3
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Connection
		w.Header().Set("Connection", "keep-alive")
	}

	// todo: set compression/encoding?

	// initial flush to send headers and check if flusher if supported
	if err := rc.Flush(); err != nil {
		return nil, fmt.Errorf("initial header flush: %w", err)
	}

	return &Server{
		resp: w,
		req:  r,
		rc:   http.NewResponseController(w),
	}, nil
}

// WriteEvent writes a fully formed event and flushes the response.
func (server *Server) WriteEvent(evt *EventWriter) error {
	if evt == nil {
		return nil
	}
	bytes := evt.Result()
	_, err := server.resp.Write(bytes)
	if err != nil {
		return fmt.Errorf("write event to wire: %w", err)
	}
	server.rc.Flush()

	return nil
}
