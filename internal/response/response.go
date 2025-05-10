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

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	if _, err := w.Write([]byte(StatusLine[statusCode])); err != nil {
		return err
	}

	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	return headers.Headers{
		"Content-Length": fmt.Sprintf("%d", contentLen),
		"Connection":     "closed",
		"Content-Type":   "text/plain",
	}
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for key, header := range headers {
		line := fmt.Sprintf("%s: %s\r\n", key, header)
		if _, err := w.Write([]byte(line)); err != nil {
			return err
		}
	}
	if _, err := w.Write([]byte("\r\n")); err != nil {
		return err
	}
	return nil
}
