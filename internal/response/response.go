package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
)

type StatusCode int

const (
	STATUS_OK           = 200
	STATUS_BAD_REQUEST  = 400
	STATUS_SERVER_ERROR = 500
)

var StatusLine = map[StatusCode]string{
	STATUS_OK:           "HTTP/1.1 200 OK",
	STATUS_BAD_REQUEST:  "HTTP/1.1 400 Bad Request",
	STATUS_SERVER_ERROR: "HTTP/1.1 500 Internal Server Error",
}

type writerState int

const (
	writerStateStatusLine writerState = iota
	writerStateHeaders
	writerStateBody
	writerStateTrailers
)

type Writer struct {
	writer io.Writer
	state  writerState
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		state:  writerStateStatusLine,
		writer: w,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != writerStateStatusLine {
		return fmt.Errorf("status line has already been written")
	}
	if _, err := w.writer.Write([]byte(StatusLine[statusCode])); err != nil {
		return err
	}

	w.state = writerStateHeaders
	return nil
}
func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.state != writerStateHeaders {
		return fmt.Errorf("status line has not been written")
	}

	for key, header := range headers {
		line := fmt.Sprintf("%s: %s\r\n", key, header)
		if _, err := w.writer.Write([]byte(line)); err != nil {
			return err
		}
	}

	_, err := w.writer.Write([]byte("\r\n"))
	w.state = writerStateBody
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.state != writerStateBody {
		return 0, fmt.Errorf("headers have to be written first")
	}

	return w.writer.Write(p)
}
func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.state != writerStateBody {
		return 0, fmt.Errorf("cannot write body in state 'wrote headers'")
	}
	chunkSize := len(p)

	nTotal := 0
	n, err := fmt.Fprintf(w.writer, "%x\r\n", chunkSize)
	if err != nil {
		return nTotal, err
	}
	nTotal += n

	n, err = w.writer.Write(p)
	if err != nil {
		return nTotal, err
	}
	nTotal += n

	n, err = w.writer.Write([]byte("\r\n"))
	if err != nil {
		return nTotal, err
	}
	nTotal += n
	return nTotal, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.state != writerStateBody {
		return 0, fmt.Errorf("cannot write body in state 'wrote body'")
	}
	n, err := w.writer.Write([]byte("0\r\n"))
	if err != nil {
		return n, err
	}

	w.state = writerStateTrailers
	return n, nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")
	return h
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.state != writerStateTrailers {
		return fmt.Errorf("cannot write trailers in state %d", w.state)
	}
	defer func() { w.state = writerStateBody }()
	for k, v := range h {
		_, err := w.writer.Write([]byte(fmt.Sprintf("%s: %s\r\n", k, v)))
		if err != nil {
			return err
		}
	}
	_, err := w.writer.Write([]byte("\r\n"))
	return err
}
