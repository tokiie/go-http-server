package server

import (
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
)

type Handler func(w *response.Writer, req *request.Request)
