package sseserver

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/awryme/sse-go"
)

// LogMiddleware is a standard http middleware to write SSE event lines.
// It can be used for debugging purposes.
func LogMiddleware(printLines func(lines []string)) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			pr, pw := io.Pipe()
			go func() {
				scanner := bufio.NewScanner(pr)
				scanner.Split(splitAt("\n\n"))
				for scanner.Scan() {
					event := scanner.Text()
					lines := strings.Split(event, "\n")
					printLines(lines)
				}
			}()
			w = &respWrapper{
				ResponseWriter: w,
				pipeWriter:     pw,
			}
			next.ServeHTTP(w, r)
		})
	}
}

type respWrapper struct {
	http.ResponseWriter
	pipeWriter io.Writer
}

func (rw *respWrapper) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

func (rw *respWrapper) Write(p []byte) (int, error) {
	if rw.ResponseWriter.Header().Get(headerKeyContentType) == sse.ContentType {
		_, err := rw.pipeWriter.Write(p)
		if err != nil {
			return 0, fmt.Errorf("write sse line to logs pipe: %w", err)
		}
	}
	return rw.ResponseWriter.Write(p)
}

// Custom split function. This will split string at 'sbustring' i.e # or // etc....
// Taken from https://gist.github.com/guleriagishere/8185da56df6d64c2ab652a59808c1011
func splitAt(substring string) func(data []byte, atEOF bool) (advance int, token []byte, err error) {
	searchBytes := []byte(substring)
	searchLen := len(substring)
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		// Return Nothing if at the end of file or no data passed.
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		// Find next separator and return token.
		if i := bytes.Index(data, searchBytes); i >= 0 {
			return i + searchLen, data[0:i], nil
		}

		// If we're at EOF, we have a final, non-terminated line. Return it.
		if atEOF {
			return len(data), data, nil
		}

		// Request more data.
		return 0, nil, nil
	}
}
