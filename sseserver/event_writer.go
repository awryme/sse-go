package sseserver

import (
	"bytes"
	"fmt"
	"time"
)

func dataLine(str string) string {
	return "data: " + str + "\n"
}

// EventWriter is an object used to write data to the event.
type EventWriter struct {
	buf *bytes.Buffer
}

// NewEventWriter create a new EventWriter.
// It uses passed buffer to store event lines and immediately writes name, id and retry lines if they are provided.
func NewEventWriter(buf *bytes.Buffer, name string, id string, retry time.Duration) *EventWriter {
	// write common fields immediately
	if id != "" {
		fmt.Fprintf(buf, "id: %s\n", id)
	}

	if retry != 0 {
		fmt.Fprintf(buf, "retry: %d\n", retry.Milliseconds())
	}

	if name != "" {
		fmt.Fprintf(buf, "event: %s\n", name)
	}

	return &EventWriter{
		buf: buf,
	}
}

// Format formats a data line using fmt package with provided format and args.
func (ev *EventWriter) Format(format string, args ...any) {
	fmt.Fprintf(ev.buf, dataLine(format), args...)
}

// Write writes a data line as is.
func (ev *EventWriter) Write(data string) {
	ev.buf.WriteString(dataLine(data))
}

// Result writes the final newline and returns resulting buffer.
// Result should be called only once.
// After Result is called, not more data should be written to the event.
func (ev *EventWriter) Result() []byte {
	ev.buf.WriteByte('\n')
	return ev.buf.Bytes()
}
