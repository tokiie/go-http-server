package main

import (
	"fmt"
	"net"
	"os"

	"httpfromtcp/internal/request"
)

func main() {
	ln, err := net.Listen("tcp", ":42069")

	if err != nil {
		fmt.Println("err while listenning the port: ", err)
		os.Exit(1)
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("err while accepting:", err)
		}
		fmt.Println("Connection has been accepted")

		request, err := request.RequestFromReader(conn)
		if err != nil {
			fmt.Println("Err while reading:", err)
		}

		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", request.RequestLine.Method)
		fmt.Printf("- Target: %s\n", request.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", request.RequestLine.HttpVersion)
		fmt.Println("Headers:")
		for key, value := range request.Headers {
			fmt.Printf("- %s: %s\n", key, value)
		}
		fmt.Println("Body:")
		fmt.Println(string(request.Body))

	}
}
