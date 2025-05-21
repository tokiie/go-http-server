package main

import (
	"crypto/sha256"
	"fmt"
	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w *response.Writer, r *request.Request) {
	if r.RequestLine.RequestTarget == "/yourproblem" {
		handler400(w, r)
		return
	}
	if r.RequestLine.RequestTarget == "/myproblem" {
		handler500(w, r)
		return
	}

	if strings.HasPrefix(r.RequestLine.RequestTarget, "/httpbin") {
		r.Headers.Set("Transfer-Encoding", "chunked")
		delete(r.Headers, "content-length")
		proxyHandler(w, r)
		return
	}
	if r.RequestLine.RequestTarget == "/video" {
		videoStreaming(w, r)
		return
	}
	handler200(w, r)
	return
}

func handler200(w *response.Writer, _ *request.Request) {
	content := []byte(`<html>
		<head>
			<title>200 OK</title>
		</head>
		<body>
			<h1>Success!</h1>
			<p>Your request was an absolute banger.</p>
		</body>
		</html>`)
	headers := response.GetDefaultHeaders(len(content))
	headers.Override("Content-Type", "text/html")
	fmt.Println(headers["Content-Type"])
	w.WriteStatusLine(response.STATUS_OK)
	w.WriteHeaders(headers)
	w.WriteBody(content)

	return
}

func handler400(w *response.Writer, _ *request.Request) {
	content := []byte(`<html>
		<head>
			<title>400 Bad Request</title>
		</head>
		<body>
			<h1>Bad Request</h1>
			<p>Your request honestly kinda sucked.</p>
		</body>
		</html>`)
	headers := response.GetDefaultHeaders(len(content))
	headers.Override("Content-Type", "text/html")
	w.WriteStatusLine(response.STATUS_BAD_REQUEST)
	w.WriteHeaders(headers)
	w.WriteBody(content)

	return
}

func handler500(w *response.Writer, _ *request.Request) {
	content := []byte(`<html>
		<head>
			<title>500 Internal Server Error</title>
		</head>
		<body>
			<h1>Internal Server Error</h1>
			<p>Okay, you know what? This one is on me.</p>
		</body>
		</html>`)
	headers := response.GetDefaultHeaders(len(content))
	headers.Override("Content-Type", "text/html")
	w.WriteStatusLine(response.STATUS_SERVER_ERROR)
	w.WriteHeaders(headers)
	w.WriteBody(content)

	return
}

func proxyHandler(w *response.Writer, req *request.Request) {
	target := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
	url := "https://httpbin.org/" + target
	fmt.Println("Proxying to", url)
	resp, err := http.Get(url)
	if err != nil {
		handler500(w, req)
		return
	}
	defer resp.Body.Close()

	w.WriteStatusLine(response.STATUS_OK)
	h := response.GetDefaultHeaders(0)
	h.Override("Transfer-Encoding", "chunked")
	h.Override("Trailer", "X-Content-SHA256, X-Content-Length")
	h.Remove("Content-Length")
	w.WriteHeaders(h)

	fullBody := make([]byte, 0)

	const maxChunkSize = 1024
	buffer := make([]byte, maxChunkSize)
	for {
		n, err := resp.Body.Read(buffer)
		fmt.Println("Read", n, "bytes")
		if n > 0 {
			_, err = w.WriteChunkedBody(buffer[:n])
			if err != nil {
				fmt.Println("Error writing chunked body:", err)
				break
			}
			fullBody = append(fullBody, buffer[:n]...)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Error reading response body:", err)
			break
		}
	}
	_, err = w.WriteChunkedBodyDone()
	if err != nil {
		fmt.Println("Error writing chunked body done:", err)
	}
	trailers := headers.NewHeaders()
	sha256 := fmt.Sprintf("%x", sha256.Sum256(fullBody))
	trailers.Override("X-Content-SHA256", sha256)
	trailers.Override("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))
	err = w.WriteTrailers(trailers)
	if err != nil {
		fmt.Println("Error writing tr	ailers:", err)
	}
	fmt.Println("Wrote trailers")
}

func videoStreaming(w *response.Writer, r *request.Request) {

	file, err := os.ReadFile("assets/vim.mp4")
	h := response.GetDefaultHeaders(len(file))
	h.Override("Content-Type", "video/mp4")

	if err != nil {
		fmt.Println("error while reading file: ", err)
		return
	}

	w.WriteStatusLine(response.STATUS_OK)
	w.WriteHeaders(h)
	w.WriteBody(file)
}
